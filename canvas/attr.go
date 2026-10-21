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

// Package canvas implements the libcaca character-grid model.
//
// The attribute encoding and the cell semantics intentionally match
// libcaca 0.99.beta20: a 32-bit attribute holds a 14-bit background colour,
// a 14-bit foreground colour and four style bits. Keeping the original
// encoding means the exporters and the tests ported from libcaca can be
// compared byte-for-byte.
package canvas

import "fmt"

// Attr is a libcaca text attribute.
//
// Layout (MSB to LSB):
//
//	bits 31..18 background: 3-bit alpha, 4-bit red, 4-bit green, 3-bit blue
//	bits 17..4  foreground: 3-bit alpha, 4-bit red, 4-bit green, 3-bit blue
//	bits 3..0   styles: bold, italics, underline, blink
type Attr uint32

// Color is an ANSI colour index or one of Default/Transparent.
type Color uint8

// Style is a bitmask of text styles.
type Style uint32

const (
	Black Color = 0x00
	Blue  Color = 0x01
	Green Color = 0x02
	Cyan  Color = 0x03
	Red   Color = 0x04
	// Magenta is the ANSI colour between red and brown.
	Magenta      Color = 0x05
	Brown        Color = 0x06
	LightGray    Color = 0x07
	DarkGray     Color = 0x08
	LightBlue    Color = 0x09
	LightGreen   Color = 0x0a
	LightCyan    Color = 0x0b
	LightRed     Color = 0x0c
	LightMagenta Color = 0x0d
	Yellow       Color = 0x0e
	White        Color = 0x0f

	Default     Color = 0x10
	Transparent Color = 0x20
)

const (
	Bold      Style = 0x01
	Italics   Style = 0x02
	Underline Style = 0x04
	Blink     Style = 0x08
)

const (
	styleMask Attr = 0x0000000f
	fgMask    Attr = 0x0003fff0
	bgMask    Attr = 0xfffc0000
)

// ansiTab16 is the gnome-terminal ANSI palette, as used by libcaca.
// Brown is deliberately 0xfa50 instead of 0xfaa0.
var ansiTab16 = [16]uint16{
	0xf000, 0xf00a, 0xf0a0, 0xf0aa, 0xfa00, 0xfa0a, 0xfa50, 0xfaaa,
	0xf555, 0xf55f, 0xf5f5, 0xf5ff, 0xff55, 0xff5f, 0xfff5, 0xffff,
}

// ansiTab14 is the same palette on 14 bits (3-4-4-3).
var ansiTab14 = [16]uint16{
	0x3800, 0x3805, 0x3850, 0x3855, 0x3d00, 0x3d05, 0x3d28, 0x3d55,
	0x3aaa, 0x3aaf, 0x3afa, 0x3aff, 0x3faa, 0x3faf, 0x3ffa, 0x3fff,
}

// ValidColor reports whether c can be passed to SetColor.
func ValidColor(c Color) bool { return c <= Transparent }

// ColorAttr builds an attribute from an ANSI colour pair. The colours must
// satisfy ValidColor; invalid values are treated as Default/Transparent.
func ColorAttr(fg, bg Color) Attr {
	if !ValidColor(fg) {
		fg = Default
	}
	if !ValidColor(bg) {
		bg = Transparent
	}
	return Attr((uint32(bg)|0x40)<<18 | (uint32(fg)|0x40)<<4)
}

// DefaultAttr is the attribute of a freshly created canvas:
// default foreground on a transparent background.
func DefaultAttr() Attr { return ColorAttr(Default, Transparent) }

// WithColor returns a with the given ANSI colours and its styles preserved.
func (a Attr) WithColor(fg, bg Color) Attr {
	return (a & styleMask) | ColorAttr(fg, bg)
}

// WithStyles returns a with the given style bits.
func (a Attr) WithStyles(s Style) Attr {
	return (a &^ styleMask) | Attr(s&0xf)
}

// Styles returns the style bits of a.
func (a Attr) Styles() Style { return Style(a & styleMask) }

// Set returns a with the given style bits set.
func (a Attr) Set(s Style) Attr { return a | Attr(s&0xf) }

// Unset returns a with the given style bits cleared.
func (a Attr) Unset(s Style) Attr { return a &^ Attr(s&0xf) }

// Toggle returns a with the given style bits flipped.
func (a Attr) Toggle(s Style) Attr { return a ^ Attr(s&0xf) }

// FG returns the nearest ANSI foreground colour.
func (a Attr) FG() Color { return Color(nearestANSI(uint16((a >> 4) & 0x3fff))) }

// BG returns the nearest ANSI background colour.
func (a Attr) BG() Color { return Color(nearestANSI(uint16(a >> 18))) }

// ANSI returns the packed DOS ANSI value: high nibble background, low
// nibble foreground. Default and transparent are replaced by LightGray and
// Black respectively, matching libcaca.
func (a Attr) ANSI() uint8 {
	fg := a.FG()
	bg := a.BG()
	if fg >= 0x10 {
		fg = LightGray
	}
	if bg >= 0x10 {
		bg = Black
	}
	return uint8(fg) | uint8(bg)<<4
}

// RGB12FG returns the 12-bit RGB foreground value (R<<8 | G<<4 | B).
func (a Attr) RGB12FG() uint16 {
	fg := uint16((a >> 4) & 0x3fff)
	if fg < (0x10 | 0x40) {
		return ansiTab16[fg^0x40] & 0x0fff
	}
	if fg == (uint16(Default)|0x40) || fg == (uint16(Transparent)|0x40) {
		return ansiTab16[LightGray] & 0x0fff
	}
	return (fg << 1) & 0x0fff
}

