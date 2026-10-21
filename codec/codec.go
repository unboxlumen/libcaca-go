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

// Package codec implements libcaca's import and export formats for the
// character canvas.
//
// The byte formats follow libcaca 0.99.beta20. In particular the ANSI,
// UTF-8 and HTML exporters keep the original escape sequences and markup so
// their output can be compared with the C implementation.
package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/font"
)

// Version is reported by the HTML exporter. It matches libcaca so golden
// files produced from the C library compare byte-for-byte.
const Version = "0.99.beta20"

var (
	// ErrFormat is returned for an unknown format name.
	ErrFormat = errors.New("codec: unknown format")
	// ErrInvalid is returned for malformed input.
	ErrInvalid = errors.New("codec: invalid data")
	// ErrUnsupported is returned for a format that is recognised but not
	// implemented by this port yet.
	ErrUnsupported = errors.New("codec: format not implemented")
)

type view struct {
	w, h  int
	cells []canvas.Cell
}

func viewOf(c *canvas.Canvas) view {
	return view{w: c.Width(), h: c.Height(), cells: c.Cells()}
}

func areaView(c *canvas.Canvas, x, y, w, h int) (view, error) {
	if w < 0 || h < 0 {
		return view{}, ErrInvalid
	}
	out := view{w: w, h: h, cells: make([]canvas.Cell, w*h)}
	blank := canvas.Cell{Ch: ' ', A: canvas.DefaultAttr()}
	for i := range out.cells {
		out.cells[i] = blank
	}
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			sx, sy := x+i, y+j
			if sx < 0 || sx >= c.Width() || sy < 0 || sy >= c.Height() {
				continue
			}
			out.cells[j*w+i] = c.CellAt(sx, sy)
		}
	}
	return out, nil
}

// Export renders the current frame of c in the named format.
func Export(c *canvas.Canvas, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "caca":
		return exportCaca(c)
	case "ansi":
		return exportANSI(viewOf(c)), nil
	case "utf8":
		return exportUTF8(viewOf(c), false), nil
	case "utf8cr":
		return exportUTF8(viewOf(c), true), nil
	case "html":
		return exportHTML(viewOf(c)), nil
	case "html3":
		return exportHTML3(viewOf(c)), nil
	case "bbfr":
		return exportBBFR(viewOf(c)), nil
	case "irc":
		return exportIRC(viewOf(c)), nil
	case "ps":
		return exportPS(viewOf(c)), nil
	case "svg":
		return exportSVG(viewOf(c)), nil
	case "tga":
		return exportTGA(viewOf(c), nil)
	case "troff":
		return exportTroff(viewOf(c)), nil
	default:
		return nil, ErrFormat
	}
}

// ExportWithFont is Export, but uses f for formats that need a bitmap font
// (currently TGA). Pass nil for the default Go Mono font. Supplying a
// libcaca native font loaded with font.LoadBitmapFont reproduces the C
// exporter's pixels.
func ExportWithFont(c *canvas.Canvas, format string, f font.Font) ([]byte, error) {
	if strings.ToLower(format) == "tga" {
		return exportTGA(viewOf(c), f)
	}
	return Export(c, format)
}

// ExportArea renders a rectangular area in the named format.
func ExportArea(c *canvas.Canvas, x, y, w, h int, format string) ([]byte, error) {
	v, err := areaView(c, x, y, w, h)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(format) {
	case "caca":
		return exportCacaView(v), nil
	case "ansi":
		return exportANSI(v), nil
	case "utf8":
		return exportUTF8(v, false), nil
	case "utf8cr":
		return exportUTF8(v, true), nil
	case "html":
		return exportHTML(v), nil
	case "html3":
		return exportHTML3(v), nil
	case "bbfr":
		return exportBBFR(v), nil
	case "irc":
		return exportIRC(v), nil
	case "ps":
		return exportPS(v), nil
	case "svg":
		return exportSVG(v), nil
	case "tga":
		return exportTGA(v, nil)
	case "troff":
		return exportTroff(v), nil
	default:
		return nil, ErrFormat
	}
}

// ExportTo writes Export's result to w.
func ExportTo(w io.Writer, c *canvas.Canvas, format string) error {
	data, err := Export(c, format)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// ImportArea imports data into a temporary canvas and blits the result at
// (x,y). It returns the number of consumed bytes.
func ImportArea(c *canvas.Canvas, x, y int, data []byte, format string) (int, error) {
	tmp := canvas.MustNew(0, 0)
	n, err := Import(tmp, data, format)
	if err != nil {
		return n, err
	}
	if n > 0 {
		if err := c.Blit(x, y, tmp, nil); err != nil {
			return n, err
		}
	}
	return n, nil
}

// ImportAreaFrom is ImportArea with a reader.
func ImportAreaFrom(c *canvas.Canvas, x, y int, r io.Reader, format string) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	return ImportArea(c, x, y, data, format)
}

// Formats returns the supported export format names.
func Formats() []string {
	return []string{"ansi", "bbfr", "caca", "html", "html3", "irc", "ps", "svg", "tga", "troff", "utf8", "utf8cr"}
}

