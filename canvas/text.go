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
	"unicode/utf8"

	"github.com/unboxlumen/libcaca-go/charset"
)

// MagicFullwidth marks the right half of a fullwidth character. It matches
// libcaca's CACA_MAGIC_FULLWIDTH.
const MagicFullwidth rune = 0x0FFFFE

// Put writes one character using the current attribute and returns the
// number of cells it occupies (1 or 2). Coordinates outside the canvas are
// ignored, but the return value still reports the character width.
func (c *Canvas) Put(x, y int, ch rune) int {
	if ch == MagicFullwidth {
		return 1
	}

	f := c.frame()
	fullwidth := charset.IsFullwidth(ch)
	ret := 1
	if fullwidth {
		ret = 2
	}

	if x >= f.w || y < 0 || y >= f.h {
		return ret
	}
	if x == -1 && fullwidth {
		x = 0
		ch = ' '
		fullwidth = false
	} else if x < 0 {
		return ret
	}

	idx := y*f.w + x
	attr := f.curAttr
	xmin, xmax := x, x

	// Overwriting the right half of a fullwidth character: blank its left half.
	if x != 0 && f.cells[idx].Ch == MagicFullwidth {
		f.cells[idx-1].Ch = ' '
		xmin--
	}

	if fullwidth {
		if x+1 == f.w {
			ch = ' '
		} else {
			xmax++
			// Overwriting the left half of a fullwidth character: blank its right half.
			if x+2 < f.w && f.cells[idx+2].Ch == MagicFullwidth {
				f.cells[idx+2].Ch = ' '
				xmax++
			}
			f.cells[idx+1].Ch = MagicFullwidth
			f.cells[idx+1].A = attr
		}
	} else {
		if x+1 != f.w && f.cells[idx+1].Ch == MagicFullwidth {
			f.cells[idx+1].Ch = ' '
			xmax++
		}
	}

	if f.cells[idx].Ch != ch || f.cells[idx].A != attr {
		c.addDirtyRect(xmin, y, xmax-xmin+1, 1)
	}
	f.cells[idx] = Cell{Ch: ch, A: attr}
	return ret
}

// Get returns the character at (x,y), or a space outside the canvas.
func (c *Canvas) Get(x, y int) rune {
	f := c.frame()
	if x < 0 || x >= f.w || y < 0 || y >= f.h {
		return ' '
	}
	return f.cells[y*f.w+x].Ch
}

// GetAttr returns the attribute at (x,y), or the current attribute outside
// the canvas.
func (c *Canvas) GetAttr(x, y int) Attr {
	f := c.frame()
	if x < 0 || x >= f.w || y < 0 || y >= f.h {
		return f.curAttr
	}
	return f.cells[y*f.w+x].A
}

// PutAttr sets the attribute at (x,y) without changing the character. If the
// cell is part of a fullwidth character, both halves get the attribute.
// Values below 0x10 only modify the style bits.
func (c *Canvas) PutAttr(x, y int, attr Attr) {
	f := c.frame()
	if x < 0 || x >= f.w || y < 0 || y >= f.h {
		return
	}
	idx := y*f.w + x
	if attr < 0x00000010 {
		f.cells[idx].A = (f.cells[idx].A &^ styleMask) | attr
	} else {
		f.cells[idx].A = attr
	}

	xmin, xmax := x, x
	if x != 0 && f.cells[idx].Ch == MagicFullwidth {
		f.cells[idx-1].A = f.cells[idx].A
		xmin--
	} else if x+1 < f.w && f.cells[idx+1].Ch == MagicFullwidth {
		f.cells[idx+1].A = f.cells[idx].A
		xmax++
	}
	c.addDirtyRect(xmin, y, xmax-xmin+1, 1)
}

// PutString writes a UTF-8 string at (x,y) and returns the number of cells
// used.
func (c *Canvas) PutString(x, y int, s string) int {
	f := c.frame()
	length := 0

	if y < 0 || y >= f.h || x >= f.w {
		for len(s) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if charset.IsFullwidth(r) {
				length += 2
			} else {
				length++
			}
			if size == 0 {
				size = 1
			}
			s = s[size:]
		}
		return length
	}

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if x+length >= -1 && x+length < f.w {
			c.Put(x+length, y, r)
		}
		if charset.IsFullwidth(r) {
			length += 2
		} else {
			length++
		}
		if size == 0 {
			size = 1
		}
		s = s[size:]
	}
	return length
}

// Printf formats and writes a string at (x,y).
func (c *Canvas) Printf(x, y int, format string, args ...any) int {
	return c.PutString(x, y, fmt.Sprintf(format, args...))
}

// Blit copies src onto c at (x,y). If mask is non-nil it must have the same
// dimensions as src; mask cells containing a space are skipped.
func (c *Canvas) Blit(x, y int, src *Canvas, mask *Canvas) error {
	dst := c.frame()
	s := src.frames[src.cur]

	if mask != nil {
		m := mask.frames[mask.cur]
		if s.w != m.w || s.h != m.h {
			return ErrInvalidSize
		}
	}

	x -= s.handleX
	y -= s.handleY

	starti := 0
	if x < 0 {
		starti = -x
	}
	startj := 0
	if y < 0 {
		startj = -y
	}
	endi := s.w
	if x+s.w >= dst.w {
		endi = dst.w - x
	}
	endj := s.h
	if y+s.h >= dst.h {
		endj = dst.h - y
	}
	stride := endi - starti
	if starti > s.w || startj > s.h || starti >= endi || startj >= endj {
		return nil
	}

	var m *Frame
	if mask != nil {
		m = mask.frames[mask.cur]
	}

	for j := startj; j < endj; j++ {
		dstix := (j+y)*dst.w + starti + x
		srcix := j*s.w + starti

		if (starti+x) != 0 && dst.cells[dstix].Ch == MagicFullwidth {
			dst.cells[dstix-1].Ch = ' '
		}
		if endi+x < dst.w && dst.cells[dstix+stride].Ch == MagicFullwidth {
			dst.cells[dstix+stride].Ch = ' '
		}

		if m != nil {
			for i := 0; i < stride; i++ {
				if m.cells[srcix+i].Ch == ' ' {
					continue
				}
				if dst.cells[dstix+i].Ch != s.cells[srcix+i].Ch ||
					dst.cells[dstix+i].A != s.cells[srcix+i].A {
					dst.cells[dstix+i] = s.cells[srcix+i]
					c.addDirtyRect(x+starti+i, y+j, 1, 1)
				}
			}
		} else {
			changed := false
			for i := 0; i < stride; i++ {
				if dst.cells[dstix+i] != s.cells[srcix+i] {
					changed = true
					break
				}
			}
			if changed {
				copy(dst.cells[dstix:dstix+stride], s.cells[srcix:srcix+stride])
				c.addDirtyRect(x+starti, y+j, stride, 1)
			}
		}

		// Fix split fullwidth characters.
		if s.cells[srcix].Ch == MagicFullwidth {
			dst.cells[dstix].Ch = ' '
		}
		// This mirrors libcaca's row-0 lookup; see caca_blit().
		if endi < s.w && s.cells[endi].Ch == MagicFullwidth {
			dst.cells[dstix+stride-1].Ch = ' '
		}
	}

	return nil
}