// RGB12BG returns the 12-bit RGB background value (R<<8 | G<<4 | B).
func (a Attr) RGB12BG() uint16 {
	bg := uint16(a >> 18)
	if bg < (0x10 | 0x40) {
		return ansiTab16[bg^0x40] & 0x0fff
	}
	if bg == (uint16(Default)|0x40) || bg == (uint16(Transparent)|0x40) {
		return ansiTab16[Black] & 0x0fff
	}
	return (bg << 1) & 0x0fff
}

// ARGB64 expands the attribute into eight 4-bit components:
// bg alpha, bg r, bg g, bg b, fg alpha, fg r, fg g, fg b.
func (a Attr) ARGB64() [8]uint8 {
	var out [8]uint8
	fg := uint16((a >> 4) & 0x3fff)
	bg := uint16(a >> 18)

	if bg < (0x10 | 0x40) {
		bg = ansiTab16[bg^0x40]
	} else if bg == (uint16(Default) | 0x40) {
		bg = ansiTab16[Black]
	} else if bg == (uint16(Transparent) | 0x40) {
		bg = 0x0fff
	} else {
		bg = ((bg << 2) & 0xf000) | ((bg << 1) & 0x0fff)
	}

	out[0] = uint8(bg >> 12)
	out[1] = uint8((bg >> 8) & 0xf)
	out[2] = uint8((bg >> 4) & 0xf)
	out[3] = uint8(bg & 0xf)

	if fg < (0x10 | 0x40) {
		fg = ansiTab16[fg^0x40]
	} else if fg == (uint16(Default) | 0x40) {
		fg = ansiTab16[LightGray]
	} else if fg == (uint16(Transparent) | 0x40) {
		fg = 0x0fff
	} else {
		fg = ((fg << 2) & 0xf000) | ((fg << 1) & 0x0fff)
	}

	out[4] = uint8(fg >> 12)
	out[5] = uint8((fg >> 8) & 0xf)
	out[6] = uint8((fg >> 4) & 0xf)
	out[7] = uint8(fg & 0xf)
	return out
}

// RGB24FG returns the 24-bit foreground colour.
func (a Attr) RGB24FG() uint32 { return rgb12to24(a.RGB12FG()) }

// RGB24BG returns the 24-bit background colour.
func (a Attr) RGB24BG() uint32 { return rgb12to24(a.RGB12BG()) }

func rgb12to24(i uint16) uint32 {
	return uint32((i&0xf00)>>8)*0x110000 |
		uint32((i&0x0f0)>>4)*0x001100 |
		uint32(i&0x00f)*0x000011
}

func nearestANSI(argb14 uint16) uint8 {
	if argb14 < (0x10 | 0x40) {
		return uint8(argb14 ^ 0x40)
	}
	if argb14 == (uint16(Default)|0x40) || argb14 == (uint16(Transparent)|0x40) {
		return uint8(argb14 ^ 0x40)
	}
	if argb14 < 0x0fff { // too transparent
		return uint8(Transparent)
	}

	best := uint8(Default)
	dist := uint32(0x3fff)
	for i := 0; i < 16; i++ {
		var d uint32
		a := int((ansiTab14[i] >> 7) & 0xf)
		b := int((argb14 >> 7) & 0xf)
		d += uint32((a - b) * (a - b))
		a = int((ansiTab14[i] >> 3) & 0xf)
		b = int((argb14 >> 3) & 0xf)
		d += uint32((a - b) * (a - b))
		a = int((ansiTab14[i] << 1) & 0xf)
		b = int((argb14 << 1) & 0xf)
		d += uint32((a - b) * (a - b))
		if d < dist {
			dist = d
			best = uint8(i)
		}
	}
	return best
}

// String returns a human-readable rendering of the attribute's nearest ANSI
// colours and its style bits.
func (a Attr) String() string {
	return fmt.Sprintf("Attr(fg=%d bg=%d style=%#x)", a.FG(), a.BG(), a.Styles())
}

// SetColor validates and installs an ANSI colour pair. Only the colour
// bits change; styles are preserved.
func (c *Canvas) SetColor(fg, bg Color) error {
	if !ValidColor(fg) || !ValidColor(bg) {
		return ErrInvalidColor
	}
	f := c.frame()
	f.curAttr = f.curAttr.WithColor(fg, bg)
	return nil
}

// SetColorARGB installs a 16-bit ARGB colour pair.
func (c *Canvas) SetColorARGB(fg, bg uint16) {
	if fg < 0x100 {
		fg += 0x100
	}
	if bg < 0x100 {
		bg += 0x100
	}
	fg = ((fg >> 1) & 0x7ff) | ((fg >> 13) << 11)
	bg = ((bg >> 1) & 0x7ff) | ((bg >> 13) << 11)
	f := c.frame()
	f.curAttr = (f.curAttr & styleMask) | Attr(bg)<<18 | Attr(fg)<<4
}

// SetAttr installs a default attribute. Values below 0x10 are treated as
// style bits and only modify the style portion, like libcaca.
func (c *Canvas) SetAttr(a Attr) {
	f := c.frame()
	if a < 0x00000010 {
		f.curAttr = (f.curAttr &^ styleMask) | a
		return
	}
	f.curAttr = a
}
