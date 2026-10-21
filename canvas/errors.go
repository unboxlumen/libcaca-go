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

import "errors"

var (
	// ErrInvalidColor is returned when an ANSI colour is outside 0..0x20.
	ErrInvalidColor = errors.New("canvas: invalid colour")
	// ErrInvalidSize is returned for negative or overflowing dimensions.
	ErrInvalidSize = errors.New("canvas: invalid canvas size")
	// ErrNoFrame is returned when a frame index is out of range.
	ErrNoFrame = errors.New("canvas: no such frame")
	// ErrManaged is returned when a resize is refused by a resize callback.
	ErrManaged = errors.New("canvas: canvas is managed and refused resize")
	// ErrTooLarge is returned when a canvas would exceed MaxCells.
	ErrTooLarge = errors.New("canvas: canvas too large")
	// ErrDirtyNotDisabled is returned by EnableDirtyRect when tracking was not
	// suspended.
	ErrDirtyNotDisabled = errors.New("canvas: dirty rectangles were not disabled")
	// ErrNotManaged is returned by UnmanageCanvas on an unmanaged canvas.
	ErrNotManaged = errors.New("canvas: canvas is not managed")
)