// ExportAreaWithFont is ExportArea with an explicit font.
func ExportAreaWithFont(c *canvas.Canvas, x, y, w, h int, format string, f font.Font) ([]byte, error) {
	v, err := areaView(c, x, y, w, h)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(format) == "tga" {
		return exportTGA(v, f)
	}
	return ExportArea(c, x, y, w, h, format)
}

// ImportFormats returns the supported import format names.
func ImportFormats() []string {
	return []string{"ansi", "bin", "caca", "text", "utf8"}
}

// Import loads data into c. It returns the number of consumed bytes.
func Import(c *canvas.Canvas, data []byte, format string) (int, error) {
	switch strings.ToLower(format) {
	case "caca":
		return importCaca(c, data)
	case "text":
		return importText(c, data)
	case "ansi":
		return importANSI(c, data, false)
	case "utf8":
		return importANSI(c, data, true)
	case "bin":
		return importBIN(c, data)
	case "":
		return importAuto(c, data)
	default:
		return 0, ErrFormat
	}
}

// ImportFrom reads all of r and imports it.
func ImportFrom(c *canvas.Canvas, r io.Reader, format string) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	return Import(c, data, format)
}

func importAuto(c *canvas.Canvas, data []byte) (int, error) {
	if len(data) >= 4 && data[0] == 0xca && data[1] == 0xca && data[2] == 'C' && data[3] == 'V' {
		return importCaca(c, data)
	}
	for i := 0; i+1 < len(data); i++ {
		if data[i] == 0x1b && data[i+1] == '[' {
			return importANSI(c, data, false)
		}
	}
	// Heuristic for BIN: many spaces at even offsets.
	j, k := 0, 0
	for i := 0; i < len(data); i += 2 {
		if data[i] == ' ' {
			j++
		}
		if i+1 < len(data) && data[i+1] == ' ' {
			k++
		}
	}
	if j > 10 && j > len(data)/40 && k < 10 {
		return importBIN(c, data)
	}
	return importText(c, data)
}

// ---- native caca format -------------------------------------------------

func exportCaca(c *canvas.Canvas) ([]byte, error) {
	frames := c.FrameCount()
	var b bytes.Buffer

	b.Write([]byte{0xca, 0xca, 'C', 'V'})
	writeBE32(&b, uint32(16+32*frames))
	writeBE32(&b, uint32(c.Width()*c.Height()*8*frames))
	writeBE16(&b, 1)
	writeBE32(&b, uint32(frames))
	writeBE16(&b, 0)

	for i := 0; i < frames; i++ {
		f, err := c.FrameAt(i)
		if err != nil {
			return nil, err
		}
		fx, fy := f.Pos()
		hx, hy := f.Handle()
		writeBE32(&b, uint32(f.Width()))
		writeBE32(&b, uint32(f.Height()))
		writeBE32(&b, 0)
		writeBE32(&b, uint32(f.Attr()))
		writeBE32(&b, uint32(int32(fx)))
		writeBE32(&b, uint32(int32(fy)))
		writeBE32(&b, uint32(int32(hx)))
		writeBE32(&b, uint32(int32(hy)))
	}

	for i := 0; i < frames; i++ {
		f, _ := c.FrameAt(i)
		for _, cell := range f.Cells() {
			writeBE32(&b, uint32(cell.Ch))
			writeBE32(&b, uint32(cell.A))
		}
	}
	return b.Bytes(), nil
}

func exportCacaView(v view) []byte {
	var b bytes.Buffer
	b.Write([]byte{0xca, 0xca, 'C', 'V'})
	writeBE32(&b, 16+32)
	writeBE32(&b, uint32(v.w*v.h*8))
	writeBE16(&b, 1)
	writeBE32(&b, 1)
	writeBE16(&b, 0)
	writeBE32(&b, uint32(v.w))
	writeBE32(&b, uint32(v.h))
	writeBE32(&b, 0)
	writeBE32(&b, uint32(canvas.DefaultAttr()))
	writeBE32(&b, 0)
	writeBE32(&b, 0)
	writeBE32(&b, 0)
	writeBE32(&b, 0)
	for _, cell := range v.cells {
		writeBE32(&b, uint32(cell.Ch))
		writeBE32(&b, uint32(cell.A))
	}
	return b.Bytes()
}

