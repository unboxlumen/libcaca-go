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

package preview

import (
	"image"
	"image/color"
	"testing"
)

func TestRender(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if (x+y)&1 == 0 {
				img.SetNRGBA(x, y, color.NRGBA{255, 255, 255, 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{0, 0, 0, 255})
			}
		}
	}
	opts := DefaultOptions()
	opts.Cols, opts.Rows = 8, 4
	out, err := Render(img, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.Rect.Dx() <= 0 || out.Rect.Dy() <= 0 {
		t.Fatalf("bad image: %v", out)
	}
	nonblank := 0
	for i := 3; i < len(out.Pix); i += 4 {
		if out.Pix[i] != 0 {
			nonblank++
		}
	}
	if nonblank == 0 {
		t.Fatal("render produced no opaque pixels")
	}
}
