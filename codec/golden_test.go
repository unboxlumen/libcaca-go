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

package codec

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// buildGoldenScene matches /tmp/caca-oracle/oracle.c's build().
func buildGoldenScene(t *testing.T) *canvas.Canvas {
	t.Helper()
	c := canvas.MustNew(12, 4)
	if err := c.SetColor(canvas.LightGray, canvas.Black); err != nil {
		t.Fatal(err)
	}
	c.FillBox(0, 0, 12, 4, ' ')
	c.PutString(0, 0, "libcaca")
	if err := c.SetColor(canvas.Red, canvas.Blue); err != nil {
		t.Fatal(err)
	}
	c.PutString(0, 1, "Go port")
	if err := c.SetColor(canvas.Yellow, canvas.DarkGray); err != nil {
		t.Fatal(err)
	}
	c.Put(9, 2, 0x2f06)
	c.SetAttr(canvas.Attr(canvas.Bold | canvas.Underline))
	c.PutString(0, 3, "styles")
	c.DrawBox(7, 0, 5, 4, '#')
	return c
}

func TestGoldenExports(t *testing.T) {
	cases := []struct {
		file   string
		format string
	}{
		{"base.ansi", "ansi"},
		{"base.utf8", "utf8"},
		{"base.utf8cr", "utf8cr"},
		{"base.html", "html"},
		{"base.html3", "html3"},
		{"base.bbfr", "bbfr"},
		{"base.irc", "irc"},
		{"base.ps", "ps"},
		{"base.svg", "svg"},
		{"base.troff", "troff"},
		{"base.caca", "caca"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			c := buildGoldenScene(t)
			got, err := Export(c, tc.format)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.file, diff(got, want))
			}
		})
	}
}

func TestGoldenTransforms(t *testing.T) {
	cases := []struct {
		file string
		op   func(*canvas.Canvas)
	}{
		{"flip.caca", (*canvas.Canvas).Flip},
		{"flop.caca", (*canvas.Canvas).Flop},
		{"rot180.caca", (*canvas.Canvas).Rotate180},
		{"rotleft.caca", (*canvas.Canvas).RotateLeft},
		{"rotright.caca", (*canvas.Canvas).RotateRight},
		{"stretchleft.caca", (*canvas.Canvas).StretchLeft},
		{"stretchright.caca", (*canvas.Canvas).StretchRight},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			c := buildGoldenScene(t)
			tc.op(c)
			got, err := Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.file, diff(got, want))
			}
		})
	}
}

func TestGoldenArea(t *testing.T) {
	c := buildGoldenScene(t)
	got, err := ExportArea(c, 2, 1, 6, 3, "caca")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/golden/area.caca")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("area.caca differs\n%s", diff(got, want))
	}
}

func diff(got, want []byte) string {
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

func TestGoldenBoundaries(t *testing.T) {
	cases := []struct {
		file       string
		x, y, w, h int
	}{
		{"bounds-crop.caca", 3, 1, 7, 3},
		{"bounds-expand.caca", -2, -1, 16, 7},
		{"bounds-mixed.caca", 2, -2, 10, 8},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			c := buildGoldenScene(t)
			if err := c.SetBoundaries(tc.x, tc.y, tc.w, tc.h); err != nil {
				t.Fatal(err)
			}
			got, err := Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.file, diff(got, want))
			}
		})
	}
}

func TestGoldenImportArea(t *testing.T) {
	c := buildGoldenScene(t)
	n, err := ImportArea(c, 2, 1, []byte("AB\nCD"), "text")
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("consumed = %d, want 5", n)
	}
	got, err := Export(c, "caca")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/golden/import-area.caca")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("import-area.caca differs\n%s", diff(got, want))
	}
}

func TestGoldenAreaFormats(t *testing.T) {
	for _, format := range []string{"ansi", "utf8"} {
		t.Run(format, func(t *testing.T) {
			c := buildGoldenScene(t)
			got, err := ExportArea(c, 2, 1, 6, 3, format)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/area." + format)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("area.%s differs\n%s", format, diff(got, want))
			}
		})
	}
}

func TestGoldenBoundariesFrames(t *testing.T) {
	c := buildGoldenScene(t)
	if err := c.CreateFrame(1); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFrame(1); err != nil {
		t.Fatal(err)
	}
	if err := c.SetColor(canvas.Green, canvas.Black); err != nil {
		t.Fatal(err)
	}
	c.PutString(0, 0, "frame2")
	if err := c.SetFrame(0); err != nil {
		t.Fatal(err)
	}
	if err := c.SetBoundaries(3, 1, 7, 3); err != nil {
		t.Fatal(err)
	}
	got, err := Export(c, "caca")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/golden/bounds-frames.caca")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("bounds-frames.caca differs\n%s", diff(got, want))
	}
}

func TestGoldenAutoDetect(t *testing.T) {
	bin := make([]byte, 320)
	for i := 0; i < len(bin); i += 2 {
		bin[i] = ' '
		bin[i+1] = 0x1f
	}
	native := canvas.MustNew(6, 1)
	if err := native.SetColor(canvas.Red, canvas.Blue); err != nil {
		t.Fatal(err)
	}
	native.PutString(0, 0, "native")
	nativeBlob, err := Export(native, "caca")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		data []byte
	}{
		{"auto-text.caca", []byte("hello")},
		{"auto-ansi.caca", []byte("\x1b[31mred")},
		{"auto-bin.caca", bin},
		{"auto-caca.caca", nativeBlob},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := canvas.MustNew(0, 0)
			if _, err := Import(c, tc.data, ""); err != nil {
				t.Fatal(err)
			}
			got, err := Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.name, diff(got, want))
			}
		})
	}
}
