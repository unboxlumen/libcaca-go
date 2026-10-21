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

package dither

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
)

func TestDrawProducesCells(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{0, 0, 0, 255})
	img.SetNRGBA(1, 0, color.NRGBA{255, 255, 255, 255})
	img.SetNRGBA(0, 1, color.NRGBA{255, 255, 255, 255})
	img.SetNRGBA(1, 1, color.NRGBA{0, 0, 0, 255})

	d, err := New(img, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	c := canvas.MustNew(8, 8)
	if err := d.Draw(c, 0, 0, 8, 8); err != nil {
		t.Fatal(err)
	}
	nonblank := 0
	for _, cell := range c.Cells() {
		if cell.Ch != ' ' {
			nonblank++
		}
	}
	if nonblank == 0 {
		t.Fatal("dither produced no visible cells")
	}
}

func TestDrawTransparentLeavesBlank(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 0})

	d, err := New(img, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	c := canvas.MustNew(2, 2)
	if err := d.Draw(c, 0, 0, 2, 2); err != nil {
		t.Fatal(err)
	}
	for _, cell := range c.Cells() {
		if cell.Ch != ' ' {
			t.Fatalf("transparent pixel drew %q", cell.Ch)
		}
	}
}

func TestOptions(t *testing.T) {
	for _, s := range []string{"none", "prefilter", "default"} {
		if _, err := ParseAntialias(s); err != nil {
			t.Fatalf("antialias %q: %v", s, err)
		}
	}
	for _, s := range []string{"mono", "gray", "8", "16", "fullgray", "full8", "full16", "default"} {
		if _, err := ParseColor(s); err != nil {
			t.Fatalf("color %q: %v", s, err)
		}
	}
	for _, s := range []string{"ascii", "shades", "blocks", "default"} {
		if _, err := ParseCharset(s); err != nil {
			t.Fatalf("charset %q: %v", s, err)
		}
	}
	for _, s := range []string{"none", "ordered2", "ordered4", "ordered8", "random", "fstein", "default"} {
		if _, err := ParseAlgorithm(s); err != nil {
			t.Fatalf("algorithm %q: %v", s, err)
		}
	}
}

func TestAllAlgorithmsRun(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 16), uint8(y * 16), 128, 255})
		}
	}
	algs := []Algorithm{NoDither, Ordered2, Ordered4, Ordered8, Random, FloydSteinberg}
	for _, a := range algs {
		opts := DefaultOptions()
		opts.Algorithm = a
		d, err := New(img, opts)
		if err != nil {
			t.Fatal(err)
		}
		c := canvas.MustNew(16, 16)
		if err := d.Draw(c, 0, 0, 16, 16); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFullGrayOnlyUsesGrayPalette(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}
	opts := DefaultOptions()
	opts.Color = FullGray
	d, err := New(img, opts)
	if err != nil {
		t.Fatal(err)
	}
	c := canvas.MustNew(4, 4)
	if err := d.Draw(c, 0, 0, 4, 4); err != nil {
		t.Fatal(err)
	}
	for _, cell := range c.Cells() {
		fg, bg := cell.A.FG(), cell.A.BG()
		for _, cc := range []canvas.Color{fg, bg} {
			if cc != canvas.Black && cc != canvas.DarkGray && cc != canvas.LightGray && cc != canvas.White {
				t.Fatalf("non-gray colour %d in fullgray mode", cc)
			}
		}
	}
}

func TestRandomSeedReproducible(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 128, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 128; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 2), uint8(y * 4), 90, 255})
		}
	}
	draw := func(seed uint64) []byte {
		o := DefaultOptions()
		o.Algorithm = Random
		o.Seed = seed
		d, err := New(img, o)
		if err != nil {
			t.Fatal(err)
		}
		c := canvas.MustNew(64, 32)
		if err := d.Draw(c, 0, 0, 64, 32); err != nil {
			t.Fatal(err)
		}
		out, err := codec.Export(c, "caca")
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	a, b := draw(42), draw(42)
	if !bytes.Equal(a, b) {
		t.Fatal("same seed produced different output")
	}
	c := draw(43)
	if bytes.Equal(a, c) {
		t.Fatal("different seeds produced identical output")
	}
}
