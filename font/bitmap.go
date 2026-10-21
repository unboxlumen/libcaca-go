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
	"image/color"
	"image/draw"
	"os"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// BitmapFont is a font in libcaca's native .data format.
type BitmapFont struct {
	cellW, cellH int
	bpp          int
	blocks       []blockInfo
	glyphs       []glyphInfo
	data         []byte
}

type blockInfo struct {
	start, stop, index uint32
}

type glyphInfo struct {
	width, height uint16
	offset        uint32
}

// LoadBitmapFont parses a libcaca native font blob (the .data format used by
// mono9.data and monobold12.data).
func LoadBitmapFont(data []byte) (*BitmapFont, error) {
	if len(data) < 32 || data[0] != 0xca || data[1] != 0xca || data[2] != 'F' || data[3] != 'T' {
		return nil, fmt.Errorf("font: bad caca font header")
	}
	controlSize := binary.BigEndian.Uint32(data[4:8])
	dataSize := binary.BigEndian.Uint32(data[8:12])
	blocks := int(binary.BigEndian.Uint16(data[14:16]))
	glyphs := int(binary.BigEndian.Uint32(data[16:20]))
	bpp := int(binary.BigEndian.Uint16(data[20:22]))
	w := int(binary.BigEndian.Uint16(data[22:24]))
	h := int(binary.BigEndian.Uint16(data[24:26]))
	flags := binary.BigEndian.Uint16(data[30:32])

	if blocks < 0 || glyphs < 0 || blocks > 1<<16 || glyphs > 1<<20 {
		return nil, fmt.Errorf("font: unreasonable block/glyph count")
	}
	if bpp != 1 && bpp != 2 && bpp != 4 && bpp != 8 {
		return nil, fmt.Errorf("font: unsupported bpp %d", bpp)
	}
	if flags&1 == 0 {
		return nil, fmt.Errorf("font: invalid flags")
	}
	if uint64(controlSize) > uint64(len(data)-4) ||
		uint64(4)+uint64(controlSize)+uint64(dataSize) > uint64(len(data)) {
		return nil, fmt.Errorf("font: truncated data")
	}

	f := &BitmapFont{cellW: w, cellH: h, bpp: bpp}
	off := 32
	for i := 0; i < blocks; i++ {
		if off+12 > 4+int(controlSize) {
			return nil, fmt.Errorf("font: truncated block table")
		}
		b := blockInfo{
			start: binary.BigEndian.Uint32(data[off : off+4]),
			stop:  binary.BigEndian.Uint32(data[off+4 : off+8]),
			index: binary.BigEndian.Uint32(data[off+8 : off+12]),
		}
		if b.start > b.stop {
			return nil, fmt.Errorf("font: bad block range")
		}
		f.blocks = append(f.blocks, b)
		off += 12
	}
	for i := 0; i < glyphs; i++ {
		if off+8 > 4+int(controlSize) {
			return nil, fmt.Errorf("font: truncated glyph table")
		}
		g := glyphInfo{
			width:  binary.BigEndian.Uint16(data[off : off+2]),
			height: binary.BigEndian.Uint16(data[off+2 : off+4]),
			offset: binary.BigEndian.Uint32(data[off+4 : off+8]),
		}
		need := (uint64(g.width)*uint64(g.height)*uint64(bpp) + 7) / 8
		if uint64(g.offset)+need > uint64(dataSize) {
			return nil, fmt.Errorf("font: glyph out of range")
		}
		f.glyphs = append(f.glyphs, g)
		off += 8
	}
	f.data = data[4+int(controlSize) : 4+int(controlSize)+int(dataSize)]
	return f, nil
}

// LoadBitmapFontFile reads a libcaca font from disk.
func LoadBitmapFontFile(path string) (*BitmapFont, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBitmapFont(b)
}

// Metrics implements Font.
func (f *BitmapFont) Metrics() (int, int) { return f.cellW, f.cellH }

// Render implements Font.
func (f *BitmapFont) Render(c *canvas.Canvas, dst *image.RGBA) {
	if dst == nil || f == nil {
		return
	}
	cw, ch := c.Size()
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			cell := c.CellAt(x, y)
			char := cell.Ch
			if char == canvas.MagicFullwidth {
				continue
			}
			g, alpha, ok := f.glyphFor(char)
			if !ok {
				continue
			}
			a := cell.A.ARGB64()
			bgA, bgR, bgG, bgB := int(a[0]), int(a[1]), int(a[2]), int(a[3])
			fgA, fgR, fgG, fgB := int(a[4]), int(a[5]), int(a[6]), int(a[7])

			if bgA != 0 {
				r := image.Rect(x*f.cellW, y*f.cellH, (x+1)*f.cellW, (y+1)*f.cellH)
				col := color.NRGBA{
					R: uint8(bgR * 0x11),
					G: uint8(bgG * 0x11),
					B: uint8(bgB * 0x11),
					A: uint8(bgA * 0x11),
				}
				op := draw.Src
				if bgA != 15 {
					op = draw.Over
				}
				draw.Draw(dst, r, image.NewUniform(col), image.Point{}, op)
			}

			gw, gh := int(g.width), int(g.height)
			for j := 0; j < gh; j++ {
				for i := 0; i < gw; i++ {
					p := int(alpha[j*gw+i])
					q := 255 - p
					r := (q*bgR + p*fgR) / 15
					gg := (q*bgG + p*fgG) / 15
					b := (q*bgB + p*fgB) / 15
					aa := (q*bgA + p*fgA) / 15
					px, py := x*f.cellW+i, y*f.cellH+j
					if px >= dst.Bounds().Dx() || py >= dst.Bounds().Dy() {
						continue
					}
					off := dst.PixOffset(px, py)
					dst.Pix[off+0] = uint8(clampByte(r))
					dst.Pix[off+1] = uint8(clampByte(gg))
					dst.Pix[off+2] = uint8(clampByte(b))
					dst.Pix[off+3] = uint8(clampByte(aa))
				}
			}
		}
	}
}

func (f *BitmapFont) glyphFor(ch rune) (glyphInfo, []byte, bool) {
	u := uint32(ch)
	for _, b := range f.blocks {
		if u >= b.start && u < b.stop {
			idx := int(b.index + u - b.start)
			if idx < 0 || idx >= len(f.glyphs) {
				return glyphInfo{}, nil, false
			}
			g := f.glyphs[idx]
			n := int(g.width) * int(g.height)
			need := (n*f.bpp + 7) / 8
			start := int(g.offset)
			if start < 0 || start+need > len(f.data) {
				return glyphInfo{}, nil, false
			}
			return g, unpack(f.data[start:start+need], n, f.bpp), true
		}
	}
	return glyphInfo{}, nil, false
}

func unpack(packed []byte, n, bpp int) []byte {
	out := make([]byte, n)
	perByte := 8 / bpp
	mask := (1 << bpp) - 1
	for i := 0; i < n; i++ {
		p := packed[i/perByte]
		shift := bpp * (perByte - 1 - i%perByte)
		v := int(p>>shift) & mask
		out[i] = byte(v * 0xff / mask)
	}
	return out
}

func clampByte(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
