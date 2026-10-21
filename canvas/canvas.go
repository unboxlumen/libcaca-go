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

import (
	"fmt"
	"math"
)

// Canvas is a character grid with one or more animation frames.
//
// A Canvas is not safe for concurrent use. That matches libcaca; callers
// that share one across goroutines must synchronise themselves.
type Canvas struct {
	frames        []*Frame
	cur           int
	dirty         []Rect
	dirtyDisabled int
	autoinc       int

	// resizeCb, when non-nil, may refuse a resize by returning false.
	resizeCb func() bool
	managed  bool
}

// MaxCells caps the number of cells a canvas may allocate. It keeps a
// malicious import from turning a small file into a huge allocation.
const MaxCells = 16 << 20

// New creates a canvas of the given size. Both dimensions must be
// non-negative and their product must not overflow int or exceed MaxCells.
func New(w, h int) (*Canvas, error) {
	if w < 0 || h < 0 {
		return nil, ErrInvalidSize
	}
	if w != 0 && h > math.MaxInt/w {
		return nil, ErrInvalidSize
	}
	if w != 0 && h > MaxCells/w {
		return nil, ErrTooLarge
	}
	c := &Canvas{
		frames: []*Frame{{
			name:    "frame#00000000",
			curAttr: DefaultAttr(),
		}},
	}
	if err := c.resizeFrames(w, h); err != nil {
		return nil, err
	}
	return c, nil
}

// MustNew is New but panics on invalid dimensions. It is meant for tests and
// constants.
func MustNew(w, h int) *Canvas {
	c, err := New(w, h)
	if err != nil {
		panic(err)
	}
	return c
}

// Width returns the current frame width.
func (c *Canvas) Width() int { return c.frames[c.cur].w }

// Height returns the current frame height.
func (c *Canvas) Height() int { return c.frames[c.cur].h }

// Size returns the current frame dimensions.
func (c *Canvas) Size() (w, h int) { return c.Width(), c.Height() }

// FrameCount returns the number of frames.
func (c *Canvas) FrameCount() int { return len(c.frames) }

// CurrentFrame returns the current frame index.
func (c *Canvas) CurrentFrame() int { return c.cur }

// SetFrame selects the current frame.
func (c *Canvas) SetFrame(n int) error {
	if n < 0 || n >= len(c.frames) {
		return ErrNoFrame
	}
	c.cur = n
	return nil
}

// CreateFrame inserts a new frame at index n. Like libcaca, the new frame is
// a copy of the current frame (cells, attributes, cursor and handle).
func (c *Canvas) CreateFrame(n int) error {
	if n < 0 {
		n = 0
	} else if n > len(c.frames) {
		n = len(c.frames)
	}
	src := c.frame()
	c.autoinc++
	f := &Frame{
		w:       src.w,
		h:       src.h,
		curAttr: src.curAttr,
		x:       src.x,
		y:       src.y,
		handleX: src.handleX,
		handleY: src.handleY,
		name:    fmt.Sprintf("frame#%08x", c.autoinc),
	}
	f.cells = append([]Cell(nil), src.cells...)
	c.frames = append(c.frames, nil)
	copy(c.frames[n+1:], c.frames[n:])
	c.frames[n] = f
	if n <= c.cur {
		c.cur++
	}
	return nil
}

// FreeFrame removes frame n. At least one frame always remains.
func (c *Canvas) FreeFrame(n int) error {
	if n < 0 || n >= len(c.frames) || len(c.frames) == 1 {
		return ErrNoFrame
	}
	c.frames = append(c.frames[:n], c.frames[n+1:]...)
	switch {
	case c.cur == n:
		c.cur = 0
		c.addDirtyRect(0, 0, c.Width(), c.Height())
	case n < c.cur:
		c.cur--
	}
	return nil
}

// FrameName returns the name of the current frame.
func (c *Canvas) FrameName() string { return c.frames[c.cur].name }

// SetFrameName sets the current frame name.
func (c *Canvas) SetFrameName(name string) { c.frames[c.cur].name = name }

// Resize changes the canvas size. Existing contents are preserved to the
// extent of the new size; new cells are filled with spaces using the current
// attribute.
func (c *Canvas) Resize(w, h int) error {
	if w < 0 || h < 0 {
		return ErrInvalidSize
	}
	if w != 0 && h > math.MaxInt/w {
		return ErrInvalidSize
	}
	if w != 0 && h > MaxCells/w {
		return ErrTooLarge
	}
	if c.resizeCb != nil && !c.resizeCb() {
		return ErrManaged
	}
	return c.resizeFrames(w, h)
}

// SetResizeCallback installs a callback that may refuse future resizes.
// Pass nil to remove it.
func (c *Canvas) SetResizeCallback(cb func() bool) { c.resizeCb = cb }

// ManageCanvas locks the canvas against resizes, like caca_manage_canvas.
// Resizes are allowed only when cb is nil or returns true. A canvas can be
// managed once; UnmanageCanvas releases it.
func (c *Canvas) ManageCanvas(cb func() bool) error {
	if c.managed {
		return ErrManaged
	}
	c.managed = true
	c.resizeCb = cb
	return nil
}

