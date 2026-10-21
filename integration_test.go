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

package art_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/dither"
)

func TestDitherToANSI(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 4),
				G: uint8(y * 8),
				B: uint8(255 - x*3),
				A: 255,
			})
		}
	}

	d, err := dither.New(img, dither.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	c, err := canvas.New(40, 20)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Draw(c, 0, 0, 40, 20); err != nil {
		t.Fatal(err)
	}

	out, err := codec.Export(c, "ansi")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("empty ANSI output")
	}
	if !bytes.Contains(out, []byte("\x1b[")) {
		t.Fatal("ANSI output has no escape sequence")
	}

	// Round-trip through the native format.
	native, err := codec.Export(c, "caca")
	if err != nil {
		t.Fatal(err)
	}
	c2 := canvas.MustNew(0, 0)
	if _, err := codec.Import(c2, native, "caca"); err != nil {
		t.Fatal(err)
	}
	if c2.Width() != 40 || c2.Height() != 20 {
		t.Fatalf("round-trip size = %dx%d", c2.Width(), c2.Height())
	}
}

func TestTGAExport(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	data, err := codec.Export(c, "tga")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 18 {
		t.Fatalf("TGA too short: %d", len(data))
	}
	if data[2] != 2 {
		t.Fatalf("TGA image type = %d, want 2", data[2])
	}
}
