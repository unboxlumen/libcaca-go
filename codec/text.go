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

package codec

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/charset"
)

var ansiPalette = [16]byte{0, 4, 2, 6, 1, 5, 3, 7, 8, 12, 10, 14, 9, 13, 11, 15}

var ircPalette = [16]byte{1, 2, 3, 10, 5, 6, 7, 15, 14, 12, 9, 11, 4, 13, 8, 0}

func exportANSI(v view) []byte {
	var b bytes.Buffer
	prevfg, prevbg := -1, -1

	for y := 0; y < v.h; y++ {
		for x := 0; x < v.w; x++ {
			cell := v.cells[y*v.w+x]
			afg, abg := cell.A.FG(), cell.A.BG()
			fg, bg := byte(canvas.LightGray), byte(canvas.Black)
			if afg < 0x10 {
				fg = ansiPalette[afg]
			}
			if abg < 0x10 {
				bg = ansiPalette[abg]
			}
			ch := cell.Ch
			if ch == canvas.MagicFullwidth {
				ch = '?'
			}

			if int(fg) != prevfg || int(bg) != prevbg {
				b.WriteString("\x1b[0;")
				if fg < 8 {
					if bg < 8 {
						fmt.Fprintf(&b, "3%d;4%dm", fg, bg)
					} else {
						fmt.Fprintf(&b, "5;3%d;4%dm", fg, bg-8)
					}
				} else {
					if bg < 8 {
						fmt.Fprintf(&b, "1;3%d;4%dm", fg-8, bg)
					} else {
						fmt.Fprintf(&b, "5;1;3%d;4%dm", fg-8, bg-8)
					}
				}
			}
			b.WriteByte(charset.ToCP437(ch))
			prevfg, prevbg = int(fg), int(bg)
		}

		if v.w == 80 {
			b.WriteString("\x1b[s\n\x1b[u")
		} else {
			b.WriteString("\x1b[0m\r\n")
			prevfg, prevbg = -1, -1
		}
	}
	return b.Bytes()
}

func exportUTF8(v view, cr bool) []byte {
	var b bytes.Buffer
	for y := 0; y < v.h; y++ {
		prevfg, prevbg := byte(0x10), byte(0x10)
		for x := 0; x < v.w; x++ {
			cell := v.cells[y*v.w+x]
			if cell.Ch == canvas.MagicFullwidth {
				continue
			}
			afg, abg := cell.A.FG(), cell.A.BG()
			fg, bg := byte(0x10), byte(0x10)
			if afg < 0x10 {
				fg = ansiPalette[afg]
			}
			if abg < 0x10 {
				bg = ansiPalette[abg]
			}

			if fg != prevfg || bg != prevbg {
				b.WriteString("\x1b[0")
				if fg < 8 {
					fmt.Fprintf(&b, ";3%d", fg)
				} else if fg < 16 {
					fmt.Fprintf(&b, ";1;3%d;9%d", fg-8, fg-8)
				}
				if bg < 8 {
					fmt.Fprintf(&b, ";4%d", bg)
				} else if bg < 16 {
					fmt.Fprintf(&b, ";5;4%d;10%d", bg-8, bg-8)
				}
				b.WriteByte('m')
			}
			b.WriteRune(cell.Ch)
			prevfg, prevbg = fg, bg
		}
		if prevfg != 0x10 || prevbg != 0x10 {
			b.WriteString("\x1b[0m")
		}
		if cr {
			b.WriteString("\r\n")
		} else {
			b.WriteByte('\n')
		}
	}
	return b.Bytes()
}

func exportIRC(v view) []byte {
	var b bytes.Buffer
	for y := 0; y < v.h; y++ {
		prevfg, prevbg := byte(0x10), byte(0x10)
		for x := 0; x < v.w; x++ {
			cell := v.cells[y*v.w+x]
			if cell.Ch == canvas.MagicFullwidth {
				continue
			}
			afg, abg := cell.A.FG(), cell.A.BG()
			fg, bg := byte(0x10), byte(0x10)
			if afg < 0x10 {
				fg = ircPalette[afg]
			}
			if abg < 0x10 {
				bg = ircPalette[abg]
			}

			if bg != prevbg || fg != prevfg {
				needEscape := false
				if bg == 0x10 {
					if fg == 0x10 {
						b.WriteByte(0x0f)
					} else {
						if prevbg == 0x10 {
							fmt.Fprintf(&b, "\x03%d", fg)
						} else {
							fmt.Fprintf(&b, "\x0f\x03%d", fg)
						}
						if cell.Ch == ',' {
							needEscape = true
						}
					}
				} else {
					if fg == 0x10 {
						fmt.Fprintf(&b, "\x0f\x03,%d", bg)
					} else {
						fmt.Fprintf(&b, "\x03%d,%d", fg, bg)
					}
				}
				if cell.Ch >= '0' && cell.Ch <= '9' {
					needEscape = true
				}
				if needEscape {
					b.WriteString("\x02\x02")
				}
			}
			b.WriteRune(cell.Ch)
			prevfg, prevbg = fg, bg
		}
		if v.w == 0 {
			b.WriteByte(' ')
		}
		b.WriteString("\r\n")
	}
	return b.Bytes()
}

