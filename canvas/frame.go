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

package canvas

// Cell is one character cell.
type Cell struct {
	Ch rune
	A  Attr
}

// Frame is one animation frame of a canvas.
type Frame struct {
	w, h    int
	cells   []Cell
	x, y    int
	handleX int
	handleY int
	curAttr Attr
	name    string
}

// Width returns the frame width in cells.
func (f *Frame) Width() int { return f.w }

// Height returns the frame height in cells.
func (f *Frame) Height() int { return f.h }

// Cells returns the frame's cell storage. The slice is owned by the frame;
// callers must not retain it across resizes and must not write to it
// directly (use Canvas.Put / Canvas.PutAttr so dirty rectangles stay
// correct).
func (f *Frame) Cells() []Cell { return f.cells }

// CellAt returns the cell at (x,y), or a blank cell outside the frame.
func (f *Frame) CellAt(x, y int) Cell {
	if x < 0 || x >= f.w || y < 0 || y >= f.h {
		return Cell{Ch: ' ', A: f.curAttr}
	}
	return f.cells[y*f.w+x]
}

// Name returns the frame name.
func (f *Frame) Name() string { return f.name }

// SetName changes the frame name.
func (f *Frame) SetName(name string) { f.name = name }

// Pos returns the frame cursor position.
func (f *Frame) Pos() (x, y int) { return f.x, f.y }

// Handle returns the frame drawing handle.
func (f *Frame) Handle() (x, y int) { return f.handleX, f.handleY }

// Attr returns the frame's current drawing attribute.
func (f *Frame) Attr() Attr { return f.curAttr }

func (f *Frame) index(x, y int) int { return y*f.w + x }

func (f *Frame) fillBlank() {
	a := f.curAttr
	for i := range f.cells {
		f.cells[i] = Cell{Ch: ' ', A: a}
	}
}
