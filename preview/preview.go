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

// Package preview renders an image as character art and then into a bitmap.
//
// It is the bridge between the dither/canvas core and the GUI preview panes,
// which already know how to show a bitmap but not how to paint character cells.
package preview

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/dither"
	"github.com/unboxlumen/libcaca-go/font"
)

// Options configures the character-art bitmap.
type Options struct {
	// Dither controls sampling, palette and glyph ramp.
	Dither dither.Options
	// Font renders the cells. Nil means font.Default() (Go Mono).
	Font font.Font
	// Cols and Rows are the character grid. They default to 72x36.
	Cols, Rows int
}

// DefaultOptions is a readable default for preview panes.
func DefaultOptions() Options {
	o := dither.DefaultOptions()
	o.Charset = dither.Blocks
	return Options{Dither: o, Cols: 72, Rows: 36}
}

// Render dithers img into a character grid and renders that grid with f.
func Render(img image.Image, opts Options) (*image.RGBA, error) {
	cols, rows := opts.Cols, opts.Rows
	if cols <= 0 {
		cols = 72
	}
	if rows <= 0 {
		rows = 36
	}
	dopts := opts.Dither
	if dopts == (dither.Options{}) {
		dopts = DefaultOptions().Dither
	}

	d, err := dither.New(img, dopts)
	if err != nil {
		return nil, err
	}
	c, err := canvas.New(cols, rows)
	if err != nil {
		return nil, err
	}
	if err := d.Draw(c, 0, 0, cols, rows); err != nil {
		return nil, err
	}

	f := opts.Font
	if f == nil {
		f, err = font.Default()
		if err != nil {
			return nil, err
		}
	}
	fw, fh := f.Metrics()
	if fw <= 0 || fh <= 0 {
		fw, fh = 1, 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, cols*fw, rows*fh))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	f.Render(c, dst)
	return dst, nil
}