// ---- plain text and BIN import ------------------------------------------

func importText(c *canvas.Canvas, data []byte) (int, error) {
	if err := c.Resize(0, 0); err != nil {
		return 0, err
	}
	width, height := 0, 0
	x, y := 0, 0
	for i := 0; i < len(data); i++ {
		ch := data[i]
		switch ch {
		case '\r':
			continue
		case '\n':
			x = 0
			y++
			continue
		}
		if x >= width || y >= height {
			if x >= width {
				width = x + 1
			}
			if y >= height {
				height = y + 1
			}
			if err := c.Resize(width, height); err != nil {
				return 0, err
			}
		}
		c.Put(x, y, rune(ch))
		x++
	}
	if y > height {
		if err := c.Resize(width, y); err != nil {
			return 0, err
		}
	}
	return len(data), nil
}

func importBIN(c *canvas.Canvas, data []byte) (int, error) {
	if err := c.Resize(0, 0); err != nil {
		return 0, err
	}
	if err := c.Resize(160, len(data)/160); err != nil {
		return 0, err
	}
	n := len(data) &^ 1
	x, y := 0, 0
	for i := 0; i+1 < n; i += 2 {
		_ = c.SetColor(canvas.Color(data[i+1]&0xf), canvas.Color(data[i+1]>>4))
		c.Put(x, y, charset.FromCP437(data[i]))
		x++
		if x >= 160 {
			x = 0
			y++
		}
	}
	return n, nil
}

// ---- ANSI / UTF-8 import ------------------------------------------------

type ansiImport struct {
	fg, bg        canvas.Color
	dfg, dbg      canvas.Color
	bold, blink   bool
	italics       bool
	negative      bool
	concealed     bool
	underline     bool
	faint, strike bool
	proportional  bool
	clearAttr     canvas.Attr
	saveX, saveY  int
}

