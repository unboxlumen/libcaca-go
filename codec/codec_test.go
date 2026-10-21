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
	"encoding/binary"
	"image"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func TestExportANSISimple(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'B')
	got := string(mustExport(t, c, "ansi"))
	want := "\x1b[0;37;40mAB\x1b[0m\r\n"
	if got != want {
		t.Fatalf("ansi = %q, want %q", got, want)
	}
}

func TestExportANSI16Color(t *testing.T) {
	c := canvas.MustNew(2, 1)
	if err := c.SetColor(canvas.Red, canvas.Blue); err != nil {
		t.Fatal(err)
	}
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'B')
	got := string(mustExport(t, c, "ansi"))
	want := "\x1b[0;31;44mAB\x1b[0m\r\n"
	if got != want {
		t.Fatalf("ansi = %q, want %q", got, want)
	}
}

func TestExportUTF8Simple(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'B')
	if got := string(mustExport(t, c, "utf8")); got != "AB\n" {
		t.Fatalf("utf8 = %q", got)
	}

	if err := c.SetColor(canvas.Red, canvas.Blue); err != nil {
		t.Fatal(err)
	}
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'B')
	if got := string(mustExport(t, c, "utf8")); got != "\x1b[0;31;44mAB\x1b[0m\n" {
		t.Fatalf("utf8 colour = %q", got)
	}
}

func TestCacaRoundTrip(t *testing.T) {
	c := canvas.MustNew(3, 2)
	if err := c.SetColor(canvas.Red, canvas.Blue); err != nil {
		t.Fatal(err)
	}
	c.Put(0, 0, 'a')
	c.Put(2, 1, 'z')

	data := mustExport(t, c, "caca")
	dst := canvas.MustNew(0, 0)
	n, err := Import(dst, data, "caca")
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatalf("consumed = %d, want %d", n, len(data))
	}
	if dst.Width() != 3 || dst.Height() != 2 {
		t.Fatalf("size = %dx%d", dst.Width(), dst.Height())
	}
	if got := dst.Get(0, 0); got != 'a' {
		t.Fatalf("cell = %q", got)
	}
	if got := dst.Get(2, 1); got != 'z' {
		t.Fatalf("cell = %q", got)
	}
	if got := dst.Get(0, 0); got != 'a' {
		t.Fatalf("cell = %q", got)
	}
	if dst.GetAttr(0, 0).FG() != canvas.Red || dst.GetAttr(0, 0).BG() != canvas.Blue {
		t.Fatalf("attr = %v", dst.GetAttr(0, 0))
	}
}

func TestCacaMultiFrame(t *testing.T) {
	c := canvas.MustNew(2, 2)
	c.Put(0, 0, 'a')
	if err := c.CreateFrame(1); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFrame(1); err != nil {
		t.Fatal(err)
	}
	c.Put(1, 1, 'b')

	data := mustExport(t, c, "caca")
	dst := canvas.MustNew(0, 0)
	if _, err := Import(dst, data, "caca"); err != nil {
		t.Fatal(err)
	}
	if dst.FrameCount() != 2 {
		t.Fatalf("frames = %d", dst.FrameCount())
	}
	if err := dst.SetFrame(0); err != nil {
		t.Fatal(err)
	}
	if got := dst.Get(0, 0); got != 'a' {
		t.Fatalf("frame 0 cell = %q", got)
	}
	if err := dst.SetFrame(1); err != nil {
		t.Fatal(err)
	}
	if got := dst.Get(1, 1); got != 'b' {
		t.Fatalf("frame 1 cell = %q", got)
	}
}

func TestImportText(t *testing.T) {
	c := canvas.MustNew(0, 0)
	if _, err := Import(c, []byte("ab\ncd"), "text"); err != nil {
		t.Fatal(err)
	}
	if c.Width() != 2 || c.Height() != 2 {
		t.Fatalf("size = %dx%d", c.Width(), c.Height())
	}
	if got := string([]rune{c.Get(0, 0), c.Get(1, 0), c.Get(0, 1), c.Get(1, 1)}); got != "abcd" {
		t.Fatalf("cells = %q", got)
	}
}

func TestImportBIN(t *testing.T) {
	c := canvas.MustNew(0, 0)
	data := make([]byte, 160)
	data[0], data[1] = 'A', 0x1f
	if _, err := Import(c, data, "bin"); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'A' {
		t.Fatalf("cell = %q", got)
	}
	a := c.GetAttr(0, 0)
	if a.FG() != canvas.White || a.BG() != canvas.Blue {
		t.Fatalf("attr fg=%d bg=%d", a.FG(), a.BG())
	}
}

func TestImportANSISGR(t *testing.T) {
	c := canvas.MustNew(0, 0)
	if _, err := Import(c, []byte("\x1b[31mX"), "ansi"); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'X' {
		t.Fatalf("cell = %q", got)
	}
	a := c.GetAttr(0, 0)
	if a.FG() != canvas.Red {
		t.Fatalf("fg = %d, want red", a.FG())
	}
	if a.BG() != canvas.Black {
		t.Fatalf("bg = %d, want black", a.BG())
	}
}

func TestImportANSIBold(t *testing.T) {
	c := canvas.MustNew(0, 0)
	if _, err := Import(c, []byte("\x1b[1;34mX"), "ansi"); err != nil {
		t.Fatal(err)
	}
	a := c.GetAttr(0, 0)
	// ANSI blue is index 4 in libcaca's ordering; bold turns it bright (12).
	if a.FG() != canvas.LightBlue {
		t.Fatalf("fg = %d, want light blue", a.FG())
	}
}

func TestExportHTMLContainsSpan(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, '<')
	c.Put(1, 0, '&')
	out := string(mustExport(t, c, "html"))
	if !bytes.Contains([]byte(out), []byte("&lt;")) || !bytes.Contains([]byte(out), []byte("&amp;")) {
		t.Fatalf("html escaping missing: %q", out)
	}
}

func TestExportSVGShape(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	out := string(mustExport(t, c, "svg"))
	if !bytes.HasPrefix([]byte(out), []byte("<?xml")) || !bytes.Contains([]byte(out), []byte("</svg>")) {
		t.Fatalf("svg shape: %q", out)
	}
}

func mustExport(t *testing.T, c *canvas.Canvas, format string) []byte {
	t.Helper()
	data, err := Export(c, format)
	if err != nil {
		t.Fatalf("export %s: %v", format, err)
	}
	return data
}

type solidFont struct{}

func (solidFont) Metrics() (int, int) { return 2, 2 }
func (solidFont) Render(c *canvas.Canvas, dst *image.RGBA) {
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = 255
		dst.Pix[i+3] = 255
	}
}

func TestExportWithFontTGA(t *testing.T) {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	data, err := ExportWithFont(c, "tga", solidFont{})
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 18+4*2*4 {
		t.Fatalf("tga length = %d", len(data))
	}
	if data[12] != 4 || data[14] != 2 {
		t.Fatalf("tga size = %dx%d", binary.LittleEndian.Uint16(data[12:14]), binary.LittleEndian.Uint16(data[14:16]))
	}
	// TGA pixels are B,G,R,A; the test font writes pure red.
	if data[18] != 0 || data[19] != 0 || data[20] != 255 || data[21] != 255 {
		t.Fatalf("first pixel = %v", data[18:22])
	}

	// A non-TGA format ignores the font.
	got, err := ExportWithFont(c, "utf8", solidFont{})
	if err != nil {
		t.Fatal(err)
	}
	want := mustExport(t, c, "utf8")
	if !bytes.Equal(got, want) {
		t.Fatalf("utf8 changed: %q vs %q", got, want)
	}
}
