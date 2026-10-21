// Copyright (c) 2026 The nfilesart Authors.
//
// nfilesart is a Go port of the core of libcaca, the WTFPL-licensed
// caca/ library sources by Sam Hocevar (libcaca 0.99.beta20,
// https://github.com/cacalabs/libcaca, commit de48e1a0888caabdc6276afdea46e9096d0adb05).
// Only the WTFPL parts of libcaca were used; see PROVENANCE.md.
//
// This program is free software. It comes without any warranty, to the
// extent permitted by applicable law. You can redistribute it and/or
// modify it under the terms of the Do What The Fuck You Want To Public
// License, Version 2, as published by Sam Hocevar. See LICENSE.

package dither_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/dither"
)

// goldenImage matches /tmp/caca-oracle/dither-oracle.c make_pixels().
func goldenImage() image.Image {
	const w, h = 16, 16
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 16),
				G: uint8(y * 16),
				B: uint8((x*37 + y*53) & 0xff),
				A: 255,
			})
		}
	}
	return img
}

func TestGoldenDither(t *testing.T) {
	cases := []struct {
		name string
		opts func() dither.Options
	}{
		{"default", dither.DefaultOptions},
		{"none", func() dither.Options {
			o := dither.DefaultOptions()
			o.Algorithm = dither.NoDither
			return o
		}},
		{"ordered2", func() dither.Options {
			o := dither.DefaultOptions()
			o.Algorithm = dither.Ordered2
			return o
		}},
		{"ordered4", func() dither.Options {
			o := dither.DefaultOptions()
			o.Algorithm = dither.Ordered4
			return o
		}},
		{"ordered8", func() dither.Options {
			o := dither.DefaultOptions()
			o.Algorithm = dither.Ordered8
			return o
		}},
		{"shades", func() dither.Options {
			o := dither.DefaultOptions()
			o.Charset = dither.Shades
			return o
		}},
		{"blocks", func() dither.Options {
			o := dither.DefaultOptions()
			o.Charset = dither.Blocks
			return o
		}},
		{"fullgray", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.FullGray
			return o
		}},
		{"mono", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.Mono
			return o
		}},
		{"gray", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.Gray
			return o
		}},
		{"16color", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.ANSI16
			return o
		}},
		{"full8", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.Full8
			return o
		}},
		{"noaa", func() dither.Options {
			o := dither.DefaultOptions()
			o.Antialias = dither.AntialiasNone
			return o
		}},
		{"gamma22", func() dither.Options {
			o := dither.DefaultOptions()
			o.Gamma = 2.2
			return o
		}},
		{"gamma05", func() dither.Options {
			o := dither.DefaultOptions()
			o.Gamma = 0.5
			return o
		}},
	}

	img := goldenImage()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := dither.New(img, tc.opts())
			if err != nil {
				t.Fatal(err)
			}
			c := canvas.MustNew(16, 8)
			if err := d.Draw(c, 0, 0, 16, 8); err != nil {
				t.Fatal(err)
			}
			got, err := codec.Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/dither-" + tc.name + ".caca")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("dither-%s differs\n%s", tc.name, diffBytes(got, want))
			}
		})
	}
}

func diffBytes(got, want []byte) string {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	off := -1
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			off = i
			break
		}
	}
	if off < 0 {
		off = n
	}
	lo := off - 40
	if lo < 0 {
		lo = 0
	}
	hiG, hiW := off+80, off+80
	if hiG > len(got) {
		hiG = len(got)
	}
	if hiW > len(want) {
		hiW = len(want)
	}
	return fmt.Sprintf("length got=%d want=%d, first diff at %d\ngot : %q\nwant: %q",
		len(got), len(want), off, got[lo:hiG], want[lo:hiW])
}

