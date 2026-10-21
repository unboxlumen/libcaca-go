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

// invert flips the ANSI colour indices of every cell while keeping the
// characters. This matches caca_invert().
func (c *Canvas) Invert() {
	f := c.frame()
	for i := range f.cells {
		f.cells[i].A ^= 0x000f000f
	}
	c.addDirtyRect(0, 0, f.w, f.h)
}

// Flip mirrors the canvas horizontally, choosing characters that look like
// the mirrored version where possible.
func (c *Canvas) Flip() {
	f := c.frame()
	for y := 0; y < f.h; y++ {
		left := y * f.w
		right := left + f.w - 1
		for left < right {
			f.cells[left], f.cells[right] = f.cells[right], f.cells[left]
			f.cells[left].Ch = flipChar(f.cells[left].Ch)
			f.cells[right].Ch = flipChar(f.cells[right].Ch)
			left++
			right--
		}
		if left == right {
			f.cells[left].Ch = flipChar(f.cells[left].Ch)
		}

		// Move the fullwidth sentinel to the right half.
		for x := 0; x < f.w-1; x++ {
			if f.cells[y*f.w+x].Ch == MagicFullwidth {
				f.cells[y*f.w+x].Ch = f.cells[y*f.w+x+1].Ch
				f.cells[y*f.w+x+1].Ch = MagicFullwidth
				x++
			}
		}
	}
	c.addDirtyRect(0, 0, f.w, f.h)
}

// Flop mirrors the canvas vertically.
func (c *Canvas) Flop() {
	f := c.frame()
	for x := 0; x < f.w; x++ {
		top := x
		bottom := x + f.w*(f.h-1)
		for top < bottom {
			f.cells[top], f.cells[bottom] = f.cells[bottom], f.cells[top]
			f.cells[top].Ch = flopChar(f.cells[top].Ch)
			f.cells[bottom].Ch = flopChar(f.cells[bottom].Ch)
			top += f.w
			bottom -= f.w
		}
		if top == bottom {
			f.cells[top].Ch = flopChar(f.cells[top].Ch)
		}
	}
	c.addDirtyRect(0, 0, f.w, f.h)
}

// Rotate180 rotates the canvas by 180 degrees.
func (c *Canvas) Rotate180() {
	f := c.frame()
	begin, end := 0, len(f.cells)-1
	for begin < end {
		f.cells[begin], f.cells[end] = f.cells[end], f.cells[begin]
		f.cells[begin].Ch = rotateChar(f.cells[begin].Ch)
		f.cells[end].Ch = rotateChar(f.cells[end].Ch)
		begin++
		end--
	}
	if begin == end {
		f.cells[begin].Ch = rotateChar(f.cells[begin].Ch)
	}
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w-1; x++ {
			if f.cells[y*f.w+x].Ch == MagicFullwidth {
				f.cells[y*f.w+x].Ch = f.cells[y*f.w+x+1].Ch
				f.cells[y*f.w+x+1].Ch = MagicFullwidth
				x++
			}
		}
	}
	c.addDirtyRect(0, 0, f.w, f.h)
}

// RotateLeft rotates the canvas 90 degrees counterclockwise. The width is
// halved (rounded up) and becomes the new height; the height doubles and
// becomes the new width.
func (c *Canvas) RotateLeft() {
	f := c.frame()
	oldW, oldH := f.w, f.h
	w2 := (oldW + 1) / 2
	h2 := oldH
	nw, nh := oldH*2, w2

	cells := make([]Cell, nw*nh)
	for y := 0; y < h2; y++ {
		for x := 0; x < w2; x++ {
			i := oldW*y + x*2
			pair := [2]rune{f.cells[i].Ch, ' '}
			a1, a2 := f.cells[i].A, f.cells[i].A
			if !(oldW&1 != 0 && x == w2-1) {
				pair[1] = f.cells[i+1].Ch
				a2 = f.cells[i+1].A
			}
			if pair[0] == ' ' {
				a1 = a2
			} else if pair[1] == ' ' {
				a2 = a1
			}
			pair = leftPair(pair)

			j := (w2-1-x)*nw + y*2
			cells[j] = Cell{Ch: pair[0], A: a1}
			cells[j+1] = Cell{Ch: pair[1], A: a2}
		}
	}

	cx, cy := f.x, f.y
	hx, hy := f.handleX, f.handleY
	f.x, f.y = cy*2, (oldW-1-cx)/2
	f.handleX, f.handleY = hy*2, (oldW-1-hx)/2
	f.w, f.h, f.cells = nw, nh, cells
	c.addDirtyRect(0, 0, nw, nh)
}

// RotateRight rotates the canvas 90 degrees clockwise. The width is halved
// (rounded up) and becomes the new height; the height doubles and becomes
// the new width.
func (c *Canvas) RotateRight() {
	f := c.frame()
	oldW, oldH := f.w, f.h
	w2 := (oldW + 1) / 2
	h2 := oldH
	nw, nh := oldH*2, w2

	cells := make([]Cell, nw*nh)
	for y := 0; y < h2; y++ {
		for x := 0; x < w2; x++ {
			i := oldW*y + x*2
			pair := [2]rune{f.cells[i].Ch, ' '}
			a1, a2 := f.cells[i].A, f.cells[i].A
			if !(oldW&1 != 0 && x == w2-1) {
				pair[1] = f.cells[i+1].Ch
				a2 = f.cells[i+1].A
			}
			if pair[0] == ' ' {
				a1 = a2
			} else if pair[1] == ' ' {
				a2 = a1
			}
			pair = rightPair(pair)

			j := x*nw + (h2-1-y)*2
			cells[j] = Cell{Ch: pair[0], A: a1}
			cells[j+1] = Cell{Ch: pair[1], A: a2}
		}
	}

	cx, cy := f.x, f.y
	hx, hy := f.handleX, f.handleY
	f.x, f.y = (oldH-1-cy)*2, cx/2
	f.handleX, f.handleY = (oldH-1-hy)*2, hx/2
	f.w, f.h, f.cells = nw, nh, cells
	c.addDirtyRect(0, 0, nw, nh)
}

