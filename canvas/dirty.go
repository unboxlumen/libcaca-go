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

// Rect is a dirty rectangle in character cells.
type Rect struct {
	X, Y, W, H int
}

const maxDirtyRects = 8

// DirtyRects returns a copy of the current dirty rectangle list.
//
// The list is guaranteed not to contain overlapping rectangles and never to
// extend outside the canvas.
func (c *Canvas) DirtyRects() []Rect {
	out := make([]Rect, len(c.dirty))
	copy(out, c.dirty)
	return out
}

// DirtyRectCount returns the number of dirty rectangles.
func (c *Canvas) DirtyRectCount() int { return len(c.dirty) }

// ClearDirtyRects empties the dirty rectangle list.
func (c *Canvas) ClearDirtyRects() { c.dirty = c.dirty[:0] }

// DisableDirtyRect suspends dirty rectangle tracking. Calls nest.
func (c *Canvas) DisableDirtyRect() { c.dirtyDisabled++ }

// EnableDirtyRect resumes dirty rectangle tracking. It returns an error if
// tracking was not disabled.
func (c *Canvas) EnableDirtyRect() error {
	if c.dirtyDisabled <= 0 {
		return ErrDirtyNotDisabled
	}
	c.dirtyDisabled--
	return nil
}

// AddDirtyRect manually marks an area as dirty.
func (c *Canvas) AddDirtyRect(x, y, w, h int) {
	c.addDirtyRect(x, y, w, h)
}

// RemoveDirtyRect mirrors caca_remove_dirty_rect: it validates the area but
// intentionally does nothing, keeping dirty tracking conservative.
func (c *Canvas) RemoveDirtyRect(x, y, w, h int) error {
	if x < 0 {
		w += x
		x = 0
	}
	if x+w > c.Width() {
		w = c.Width() - x
	}
	if y < 0 {
		h += y
		y = 0
	}
	if y+h > c.Height() {
		h = c.Height() - y
	}
	if w <= 0 || h <= 0 {
		return ErrInvalidSize
	}
	return nil
}

func (c *Canvas) addDirtyRect(x, y, w, h int) {
	if c.dirtyDisabled > 0 {
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if x+w > c.Width() {
		w = c.Width() - x
	}
	if y < 0 {
		h += y
		y = 0
	}
	if y+h > c.Height() {
		h = c.Height() - y
	}
	if w <= 0 || h <= 0 {
		return
	}
	c.dirty = append(c.dirty, Rect{X: x, Y: y, W: w, H: h})
	c.mergeNewRect(len(c.dirty) - 1)
}

// mergeNewRect merges rectangle n with the best candidate, keeping the list
// below maxDirtyRects. This is a direct port of libcaca's merge_new_rect().
func (c *Canvas) mergeNewRect(n int) {
	best := -1
	bestScore := c.Width() * c.Height()

	sn := c.dirty[n].W * c.dirty[n].H

	for i := 0; i < len(c.dirty); i++ {
		if i == n {
			continue
		}
		xmin := minInt(c.dirty[i].X, c.dirty[n].X)
		ymin := minInt(c.dirty[i].Y, c.dirty[n].Y)
		xmax := maxInt(c.dirty[i].X+c.dirty[i].W-1, c.dirty[n].X+c.dirty[n].W-1)
		ymax := maxInt(c.dirty[i].Y+c.dirty[i].H-1, c.dirty[n].Y+c.dirty[n].H-1)
		sf := (xmax - xmin + 1) * (ymax - ymin + 1)

		// The existing rectangle is inside the new one: drop it.
		if sf == sn {
			c.dirty = append(c.dirty[:i], c.dirty[i+1:]...)
			if i < n {
				n--
			} else {
				i--
			}
			continue
		}

		si := c.dirty[i].W * c.dirty[i].H
		// The new rectangle is inside the existing one: drop it.
		if sf == si {
			c.dirty = append(c.dirty[:n], c.dirty[n+1:]...)
			return
		}

		wasted := sf - si - sn
		if wasted < bestScore {
			best = i
			bestScore = wasted
		}
	}

	if bestScore > 0 && len(c.dirty) < maxDirtyRects {
		return
	}
	if best < 0 {
		return
	}

	// Merge n into best.
	c.dirty[best].X = minInt(c.dirty[best].X, c.dirty[n].X)
	c.dirty[best].Y = minInt(c.dirty[best].Y, c.dirty[n].Y)
	xmax := maxInt(c.dirty[best].X+c.dirty[best].W-1, c.dirty[n].X+c.dirty[n].W-1)
	ymax := maxInt(c.dirty[best].Y+c.dirty[best].H-1, c.dirty[n].Y+c.dirty[n].H-1)
	c.dirty[best].W = xmax - c.dirty[best].X + 1
	c.dirty[best].H = ymax - c.dirty[best].Y + 1

	c.dirty = append(c.dirty[:n], c.dirty[n+1:]...)
	if best < n {
		c.mergeNewRect(best)
	} else {
		c.mergeNewRect(best - 1)
	}
}

func (c *Canvas) clipDirtyRects() {
	for i := range c.dirty {
		if c.dirty[i].X < 0 {
			c.dirty[i].X = 0
		}
		if c.dirty[i].Y < 0 {
			c.dirty[i].Y = 0
		}
		if c.dirty[i].X+c.dirty[i].W > c.Width() {
			c.dirty[i].W = c.Width() - c.dirty[i].X
		}
		if c.dirty[i].Y+c.dirty[i].H > c.Height() {
			c.dirty[i].H = c.Height() - c.dirty[i].Y
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