func importANSI(c *canvas.Canvas, data []byte, isUTF8 bool) (int, error) {
	im := &ansiImport{}
	var width, height int
	var growx, growy bool
	x, y := 0, 0

	if isUTF8 {
		width, height = c.Width(), c.Height()
		growx = width == 0
		growy = height == 0
		x, y = c.WhereX(), c.WhereY()
		im.dfg, im.dbg = canvas.Default, canvas.Transparent
	} else {
		if err := c.Resize(80, 0); err != nil {
			return 0, err
		}
		width, height = 80, 0
		growy = true
		im.dfg, im.dbg = canvas.LightGray, canvas.Black
	}
	im.fg, im.bg = im.dfg, im.dbg
	if err := c.SetColor(im.fg, im.bg); err != nil {
		return 0, err
	}
	im.clearAttr = c.CurAttr()
	ansiParseGRCM(c, im, []int{0})

	i := 0
	for i < len(data) {
		ch := data[i]
		switch {
		case !isUTF8 && ch == 0x1a && i+7 < len(data) && string(data[i+1:i+8]) == "SAUCE00":
			c.GotoXY(x, y)
			return i, nil

		case ch == '\r':
			x = 0
			i++

		case ch == '\n':
			x = 0
			y++
			i++

		case ch == '\t':
			x = (x + 8) &^ 7
			i++

		case ch == '\b':
			if x > 0 {
				x--
			}
			i++

		case ch == 0x0c && i+1 < len(data) && data[i+1] == '\n':
			n := c.FrameCount()
			if err := c.CreateFrame(n); err != nil {
				return 0, err
			}
			if err := c.SetFrame(n); err != nil {
				return 0, err
			}
			x, y = 0, 0
			i += 2

		case ch == 0x1b && i+2 >= len(data):
			c.GotoXY(x, y)
			return i, nil

		case ch == 0x1b && data[i+1] == '(' && data[i+2] == 'B':
			i += 3

		case ch == 0x1b && data[i+1] == '[':
			c.GotoXY(x, y)
			n, nw, nh, stopped := importCSI(c, im, data, i)
			if stopped {
				c.GotoXY(x, y)
				return i, nil
			}
			// CSI handlers may change the cursor or the canvas size.
			width, height = c.Width(), c.Height()
			x, y = c.WhereX(), c.WhereY()
			if n > 0 {
				i = n
			} else {
				i++
			}
			_ = nw
			_ = nh

		case ch == 0x1b && i+1 < len(data) && data[i+1] == ']':
			// OSC: skip to BEL, or stop if incomplete.
			j := i + 2
			for j < len(data) && data[j] != 0x07 {
				j++
			}
			if j >= len(data) {
				c.GotoXY(x, y)
				return i, nil
			}
			i = j + 1

		case ch == 0x1b:
			i += 2

		default:
			var r rune
			var size int
			wch := 1
			if isUTF8 {
				r, size = utf8.DecodeRune(data[i:])
				if r == utf8.RuneError && size == 1 {
					r = rune(data[i])
					size = 1
				}
				if charset.IsFullwidth(r) {
					wch = 2
				}
				i += size
			} else {
				r = charset.FromCP437(data[i])
				i++
			}

			for x+wch > width {
				if growx {
					saved := c.CurAttr()
					c.SetAttr(im.clearAttr)
					if err := c.Resize(x+wch, height); err != nil {
						return 0, err
					}
					width = x + wch
					c.SetAttr(saved)
				} else {
					x -= width
					y++
				}
			}
			if y >= height {
				saved := c.CurAttr()
				c.SetAttr(im.clearAttr)
				if growy {
					if err := c.Resize(width, y+1); err != nil {
						return 0, err
					}
					height = y + 1
				} else {
					lines := y - height + 1
					for j := 0; j+lines < height; j++ {
						for k := 0; k < width; k++ {
							c.SetCell(k, j, c.CellAt(k, j+lines))
						}
					}
					for j := height - lines; j < height; j++ {
						for k := 0; k < width; k++ {
							c.SetCell(k, j, canvas.Cell{Ch: ' ', A: im.clearAttr})
						}
					}
					y -= lines
				}
				c.SetAttr(saved)
			}
			if wch != 0 {
				c.Put(x, y, r)
				x += wch
			}
		}
	}

	if growy && y > height {
		saved := c.CurAttr()
		c.SetAttr(im.clearAttr)
		if err := c.Resize(width, y); err != nil {
			return 0, err
		}
		c.SetAttr(saved)
	}
	c.GotoXY(x, y)
	return i, nil
}