// StretchLeft rotates the canvas 90 degrees counterclockwise without the
// half-width pairing; the aspect ratio becomes stretched.
func (c *Canvas) StretchLeft() {
	f := c.frame()
	oldW, oldH := f.w, f.h
	cells := make([]Cell, oldW*oldH)
	for y := 0; y < oldH; y++ {
		for x := 0; x < oldW; x++ {
			cell := f.cells[oldW*y+x]
			cell.Ch = leftChar(cell.Ch)
			cells[(oldW-1-x)*oldH+y] = cell
		}
	}

	cx, cy := f.x, f.y
	hx, hy := f.handleX, f.handleY
	f.x, f.y = cy, oldW-1-cx
	f.handleX, f.handleY = hy, oldW-1-hx
	f.w, f.h, f.cells = oldH, oldW, cells
	c.addDirtyRect(0, 0, f.w, f.h)
}

// StretchRight rotates the canvas 90 degrees clockwise without pairing.
func (c *Canvas) StretchRight() {
	f := c.frame()
	oldW, oldH := f.w, f.h
	cells := make([]Cell, oldW*oldH)
	for y := 0; y < oldH; y++ {
		for x := 0; x < oldW; x++ {
			cell := f.cells[oldW*y+x]
			cell.Ch = rightChar(cell.Ch)
			cells[x*oldH+(oldH-1-y)] = cell
		}
	}

	cx, cy := f.x, f.y
	hx, hy := f.handleX, f.handleY
	f.x, f.y = oldH-1-cy, cx
	f.handleX, f.handleY = oldH-1-hy, hx
	f.w, f.h, f.cells = oldH, oldW, cells
	c.addDirtyRect(0, 0, f.w, f.h)
}

// ---- character mappings -------------------------------------------------

func lookupFlip(ch rune, no, pairs []rune) rune {
	for _, v := range no {
		if v == 0 {
			break
		}
		if ch == v {
			return ch
		}
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == 0 {
			break
		}
		if ch == pairs[i] {
			return pairs[i+1]
		}
		if ch == pairs[i+1] {
			return pairs[i]
		}
	}
	return ch
}

func flipChar(ch rune) rune   { return lookupFlip(ch, flipNoflip[:], flipPairs[:]) }
func flopChar(ch rune) rune   { return lookupFlip(ch, flopNoflop[:], flopPairs[:]) }
func rotateChar(ch rune) rune { return lookupFlip(ch, rotateNorotate[:], rotatePairs[:]) }

func leftChar(ch rune) rune {
	for i := 0; i+1 < len(leftright2); i += 2 {
		if leftright2[i] == 0 {
			break
		}
		if ch == leftright2[i] {
			return leftright2[(i&^1)|((i+1)&1)]
		}
	}
	for i := 0; i < len(leftright4); i++ {
		if leftright4[i] == 0 {
			break
		}
		if ch == leftright4[i] {
			return leftright4[(i&^3)|((i+1)&3)]
		}
	}
	return ch
}

func rightChar(ch rune) rune {
	for i := 0; i+1 < len(leftright2); i += 2 {
		if leftright2[i] == 0 {
			break
		}
		if ch == leftright2[i] {
			return leftright2[(i&^1)|((i-1)&1)]
		}
	}
	for i := 0; i < len(leftright4); i++ {
		if leftright4[i] == 0 {
			break
		}
		if ch == leftright4[i] {
			return leftright4[(i&^3)|((i-1)&3)]
		}
	}
	return ch
}

func leftPair(p [2]rune) [2]rune {
	for i := 0; i+1 < len(leftright2x2); i += 2 {
		if leftright2x2[i] == 0 {
			break
		}
		if p[0] == leftright2x2[i] && p[1] == leftright2x2[i+1] {
			j := (i &^ 3) | ((i + 2) & 3)
			return [2]rune{leftright2x2[j], leftright2x2[j+1]}
		}
	}
	for i := 0; i+1 < len(leftright2x4); i += 2 {
		if leftright2x4[i] == 0 && leftright2x4[i+1] == 0 {
			break
		}
		if p[0] == leftright2x4[i] && p[1] == leftright2x4[i+1] {
			j := (i &^ 7) | ((i + 2) & 7)
			return [2]rune{leftright2x4[j], leftright2x4[j+1]}
		}
	}
	return p
}

func rightPair(p [2]rune) [2]rune {
	for i := 0; i+1 < len(leftright2x2); i += 2 {
		if leftright2x2[i] == 0 {
			break
		}
		if p[0] == leftright2x2[i] && p[1] == leftright2x2[i+1] {
			j := (i &^ 3) | ((i - 2) & 3)
			return [2]rune{leftright2x2[j], leftright2x2[j+1]}
		}
	}
	for i := 0; i+1 < len(leftright2x4); i += 2 {
		if leftright2x4[i] == 0 && leftright2x4[i+1] == 0 {
			break
		}
		if p[0] == leftright2x4[i] && p[1] == leftright2x4[i+1] {
			j := (i &^ 7) | ((i - 2) & 7)
			return [2]rune{leftright2x4[j], leftright2x4[j+1]}
		}
	}
	return p
}
