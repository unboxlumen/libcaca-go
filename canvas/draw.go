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

type line struct {
	x1, y1, x2, y2 int
	ch             rune
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// DrawLine draws a solid line with ch.
func (c *Canvas) DrawLine(x1, y1, x2, y2 int, ch rune) {
	l := &line{x1: x1, y1: y1, x2: x2, y2: y2, ch: ch}
	c.clipLine(l, (*Canvas).drawSolidLine)
}

// DrawPolyline draws n line segments. x and y must have n+1 elements.
func (c *Canvas) DrawPolyline(x, y []int, n int, ch rune) {
	for i := 0; i < n; i++ {
		c.DrawLine(x[i], y[i], x[i+1], y[i+1], ch)
	}
}

// DrawThinLine draws an ASCII-art line.
func (c *Canvas) DrawThinLine(x1, y1, x2, y2 int) {
	l := &line{x1: x1, y1: y1, x2: x2, y2: y2}
	c.clipLine(l, (*Canvas).drawThinLine)
}

// DrawThinPolyline draws n ASCII-art line segments.
func (c *Canvas) DrawThinPolyline(x, y []int, n int) {
	for i := 0; i < n; i++ {
		c.DrawThinLine(x[i], y[i], x[i+1], y[i+1])
	}
}

func (c *Canvas) clipLine(l *line, draw func(*Canvas, *line)) {
	for {
		b1 := c.clipBits(l.x1, l.y1)
		b2 := c.clipBits(l.x2, l.y2)
		if b1&b2 != 0 {
			return
		}
		if b1 == 0 {
			if b2 == 0 {
				draw(c, l)
				return
			}
			l.x1, l.x2 = l.x2, l.x1
			l.y1, l.y2 = l.y2, l.y1
			continue
		}

		switch {
		case b1&(1<<0) != 0:
			den := l.x2 - l.x1
			if den == 0 {
				return
			}
			l.y1 = l.y2 - (l.x2-0)*(l.y2-l.y1)/den
			l.x1 = 0
		case b1&(1<<1) != 0:
			xmax := c.Width() - 1
			den := l.x2 - l.x1
			if den == 0 {
				return
			}
			l.y1 = l.y2 - (l.x2-xmax)*(l.y2-l.y1)/den
			l.x1 = xmax
		case b1&(1<<2) != 0:
			den := l.y2 - l.y1
			if den == 0 {
				return
			}
			l.x1 = l.x2 - (l.y2-0)*(l.x2-l.x1)/den
			l.y1 = 0
		case b1&(1<<3) != 0:
			ymax := c.Height() - 1
			den := l.y2 - l.y1
			if den == 0 {
				return
			}
			l.x1 = l.x2 - (l.y2-ymax)*(l.x2-l.x1)/den
			l.y1 = ymax
		default:
			return
		}
	}
}

func (c *Canvas) clipBits(x, y int) uint8 {
	var b uint8
	if x < 0 {
		b |= 1 << 0
	} else if x >= c.Width() {
		b |= 1 << 1
	}
	if y < 0 {
		b |= 1 << 2
	} else if y >= c.Height() {
		b |= 1 << 3
	}
	return b
}

func (c *Canvas) drawSolidLine(l *line) {
	x1, y1, x2, y2 := l.x1, l.y1, l.x2, l.y2
	dx, dy := absInt(x2-x1), absInt(y2-y1)
	xinc, yinc := 1, 1
	if x1 > x2 {
		xinc = -1
	}
	if y1 > y2 {
		yinc = -1
	}

	if dx >= dy {
		dpr := dy << 1
		dpru := dpr - (dx << 1)
		delta := dpr - dx
		for ; dx >= 0; dx-- {
			c.Put(x1, y1, l.ch)
			if delta > 0 {
				x1 += xinc
				y1 += yinc
				delta += dpru
			} else {
				x1 += xinc
				delta += dpr
			}
		}
	} else {
		dpr := dx << 1
		dpru := dpr - (dy << 1)
		delta := dpr - dy
		for ; dy >= 0; dy-- {
			c.Put(x1, y1, l.ch)
			if delta > 0 {
				x1 += xinc
				y1 += yinc
				delta += dpru
			} else {
				y1 += yinc
				delta += dpr
			}
		}
	}
}

func (c *Canvas) drawThinLine(s *line) {
	var mapx, mapy [2]rune
	var x1, y1, x2, y2 int
	yinc := 1

	if s.x2 >= s.x1 {
		if s.y1 > s.y2 {
			mapx = [2]rune{',', '\''}
		} else {
			mapx = [2]rune{'`', '.'}
		}
		x1, y1, x2, y2 = s.x1, s.y1, s.x2, s.y2
	} else {
		if s.y1 > s.y2 {
			mapx = [2]rune{'`', '.'}
		} else {
			mapx = [2]rune{',', '\''}
		}
		x2, y2, x1, y1 = s.x1, s.y1, s.x2, s.y2
	}

	dx, dy := absInt(x2-x1), absInt(y2-y1)
	if y1 > y2 {
		mapy = [2]rune{',', '\''}
		yinc = -1
	} else {
		mapy = [2]rune{'`', '.'}
	}

	if dx >= dy {
		dpr := dy << 1
		dpru := dpr - (dx << 1)
		delta := dpr - dx
		prev := 0
		for ; dx >= 0; dx-- {
			if delta > 0 {
				c.Put(x1, y1, mapy[1])
				x1++
				y1 += yinc
				delta += dpru
				prev = 1
			} else {
				if prev != 0 {
					c.Put(x1, y1, mapy[0])
				} else {
					c.Put(x1, y1, '-')
				}
				x1++
				delta += dpr
				prev = 0
			}
		}
	} else {
		dpr := dx << 1
		dpru := dpr - (dy << 1)
		delta := dpr - dy
		for ; dy >= 0; dy-- {
			if delta > 0 {
				c.Put(x1, y1, mapx[0])
				c.Put(x1+1, y1, mapx[1])
				x1++
				y1 += yinc
				delta += dpru
			} else {
				c.Put(x1, y1, '|')
				y1 += yinc
				delta += dpr
			}
		}
	}
}

// DrawBox draws a box outline using ch for every edge and corner.
func (c *Canvas) DrawBox(x, y, w, h int, ch rune) {
	x2, y2 := x+w-1, y+h-1
	c.DrawLine(x, y, x, y2, ch)
	c.DrawLine(x, y2, x2, y2, ch)
	c.DrawLine(x2, y2, x2, y, ch)
	c.DrawLine(x2, y, x, y, ch)
}

// DrawThinBox draws an ASCII-art box.
func (c *Canvas) DrawThinBox(x, y, w, h int) {
	c.drawBox(x, y, w, h, [6]rune{'-', '|', ',', '`', '.', '\''})
}

// DrawCP437Box draws a box using CP437 line-drawing characters.
func (c *Canvas) DrawCP437Box(x, y, w, h int) {
	c.drawBox(x, y, w, h, [6]rune{0x2500, 0x2502, 0x250c, 0x2514, 0x2510, 0x2518})
}

func (c *Canvas) drawBox(x, y, w, h int, chars [6]rune) {
	x2, y2 := x+w-1, y+h-1
	if x > x2 {
		x, x2 = x2, x
	}
	if y > y2 {
		y, y2 = y2, y
	}
	xmax, ymax := c.Width()-1, c.Height()-1
	if x2 < 0 || y2 < 0 || x > xmax || y > ymax {
		return
	}

	if y >= 0 {
		start := x + 1
		if x < 0 {
			start = 1
		}
		for i := start; i < x2 && i < xmax; i++ {
			c.Put(i, y, chars[0])
		}
	}
	if y2 <= ymax {
		start := x + 1
		if x < 0 {
			start = 1
		}
		for i := start; i < x2 && i < xmax; i++ {
			c.Put(i, y2, chars[0])
		}
	}
	if x >= 0 {
		start := y + 1
		if y < 0 {
			start = 1
		}
		for j := start; j < y2 && j < ymax; j++ {
			c.Put(x, j, chars[1])
		}
	}
	if x2 <= xmax {
		start := y + 1
		if y < 0 {
			start = 1
		}
		for j := start; j < y2 && j < ymax; j++ {
			c.Put(x2, j, chars[1])
		}
	}

	c.Put(x, y, chars[2])
	c.Put(x, y2, chars[3])
	c.Put(x2, y, chars[4])
	c.Put(x2, y2, chars[5])
}

// FillBox fills a rectangle with ch.
func (c *Canvas) FillBox(x, y, w, h int, ch rune) {
	x2, y2 := x+w-1, y+h-1
	if x > x2 {
		x, x2 = x2, x
	}
	if y > y2 {
		y, y2 = y2, y
	}
	xmax, ymax := c.Width()-1, c.Height()-1
	if x2 < 0 || y2 < 0 || x > xmax || y > ymax {
		return
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x2 > xmax {
		x2 = xmax
	}
	if y2 > ymax {
		y2 = ymax
	}
	for j := y; j <= y2; j++ {
		for i := x; i <= x2; i++ {
			c.Put(i, j, ch)
		}
	}
}

// DrawCircle draws a circle outline.
func (c *Canvas) DrawCircle(x, y, r int, ch rune) {
	test, dx, dy := 0, 0, r
	for ; dx <= dy; dx++ {
		c.ellipsePoints(x, y, dx, dy, ch, false)
		c.ellipsePoints(x, y, dy, dx, ch, false)
		if test > 0 {
			test += dx - dy
			dy--
		} else {
			test += dx
		}
	}
}

// DrawEllipse draws an ellipse outline.
func (c *Canvas) DrawEllipse(xo, yo, a, b int, ch rune) {
	x, y := 0, b
	d1 := b*b - a*a*b + a*a/4
	c.ellipsePoints(xo, yo, x, y, ch, false)
	for a*a*y-a*a/2 > b*b*(x+1) {
		if d1 < 0 {
			d1 += b * b * (2*x + 1)
		} else {
			d1 += b*b*(2*x) + a*a*(-2*y+2)
			y--
		}
		x++
		c.ellipsePoints(xo, yo, x, y, ch, false)
	}

	d2 := int(float64(b*b)*(float64(x)+0.5)*(float64(x)+0.5) + float64(a*a*(y-1)*(y-1)) - float64(a*a*b*b))
	for y > 0 {
		if d2 < 0 {
			d2 += b*b*(2*x+2) + a*a*(-2*y+3)
			x++
		} else {
			d2 += a * a * (-2*y + 3)
		}
		y--
		c.ellipsePoints(xo, yo, x, y, ch, false)
	}
}

// FillEllipse fills an ellipse.
func (c *Canvas) FillEllipse(xo, yo, a, b int, ch rune) {
	x, y := 0, b
	d1 := b*b - a*a*b + a*a/4
	for a*a*y-a*a/2 > b*b*(x+1) {
		if d1 < 0 {
			d1 += b * b * (2*x + 1)
		} else {
			d1 += b*b*(2*x) + a*a*(-2*y+2)
			c.DrawLine(xo-x, yo-y, xo+x, yo-y, ch)
			c.DrawLine(xo-x, yo+y, xo+x, yo+y, ch)
			y--
		}
		x++
	}
	c.DrawLine(xo-x, yo-y, xo+x, yo-y, ch)
	c.DrawLine(xo-x, yo+y, xo+x, yo+y, ch)

	d2 := int(float64(b*b)*float64(x) + 0.5*float64(x) + 0.5 + float64(a*a*(y-1)*(y-1)) - float64(a*a*b*b))
	for y > 0 {
		if d2 < 0 {
			d2 += b*b*(2*x+2) + a*a*(-2*y+3)
			x++
		} else {
			d2 += a * a * (-2*y + 3)
		}
		y--
		c.DrawLine(xo-x, yo-y, xo+x, yo-y, ch)
		c.DrawLine(xo-x, yo+y, xo+x, yo+y, ch)
	}
}

// DrawThinEllipse draws an ASCII-art ellipse.
func (c *Canvas) DrawThinEllipse(xo, yo, a, b int) {
	x, y := 0, b
	d1 := b*b - a*a*b + a*a/4
	c.ellipsePoints(xo, yo, x, y, '-', true)
	for a*a*y-a*a/2 > b*b*(x+1) {
		if d1 < 0 {
			d1 += b * b * (2*x + 1)
			c.ellipsePoints(xo, yo, x+1, y, '0', true)
		} else {
			d1 += b*b*(2*x) + a*a*(-2*y+2)
			y--
			c.ellipsePoints(xo, yo, x+1, y, '1', true)
		}
		x++
	}
	d2 := int(float64(b*b)*float64(x) + 0.5*float64(x) + 0.5 + float64(a*a*(y-1)*(y-1)) - float64(a*a*b*b))
	for y > 0 {
		if d2 < 0 {
			d2 += b*b*(2*x+2) + a*a*(-2*y+3)
			x++
			c.ellipsePoints(xo, yo, x, y-1, '2', true)
		} else {
			d2 += a * a * (-2*y + 3)
			c.ellipsePoints(xo, yo, x, y-1, '3', true)
		}
		y--
	}
}

func (c *Canvas) ellipsePoints(xo, yo, x, y int, ch rune, thin bool) {
	var b uint8
	if xo+x >= 0 && xo+x < c.Width() {
		b |= 0x1
	}
	if xo-x >= 0 && xo-x < c.Width() {
		b |= 0x2
	}
	if yo+y >= 0 && yo+y < c.Height() {
		b |= 0x4
	}
	if yo-y >= 0 && yo-y < c.Height() {
		b |= 0x8
	}

	if b&0x1 != 0 && b&0x4 != 0 {
		c.Put(xo+x, yo+y, ellipseThinChar(ch, thin, '-', ',', '/', '|'))
	}
	if b&0x2 != 0 && b&0x4 != 0 {
		c.Put(xo-x, yo+y, ellipseThinChar(ch, thin, '-', '.', '\\', '|'))
	}
	if b&0x1 != 0 && b&0x8 != 0 {
		c.Put(xo+x, yo-y, ellipseThinChar(ch, thin, '-', '`', '\\', '|'))
	}
	if b&0x2 != 0 && b&0x8 != 0 {
		c.Put(xo-x, yo-y, ellipseThinChar(ch, thin, '-', '\'', '/', '|'))
	}
}

func ellipseThinChar(ch rune, thin bool, zero, one, two, three rune) rune {
	if !thin {
		return ch
	}
	switch ch {
	case '0':
		return zero
	case '1':
		return one
	case '2':
		return two
	case '3':
		return three
	}
	return ch
}

// DrawTriangle draws a triangle outline.
func (c *Canvas) DrawTriangle(x1, y1, x2, y2, x3, y3 int, ch rune) {
	c.DrawLine(x1, y1, x2, y2, ch)
	c.DrawLine(x2, y2, x3, y3, ch)
	c.DrawLine(x3, y3, x1, y1, ch)
}

// DrawThinTriangle draws an ASCII-art triangle outline.
func (c *Canvas) DrawThinTriangle(x1, y1, x2, y2, x3, y3 int) {
	c.DrawThinLine(x1, y1, x2, y2)
	c.DrawThinLine(x2, y2, x3, y3)
	c.DrawThinLine(x3, y3, x1, y1)
}

// FillTriangle fills a triangle.
func (c *Canvas) FillTriangle(x1, y1, x2, y2, x3, y3 int, ch rune) {
	if y1 > y2 {
		c.FillTriangle(x2, y2, x1, y1, x3, y3, ch)
		return
	}
	if y2 > y3 {
		c.FillTriangle(x1, y1, x3, y3, x2, y2, ch)
		return
	}

	sl21, sl31, sl32 := 0, 0, 0
	if y2 != y1 {
		sl21 = (x2 - x1) * 0x10000 / (y2 - y1)
	}
	if y3 != y1 {
		sl31 = (x3 - x1) * 0x10000 / (y3 - y1)
	}
	if y3 != y2 {
		sl32 = (x3 - x2) * 0x10000 / (y3 - y2)
	}

	x1 *= 0x10000
	x2 *= 0x10000
	x3 *= 0x10000

	ymin := y1
	if ymin < 0 {
		ymin = 0
	}
	ymax := y3 + 1
	if ymax > c.Height() {
		ymax = c.Height()
	}

	var xa, xb int
	switch {
	case ymin < y2:
		xa = x1 + sl21*(ymin-y1)
		xb = x1 + sl31*(ymin-y1)
	case ymin == y2:
		xa = x2
		if y1 == y3 {
			xb = x3
		} else {
			xb = x1 + sl31*(ymin-y1)
		}
	default:
		xa = x3 + sl32*(ymin-y3)
		xb = x3 + sl31*(ymin-y3)
	}

	for y := ymin; y < ymax; y++ {
		var xx1, xx2 int
		if xa < xb {
			xx1 = (xa + 0x800) / 0x10000
			xx2 = (xb + 0x801) / 0x10000
		} else {
			xx1 = (xb + 0x800) / 0x10000
			xx2 = (xa + 0x801) / 0x10000
		}

		xmin := xx1
		if xmin < 0 {
			xmin = 0
		}
		xmax := xx2 + 1
		if xmax > c.Width() {
			xmax = c.Width()
		}
		for x := xmin; x < xmax; x++ {
			c.Put(x, y, ch)
		}

		if y < y2 {
			xa += sl21
		} else {
			xa += sl32
		}
		xb += sl31
	}
}
