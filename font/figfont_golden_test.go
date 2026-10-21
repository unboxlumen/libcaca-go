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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/font"
)

func TestGoldenFigfont(t *testing.T) {
	cases := []struct {
		name, file, text string
		width            int
	}{
		{"full", "fig-full.flf", "AB", 1000},
		{"kern", "fig-kern.flf", "AB", 1000},
		{"smush", "fig-smush.flf", "AB", 1000},
		{"newline", "fig-full.flf", "AB\nAB", 1000},
		{"wrap", "fig-full.flf", "ABA", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := font.LoadFigfont(filepath.Join("testdata", "figfonts", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			f.TermWidth = tc.width
			rows := f.Render(tc.text)
			w, h := 0, len(rows)
			for _, row := range rows {
				if n := len([]rune(row)); n > w {
					w = n
				}
			}
			c := canvas.MustNew(w, h)
			f.Draw(c, 0, 0, tc.text)
			got, err := codec.Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", "golden", "fig-"+tc.name+".caca"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("fig-%s differs\nrows=%q\ngot=%d want=%d", tc.name, rows, len(got), len(want))
			}
		})
	}
}