func importCaca(c *canvas.Canvas, data []byte) (int, error) {
	if len(data) < 20 {
		return 0, ErrInvalid
	}
	if data[0] != 0xca || data[1] != 0xca || data[2] != 'C' || data[3] != 'V' {
		return 0, ErrInvalid
	}
	controlSize := binary.BigEndian.Uint32(data[4:8])
	dataSize := binary.BigEndian.Uint32(data[8:12])
	version := binary.BigEndian.Uint16(data[12:14])
	frames := binary.BigEndian.Uint32(data[14:18])
	flags := binary.BigEndian.Uint16(data[18:20])
	_ = version
	_ = flags
	if frames == 0 || frames > 1<<20 {
		return 0, ErrInvalid
	}
	total := uint64(4) + uint64(controlSize) + uint64(dataSize)
	if total > uint64(len(data)) {
		return 0, ErrInvalid
	}
	if controlSize < 16+frames*32 {
		return 0, ErrInvalid
	}

	type frameInfo struct {
		w, h, duration, attr uint32
		x, y, hx, hy         int32
	}
	infos := make([]frameInfo, frames)
	var xmin, ymin, xmax, ymax int32
	for f := uint32(0); f < frames; f++ {
		off := 4 + 16 + int(f)*32
		infos[f] = frameInfo{
			w:  binary.BigEndian.Uint32(data[off : off+4]),
			h:  binary.BigEndian.Uint32(data[off+4 : off+8]),
			x:  int32(binary.BigEndian.Uint32(data[off+16 : off+20])),
			y:  int32(binary.BigEndian.Uint32(data[off+20 : off+24])),
			hx: int32(binary.BigEndian.Uint32(data[off+24 : off+28])),
			hy: int32(binary.BigEndian.Uint32(data[off+28 : off+32])),
		}
		infos[f].duration = binary.BigEndian.Uint32(data[off+8 : off+12])
		infos[f].attr = binary.BigEndian.Uint32(data[off+12 : off+16])
		if infos[f].w > 1<<16 || infos[f].h > 1<<16 {
			return 0, ErrInvalid
		}
		if -infos[f].hx < xmin {
			xmin = -infos[f].hx
		}
		if -infos[f].hy < ymin {
			ymin = -infos[f].hy
		}
		if int32(infos[f].w)-infos[f].hx > xmax {
			xmax = int32(infos[f].w) - infos[f].hx
		}
		if int32(infos[f].h)-infos[f].hy > ymax {
			ymax = int32(infos[f].h) - infos[f].hy
		}
	}

	cw, ch := int(xmax-xmin), int(ymax-ymin)
	if cw < 0 || ch < 0 || (cw != 0 && ch > int(^uint(0)>>1)/cw) {
		return 0, ErrInvalid
	}
	if err := c.Resize(cw, ch); err != nil {
		return 0, err
	}
	for c.FrameCount() < int(frames) {
		if err := c.CreateFrame(c.FrameCount()); err != nil {
			return 0, err
		}
	}

	offset := uint32(0)
	for f := uint32(0); f < frames; f++ {
		info := infos[f]
		if err := c.SetFrame(int(f)); err != nil {
			return 0, err
		}
		c.SetAttr(canvas.Attr(info.attr))
		c.GotoXY(int(info.x), int(info.y))
		c.SetHandle(int(info.hx), int(info.hy))

		cellCount := uint64(info.w) * uint64(info.h)
		if uint64(offset)+cellCount*8 > uint64(dataSize) {
			return 0, ErrInvalid
		}
		base := 4 + int(controlSize)
		for n := uint64(0); n < cellCount; n++ {
			p := base + int(offset) + int(n)*8
			chr := rune(binary.BigEndian.Uint32(data[p : p+4]))
			attr := canvas.Attr(binary.BigEndian.Uint32(data[p+4 : p+8]))
			nx := int(n%uint64(info.w)) - int(info.hx) - int(xmin)
			ny := int(n/uint64(info.w)) - int(info.hy) - int(ymin)
			c.SetCell(nx, ny, canvas.Cell{Ch: chr, A: attr})
		}
		offset += uint32(cellCount * 8)
	}

	// libcaca shifts frame cursors/handles into the new coordinate space.
	for f := 0; f < int(frames); f++ {
		if err := c.SetFrame(f); err != nil {
			return 0, err
		}
		fr, _ := c.FrameAt(f)
		fx, fy := fr.Pos()
		c.GotoXY(fx-int(infos[f].hx), fy-int(infos[f].hy))
		c.SetHandle(int(-xmin), int(-ymin))
	}
	c.SetFrame(0)
	return int(total), nil
}

func writeBE32(b *bytes.Buffer, v uint32) {
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], v)
	b.Write(tmp[:])
}

func writeBE16(b *bytes.Buffer, v uint16) {
	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], v)
	b.Write(tmp[:])
}

// ---- small helpers ------------------------------------------------------

func writeRune(b *bytes.Buffer, r rune) {
	b.WriteRune(r)
}

func ansiPaletteIndex(c canvas.Color) byte {
	// ANSI code order used by the ANSI/UTF-8 exporters.
	palette := [16]byte{0, 4, 2, 6, 1, 5, 3, 7, 8, 12, 10, 14, 9, 13, 11, 15}
	if c < 0x10 {
		return palette[c]
	}
	return 0x10
}

func colorOr(c canvas.Color, fallback canvas.Color) byte {
	if c < 0x10 {
		return byte(c)
	}
	return byte(fallback)
}

func writef(b *bytes.Buffer, format string, args ...any) {
	fmt.Fprintf(b, format, args...)
}
