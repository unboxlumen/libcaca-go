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

// Package font renders a character canvas into a bitmap.
//
// The default font is Go Mono (BSD licensed), loaded from
// golang.org/x/image/font/gofont. The libcaca native bitmap-font format is
// intentionally not required for the pure Go build; a parser can be added
// later without changing this interface.
package font

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// Font is a monospace bitmap font.
type Font interface {
	// Metrics returns the cell width and height in pixels.
	Metrics() (w, h int)
	// Render draws c into dst.
	Render(c *canvas.Canvas, dst *image.RGBA)
}

// TTF is a TrueType-backed monospace font.
type TTF struct {
	face   font.Face
	cellW  int
	cellH  int
	ascent int
}

// NewTTF parses a TrueType font and selects a monospace face.
func NewTTF(data []byte, size, dpi float64) (*TTF, error) {
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	m := face.Metrics()
	adv, ok := face.GlyphAdvance('M')
	if !ok {
		adv = fixed.I(8)
	}
	return &TTF{
		face:   face,
		cellW:  fixedToInt(adv),
		cellH:  fixedToInt(m.Height),
		ascent: fixedToInt(m.Ascent),
	}, nil
}

// NewGoMono returns Go Mono at the given point size.
func NewGoMono(size float64) (*TTF, error) {
	return NewTTF(gomono.TTF, size, 96)
}

var (
	defaultOnce sync.Once
	defaultFont *TTF
	defaultErr  error
)

// Default returns the shared Go Mono font used by exporters.
func Default() (*TTF, error) {
	defaultOnce.Do(func() {
		defaultFont, defaultErr = NewGoMono(12)
	})
	return defaultFont, defaultErr
}

// Metrics implements Font.
func (f *TTF) Metrics() (int, int) { return f.cellW, f.cellH }

// Render implements Font.
func (f *TTF) Render(c *canvas.Canvas, dst *image.RGBA) {
	if dst == nil || f == nil {
		return
	}
	cw, ch := c.Size()
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			cell := c.CellAt(x, y)
			a := cell.A.ARGB64()

			// Background.
			if a[0] != 0 {
				bg := color.NRGBA{
					R: uint8(a[1] * 0x11),
					G: uint8(a[2] * 0x11),
					B: uint8(a[3] * 0x11),
					A: uint8(a[0] * 0x11),
				}
				r := image.Rect(x*f.cellW, y*f.cellH, (x+1)*f.cellW, (y+1)*f.cellH)
				op := draw.Src
				if bg.A != 0xff {
					op = draw.Over
				}
				draw.Draw(dst, r, image.NewUniform(bg), image.Point{}, op)
			}

			if cell.Ch == canvas.MagicFullwidth || cell.Ch == ' ' || cell.Ch < 0x20 {
				continue
			}

			fg := color.NRGBA{
				R: uint8(a[5] * 0x11),
				G: uint8(a[6] * 0x11),
				B: uint8(a[7] * 0x11),
				A: uint8(a[4] * 0x11),
			}
			d := font.Drawer{
				Dst:  dst,
				Src:  image.NewUniform(fg),
				Face: f.face,
				Dot:  fixed.P(x*f.cellW, y*f.cellH+f.ascent),
			}
			d.DrawString(string(cell.Ch))
		}
	}
}

func fixedToInt(v fixed.Int26_6) int {
	n := int(v >> 6)
	if v&63 != 0 {
		n++
	}
	return n
}
