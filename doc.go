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

// Package art is the umbrella package of nfilesart, a pure Go colour
// ASCII-art toolkit ported from the WTFPL-licensed core of libcaca.
//
// # Lineage
//
// nfilesart is not a binding and does not link against libcaca. It is an
// independent Go reimplementation whose behaviour is modelled on the C
// sources of libcaca 0.99.beta20 (commit
// de48e1a0888caabdc6276afdea46e9096d0adb05). The attribute encoding, the
// dirty-rectangle merging, the dithering pipeline and the byte layout of the
// import and export formats are intentionally identical, so output can be
// compared byte-for-byte with the C library. See PROVENANCE.md for the exact
// list of source files, and LICENSE for the WTFPL v2 terms.
//
// # Packages
//
//   - canvas — the character grid: cells, attributes, frames, dirty
//     rectangles, drawing primitives and transforms.
//   - dither — bitmap to character-cell conversion with six algorithms,
//     seven colour modes and three glyph ramps.
//   - codec — libcaca-compatible import and export of the native caca,
//     ANSI, UTF-8, HTML, HTML3, BBCode, IRC, PostScript, SVG, TGA and troff
//     formats.
//   - font — bitmap fonts (Go Mono by default, or libcaca's native .data
//     format), FIGlet/TOIlet FIGfonts, and canvas rendering.
//   - charset — UTF-32 width classification plus CP437 and ASCII mappings.
//   - display — the display-backend contract and the libcaca event model.
//   - backend/tcell — a pure Go terminal backend built on tcell.
//   - backend/null — a no-op backend for tests and headless use.
//   - preview — image to character art to bitmap, for GUI preview panes.
//
// # Example
//
// A minimal pipeline: dither an image into a canvas and export ANSI art.
//
//	d, err := dither.New(img, dither.DefaultOptions())
//	if err != nil {
//		return err
//	}
//	c, err := canvas.New(80, 24)
//	if err != nil {
//		return err
//	}
//	if err := d.Draw(c, 0, 0, 80, 24); err != nil {
//		return err
//	}
//	out, err := codec.Export(c, "ansi")
//
// The cmd/ directory contains runnable demos: cmd/demo draws an animated
// terminal scene, and cmd/img2txt renders an image file into the terminal or
// into exported text formats.
//
// # Concurrency
//
// A Canvas is not safe for concurrent use, matching libcaca. A Dither is safe
// to reuse for repeated Draw calls on one goroutine; each Dither owns its own
// error-diffusion state, so distinct Dither values may run in parallel.
package art
