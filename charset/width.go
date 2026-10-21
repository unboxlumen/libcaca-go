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

// Package charset contains the character-set conversions used by the
// character canvas: UTF-32 width classification and CP437 mapping.
package charset

// IsFullwidth reports whether ch occupies two terminal cells.
//
// This is the exact table used by libcaca 0.99.beta20. It is intentionally
// not replaced by uniseg here: the canvas and its exporters must agree with
// the C implementation. Display adapters can use uniseg for the final
// terminal cell layout.
func IsFullwidth(ch rune) bool {
	u := uint32(ch)
	switch {
	case u < 0x2e80: // standard stuff
		return false
	case u < 0xa700: // Japanese, Korean, CJK, Yi...
		return true
	case u < 0xac00: // modified tone letters, Syloti Nagri
		return false
	case u < 0xd800: // Hangul syllables
		return true
	case u < 0xf900:
		return false
	case u < 0xfb00: // more CJK
		return true
	case u < 0xfe20:
		return false
	case u < 0xfe70: // more CJK
		return true
	case u < 0xff00:
		return false
	case u < 0xff61: // fullwidth forms
		return true
	case u < 0xffe0: // halfwidth forms
		return false
	case u < 0xffe8:
		return true
	case u < 0x20000:
		return false
	case u < 0xe0000: // more CJK
		return true
	default:
		return false
	}
}