// importCSI handles one ESC [ sequence starting at i. It returns the next
// index, the new width/height from the caller's perspective (unused), and
// whether parsing must stop because the sequence is incomplete.
func importCSI(c *canvas.Canvas, im *ansiImport, data []byte, i int) (next, w, h int, stop bool) {
	param := i + 2
	inter := param
	for inter < len(data) && data[inter] >= 0x30 && data[inter] <= 0x3f {
		inter++
	}
	final := inter
	for final < len(data) && data[final] >= 0x20 && data[final] <= 0x2f {
		final++
	}
	if final >= len(data) || data[final] < 0x40 || data[final] > 0x7e {
		return 0, 0, 0, true
	}
	next = final + 1

	var argv []int
	if param < inter {
		if data[param] >= 0x3c {
			return next, 0, 0, false // private sequence
		}
		cur := 0
		started := false
		for j := param; j < inter; j++ {
			if data[j] == ';' || data[j] == ':' {
				argv = append(argv, cur)
				cur = 0
				started = true
			} else if data[j] >= '0' && data[j] <= '9' {
				cur = cur*10 + int(data[j]-'0')
				started = true
			}
		}
		if started {
			argv = append(argv, cur)
		}
	}

	arg := func(n, def int) int {
		if n < len(argv) {
			return argv[n]
		}
		return def
	}
	cx, cy := c.WhereX(), c.WhereY()

	switch data[final] {
	case 'H', 'f':
		cx = arg(1, 1) - 1
		cy = arg(0, 1) - 1
		if cx < 0 {
			cx = 0
		}
		if cy < 0 {
			cy = 0
		}
	case 'A':
		cy -= max(1, arg(0, 1))
		if cy < 0 {
			cy = 0
		}
	case 'B':
		cy += max(1, arg(0, 1))
	case 'C':
		cx += max(1, arg(0, 1))
	case 'D':
		cx -= max(1, arg(0, 1))
		if cx < 0 {
			cx = 0
		}
	case 'G':
		cx = arg(0, 1) - 1
		if cx < 0 {
			cx = 0
		}
	case 'd':
		cy = arg(0, 1) - 1
		if cy < 0 {
			cy = 0
		}
	case 'J':
		saved := c.CurAttr()
		c.SetAttr(im.clearAttr)
		switch arg(0, 0) {
		case 0:
			c.DrawLine(cx, cy, c.Width(), cy, ' ')
			c.FillBox(0, cy+1, c.Width(), c.Height()-cy-1, ' ')
		case 1:
			c.FillBox(0, 0, c.Width(), cy, ' ')
			c.DrawLine(0, cy, cx, cy, ' ')
		case 2:
			c.FillBox(0, 0, c.Width(), c.Height(), ' ')
		}
		c.SetAttr(saved)
	case 'K':
		switch arg(0, 0) {
		case 0:
			c.DrawLine(cx, cy, c.Width(), cy, ' ')
		case 1:
			c.DrawLine(0, cy, cx, cy, ' ')
		case 2:
			c.DrawLine(cx, cy, c.Width()-1, cy, ' ')
		}
	case 'P':
		n := max(1, arg(0, 1))
		for j := 0; j+n < c.Width(); j++ {
			c.SetCell(j, cy, c.CellAt(j+n, cy))
		}
	case 'X':
		n := arg(0, 0)
		if n > 0 {
			saved := c.CurAttr()
			c.SetAttr(im.clearAttr)
			c.DrawLine(cx, cy, cx+n-1, cy, ' ')
			c.SetAttr(saved)
		}
	case 'm':
		ansiParseGRCM(c, im, argv)
	case 's':
		im.saveX, im.saveY = cx, cy
	case 'u':
		cx, cy = im.saveX, im.saveY
	}
	c.GotoXY(cx, cy)
	return next, c.Width(), c.Height(), false
}

func ansiParseGRCM(c *canvas.Canvas, im *ansiImport, argv []int) {
	ansi2caca := [8]canvas.Color{
		canvas.Black, canvas.Red, canvas.Green, canvas.Brown,
		canvas.Blue, canvas.Magenta, canvas.Cyan, canvas.LightGray,
	}
	for _, v := range argv {
		switch {
		case v >= 30 && v <= 37:
			im.fg = ansi2caca[v-30]
		case v >= 40 && v <= 47:
			im.bg = ansi2caca[v-40]
		case v >= 90 && v <= 97:
			im.fg = ansi2caca[v-90] + 8
		case v >= 100 && v <= 107:
			im.bg = ansi2caca[v-100] + 8
		default:
			switch v {
			case 0:
				im.fg, im.bg = im.dfg, im.dbg
				im.bold, im.blink, im.italics = false, false, false
				im.negative, im.concealed, im.underline = false, false, false
				im.faint, im.strike, im.proportional = false, false, false
			case 1:
				im.bold = true
			case 2:
				im.faint = true
			case 3:
				im.italics = true
			case 4, 21:
				im.underline = true
			case 5, 6:
				im.blink = true
			case 7:
				im.negative = true
			case 8:
				im.concealed = true
			case 9:
				im.strike = true
			case 22:
				im.bold, im.faint = false, false
			case 23:
				im.italics = false
			case 24:
				im.underline = false
			case 25:
				im.blink = false
			case 26:
				im.proportional = true
			case 27:
				im.negative = false
			case 28:
				im.concealed = false
			case 29:
				im.strike = false
			case 39:
				im.fg = im.dfg
			case 49:
				im.bg = im.dbg
			case 50:
				im.proportional = false
			}
		}
	}

	var efg, ebg canvas.Color
	if im.concealed {
		efg, ebg = canvas.Transparent, canvas.Transparent
	} else {
		if im.negative {
			efg, ebg = im.bg, im.fg
		} else {
			efg, ebg = im.fg, im.bg
		}
		if im.bold {
			if efg < 8 {
				efg += 8
			} else if efg == canvas.Default {
				efg = canvas.White
			}
		}
	}
	_ = c.SetColor(efg, ebg)
}