func TestGoldenPackedPixels(t *testing.T) {
	opts := dither.DefaultOptions()

	// 8bpp grayscale palette.
	pix8 := make([]byte, 16*16)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			pix8[y*16+x] = byte((x*13 + y*7) & 0xff)
		}
	}
	d8, err := dither.NewPixels(pix8, dither.PixelFormat{BPP: 8, Width: 16, Height: 16, Pitch: 16}, opts)
	if err != nil {
		t.Fatal(err)
	}
	checkPackedGolden(t, d8, "dither-pix8.caca")

	// 8bpp custom palette; alpha is opaque so every pixel is drawn.
	var pr, pg, pb, pa [256]uint16
	for i := 0; i < 256; i++ {
		pr[i] = uint16((i * 7) & 0xfff)
		pg[i] = uint16((i * 13) & 0xfff)
		pb[i] = uint16((i * 29) & 0xfff)
		pa[i] = 0xfff
	}
	d8p, err := dither.NewPixels(pix8, dither.PixelFormat{BPP: 8, Width: 16, Height: 16, Pitch: 16}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := d8p.SetPalette(pr[:], pg[:], pb[:], pa[:]); err != nil {
		t.Fatal(err)
	}
	checkPackedGolden(t, d8p, "dither-pix8pal.caca")

	// 16bpp RGB565.
	pix16 := make([]byte, 16*16*2)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r := uint32(x * 31 / 15)
			g := uint32(y * 63 / 15)
			b := uint32((x*3 + y*5) & 31)
			binary.LittleEndian.PutUint16(pix16[(y*16+x)*2:], uint16(r<<11|g<<5|b))
		}
	}
	d16, err := dither.NewPixels(pix16, dither.PixelFormat{
		BPP: 16, Width: 16, Height: 16, Pitch: 32,
		RMask: 0xf800, GMask: 0x07e0, BMask: 0x001f,
	}, opts)
	if err != nil {
		t.Fatal(err)
	}
	checkPackedGolden(t, d16, "dither-pix16.caca")

	// 32bpp ARGB.
	pix32 := make([]byte, 16*16*4)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r := uint32(x * 16)
			g := uint32(y * 16)
			b := uint32((x*37 + y*53) & 0xff)
			binary.LittleEndian.PutUint32(pix32[(y*16+x)*4:], 0xff<<24|r<<16|g<<8|b)
		}
	}
	d32, err := dither.NewPixels(pix32, dither.PixelFormat{
		BPP: 32, Width: 16, Height: 16, Pitch: 64,
		RMask: 0x00ff0000, GMask: 0x0000ff00, BMask: 0x000000ff, AMask: 0xff000000,
	}, opts)
	if err != nil {
		t.Fatal(err)
	}
	checkPackedGolden(t, d32, "dither-pix32.caca")

	// 24bpp BGR must produce exactly the image-based default golden.
	bgr := make([]byte, 16*16*3)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r := byte(x * 16)
			g := byte(y * 16)
			b := byte((x*37 + y*53) & 0xff)
			o := (y*16 + x) * 3
			bgr[o+0], bgr[o+1], bgr[o+2] = b, g, r
		}
	}
	d24, err := dither.NewPixels(bgr, dither.PixelFormat{
		BPP: 24, Width: 16, Height: 16, Pitch: 48,
		RMask: 0xff0000, GMask: 0x00ff00, BMask: 0x0000ff,
	}, opts)
	if err != nil {
		t.Fatal(err)
	}
	checkPackedGolden(t, d24, "dither-default.caca")
}

func checkPackedGolden(t *testing.T, d *dither.Dither, name string) {
	t.Helper()
	c := canvas.MustNew(16, 8)
	if err := d.Draw(c, 0, 0, 16, 8); err != nil {
		t.Fatal(err)
	}
	got, err := codec.Export(c, "caca")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/golden/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s differs\n%s", name, diffBytes(got, want))
	}
}

func TestPackedPixelValidation(t *testing.T) {
	if _, err := dither.NewPixels(nil, dither.PixelFormat{BPP: 12}, dither.DefaultOptions()); err == nil {
		t.Fatal("expected bad bpp to fail")
	}
	if _, err := dither.NewPixels(make([]byte, 4), dither.PixelFormat{BPP: 8, Width: 4, Height: 2, Pitch: 4}, dither.DefaultOptions()); err == nil {
		t.Fatal("expected short buffer to fail")
	}
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	d, err := dither.New(img, dither.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.SetPalette(make([]uint16, 256), make([]uint16, 256), make([]uint16, 256), make([]uint16, 256)); err == nil {
		t.Fatal("palette on an image dither should fail")
	}
}

func extremeImage() image.Image {
	const w, h = 32, 32
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(0)
			if (x^y)&1 != 0 {
				v = 255
			}
			r, g, b := v, v, v
			if x%5 == 0 {
				r, g, b = 255, 0, 0
			}
			if y%5 == 0 {
				r, g, b = 0, 0, 255
			}
			if (x+y)%7 == 0 {
				r, g, b = 0, 255, 0
			}
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func TestGoldenExtremeDither(t *testing.T) {
	cases := []struct {
		name string
		opts func() dither.Options
	}{
		{"extreme-fstein", dither.DefaultOptions},
		{"extreme-noaa", func() dither.Options {
			o := dither.DefaultOptions()
			o.Antialias = dither.AntialiasNone
			return o
		}},
		{"extreme-ordered4", func() dither.Options {
			o := dither.DefaultOptions()
			o.Algorithm = dither.Ordered4
			return o
		}},
		{"extreme-shades", func() dither.Options {
			o := dither.DefaultOptions()
			o.Charset = dither.Shades
			return o
		}},
		{"extreme-fullgray", func() dither.Options {
			o := dither.DefaultOptions()
			o.Color = dither.FullGray
			return o
		}},
	}
	img := extremeImage()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := dither.New(img, tc.opts())
			if err != nil {
				t.Fatal(err)
			}
			c := canvas.MustNew(32, 16)
			if err := d.Draw(c, 0, 0, 32, 16); err != nil {
				t.Fatal(err)
			}
			got, err := codec.Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.name + ".caca")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.name, diffBytes(got, want))
			}
		})
	}
}
