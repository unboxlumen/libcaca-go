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

package font_test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/font"
)

func TestGoldenBitmapFontPixels(t *testing.T) {
	for _, bpp := range []int{1, 2, 4, 8} {
		t.Run(string(rune('0'+bpp)), func(t *testing.T) {
			f, err := font.LoadBitmapFontFile(filepath.Join("testdata", "fonts", "bpp"+string(rune('0'+bpp))+".caf"))
			if err != nil {
				t.Fatal(err)
			}
			c := canvas.MustNew(1, 1)
			if err := c.SetColor(canvas.White, canvas.Black); err != nil {
				t.Fatal(err)
			}
			c.Put(0, 0, 'A')
			img := image.NewRGBA(image.Rect(0, 0, 2, 2))
			f.Render(c, img)

			want, err := os.ReadFile(filepath.Join("testdata", "golden", "font-bpp"+string(rune('0'+bpp))+".argb"))
			if err != nil {
				t.Fatal(err)
			}
			if len(want) != 16 {
				t.Fatalf("golden length = %d", len(want))
			}
			for i := 0; i < 4; i++ {
				// C writes ARGB; Go writes RGBA.
				got := img.Pix[i*4 : i*4+4]
				if got[0] != want[i*4+1] || got[1] != want[i*4+2] || got[2] != want[i*4+3] || got[3] != want[i*4+0] {
					t.Fatalf("pixel %d = %v, want ARGB %v", i, got, want[i*4:i*4+4])
				}
			}
		})
	}
}
