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

package font

import (
	"encoding/binary"
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func tinyFigfont() string {
	var b strings.Builder
	b.WriteString("flf2a$ 2 1 2 -1 0 0 0 0\n")
	for ch := rune(32); ch <= 126; ch++ {
		if ch == ' ' {
			b.WriteString("  @\n  @\n")
			continue
		}
		fmt.Fprintf(&b, "%c%c@\n%c%c@\n", ch, ch, ch, ch)
	}
	return b.String()
}

func TestParseAndRenderFigfont(t *testing.T) {
	f, err := ParseFigfont(strings.NewReader(tinyFigfont()))
	if err != nil {
		t.Fatal(err)
	}
	if f.Height != 2 {
		t.Fatalf("height = %d", f.Height)
	}
	if f.Hardblank != '$' {
		t.Fatalf("hardblank = %q", f.Hardblank)
	}
	rows := f.Render("A")
	if len(rows) != 2 || rows[0] != "AA" || rows[1] != "AA" {
		t.Fatalf("rows = %#v", rows)
	}

	c := canvas.MustNew(2, 2)
	f.Draw(c, 0, 0, "A")
	if got := c.Get(0, 0); got != 'A' {
		t.Fatalf("cell = %q", got)
	}
	if got := c.Get(1, 1); got != 'A' {
		t.Fatalf("cell = %q", got)
	}
}

func TestTTFMetrics(t *testing.T) {
	f, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	w, h := f.Metrics()
	if w <= 0 || h <= 0 {
		t.Fatalf("metrics = %dx%d", w, h)
	}
}

func tinyBitmapFont() []byte {
	b := make([]byte, 32, 32+12+8+1)
	b[0], b[1], b[2], b[3] = 0xca, 0xca, 'F', 'T'
	control := uint32(28 + 12 + 8)
	binary.BigEndian.PutUint32(b[4:8], control)
	binary.BigEndian.PutUint32(b[8:12], 1)
	binary.BigEndian.PutUint16(b[12:14], 1)
	binary.BigEndian.PutUint16(b[14:16], 1)
	binary.BigEndian.PutUint32(b[16:20], 1)
	binary.BigEndian.PutUint16(b[20:22], 1) // bpp
	binary.BigEndian.PutUint16(b[22:24], 2) // width
	binary.BigEndian.PutUint16(b[24:26], 2) // height
	binary.BigEndian.PutUint16(b[26:28], 2)
	binary.BigEndian.PutUint16(b[28:30], 2)
	binary.BigEndian.PutUint16(b[30:32], 1)
	block := make([]byte, 12)
	binary.BigEndian.PutUint32(block[0:4], 'A')
	binary.BigEndian.PutUint32(block[4:8], 'B')
	binary.BigEndian.PutUint32(block[8:12], 0)
	b = append(b, block...)
	glyph := make([]byte, 8)
	binary.BigEndian.PutUint16(glyph[0:2], 2)
	binary.BigEndian.PutUint16(glyph[2:4], 2)
	binary.BigEndian.PutUint32(glyph[4:8], 0)
	b = append(b, glyph...)
	b = append(b, 0x90) // 1 0 / 0 1
	return b
}

func TestBitmapFontRender(t *testing.T) {
	f, err := LoadBitmapFont(tinyBitmapFont())
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
	if img.Pix[0] < 200 || img.Pix[1] < 200 || img.Pix[2] < 200 {
		t.Fatalf("top-left pixel should be white, got %v", img.Pix[:4])
	}
	if img.Pix[4] > 50 || img.Pix[5] > 50 || img.Pix[6] > 50 {
		t.Fatalf("top-right pixel should be black, got %v", img.Pix[4:8])
	}
}

func TestHSmushRules(t *testing.T) {
	cases := []struct {
		a, b rune
		rule int
		want rune
	}{
		{'A', 'A', 0x01, 'A'},
		{0xa0, 0xa0, 0x01, 0},
		{'_', '|', 0x02, '|'},
		{'|', '_', 0x02, '|'},
		{'/', '[', 0x04, '['},
		{'[', '/', 0x04, '['},
		{'{', '}', 0x08, '|'},
		{'[', ']', 0x08, '|'},
		{'(', ')', 0x08, '|'},
		{'/', '\\', 0x10, '|'},
		{'\\', '/', 0x10, 'Y'},
		{'>', '<', 0x10, 'X'},
		{'A', 'B', 0x01, 0},
	}
	for _, tc := range cases {
		if got := hsmush(tc.a, tc.b, tc.rule); got != tc.want {
			t.Errorf("hsmush(%q,%q,%#x) = %q, want %q", tc.a, tc.b, tc.rule, got, tc.want)
		}
	}
}

func TestFigfontSetSmush(t *testing.T) {
	f, err := ParseFigfont(strings.NewReader(tinyFigfont()))
	if err != nil {
		t.Fatal(err)
	}
	f.SetSmush("none")
	if rows := f.Render("AB"); len(rows) != 2 || rows[0] != "AABB" {
		t.Fatalf("none rows = %#v", rows)
	}
	f.SetSmush("kern")
	if rows := f.Render("AB"); len(rows) != 2 {
		t.Fatalf("kern rows = %#v", rows)
	}
	f.SetSmush("bogus")
	if f.hmode != hNone {
		t.Fatalf("bogus mode = %v, want header default none", f.hmode)
	}
	f.SetWidth(2)
	if f.TermWidth != 2 {
		t.Fatalf("width = %d", f.TermWidth)
	}
}