// UnmanageCanvas releases a canvas locked by ManageCanvas.
func (c *Canvas) UnmanageCanvas() error {
	if !c.managed {
		return ErrNotManaged
	}
	c.managed = false
	c.resizeCb = nil
	return nil
}

func (c *Canvas) resizeFrames(w, h int) error {
	oldW, oldH := c.Width(), c.Height()
	if w == oldW && h == oldH {
		return nil
	}

	for _, f := range c.frames {
		ow, oh := f.w, f.h
		nw, nh := w, h

		// Shrinking: clipping happens against the new size.
		if nw < ow || nh < oh {
			// Dirty rectangles are clipped below, once the size is known.
		}

		cells := make([]Cell, nw*nh)
		blank := Cell{Ch: ' ', A: f.curAttr}
		for i := range cells {
			cells[i] = blank
		}

		copyW, copyH := minInt(ow, nw), minInt(oh, nh)
		for y := 0; y < copyH; y++ {
			copy(cells[y*nw:y*nw+copyW], f.cells[y*ow:y*ow+copyW])
		}

		f.w, f.h = nw, nh
		f.cells = cells
		if f.x > nw {
			f.x = nw
		}
		if f.y > nh {
			f.y = nh
		}
	}

	c.clipDirtyRects()

	if w > oldW {
		c.addDirtyRect(oldW, 0, w-oldW, minInt(oldH, h))
	}
	if h > oldH {
		c.addDirtyRect(0, oldH, minInt(oldW, w), h-oldH)
	}
	if w > oldW && h > oldH {
		c.addDirtyRect(oldW, oldH, w-oldW, h-oldH)
	}
	return nil
}

// Clear fills the current frame with spaces using the current attribute.
func (c *Canvas) Clear() {
	f := c.frame()
	f.fillBlank()
	c.addDirtyRect(0, 0, f.w, f.h)
}

// CurAttr returns the current drawing attribute.
func (c *Canvas) CurAttr() Attr { return c.frame().curAttr }

// GotoXY moves the frame cursor.
func (c *Canvas) GotoXY(x, y int) { c.frame().x, c.frame().y = x, y }

// WhereX returns the frame cursor X coordinate.
func (c *Canvas) WhereX() int { return c.frame().x }

// WhereY returns the frame cursor Y coordinate.
func (c *Canvas) WhereY() int { return c.frame().y }

// Handle returns the drawing handle.
func (c *Canvas) Handle() (x, y int) {
	f := c.frame()
	return f.handleX, f.handleY
}

// SetHandle sets the drawing handle used by Blit.
func (c *Canvas) SetHandle(x, y int) { c.frame().handleX, c.frame().handleY = x, y }

// Cells returns the current frame's cell storage. See Frame.Cells for the
// ownership rules.
func (c *Canvas) Cells() []Cell { return c.frames[c.cur].cells }

// CellAt returns the cell at (x,y) without bounds errors. Outside the canvas
// it returns a blank cell with the current attribute.
func (c *Canvas) CellAt(x, y int) Cell { return c.frames[c.cur].CellAt(x, y) }

// FrameAt returns frame n.
func (c *Canvas) FrameAt(n int) (*Frame, error) {
	if n < 0 || n >= len(c.frames) {
		return nil, ErrNoFrame
	}
	return c.frames[n], nil
}

// SetCell writes a whole cell at (x,y). It is intended for importers that
// already know the final character and attribute, including the fullwidth
// sentinel.
func (c *Canvas) SetCell(x, y int, cell Cell) {
	f := c.frame()
	if x < 0 || x >= f.w || y < 0 || y >= f.h {
		return
	}
	f.cells[y*f.w+x] = cell
	c.addDirtyRect(x, y, 1, 1)
}

func (c *Canvas) frame() *Frame { return c.frames[c.cur] }

// SetBoundaries crops or expands the canvas in place. All frames are
// affected; content outside the old canvas is blank. The current drawing
// attributes reset because the replacement frames are new.
func (c *Canvas) SetBoundaries(x, y, w, h int) error {
	if w < 0 || h < 0 {
		return ErrInvalidSize
	}
	if w != 0 && h > MaxCells/w {
		return ErrTooLarge
	}
	if c.resizeCb != nil && !c.resizeCb() {
		return ErrManaged
	}

	saved := c.cur
	nw := MustNew(w, h)
	for i := 1; i < len(c.frames); i++ {
		if err := nw.CreateFrame(i); err != nil {
			return err
		}
	}
	for i := 0; i < len(c.frames); i++ {
		if err := c.SetFrame(i); err != nil {
			return err
		}
		if err := nw.SetFrame(i); err != nil {
			return err
		}
		if err := nw.Blit(-x, -y, c, nil); err != nil {
			return err
		}
	}

	c.frames = nw.frames
	c.cur = saved
	if c.cur >= len(c.frames) {
		c.cur = len(c.frames) - 1
	}
	c.ClearDirtyRects()
	c.addDirtyRect(0, 0, w, h)
	return nil
}
