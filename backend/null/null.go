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

// Package null implements a display backend that draws nothing. It is useful
// for tests and for running applications headless.
package null

import (
	"time"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/display"
)

// Backend is a no-op display.
type Backend struct {
	W, H int
}

// New returns a null backend of the given size.
func New(w, h int) *Backend { return &Backend{W: w, H: h} }

// Init implements display.Backend.
func (b *Backend) Init() error { return nil }

// Close implements display.Backend.
func (b *Backend) Close() error { return nil }

// Size implements display.Backend.
func (b *Backend) Size() (int, int) { return b.W, b.H }

// SetTitle implements display.Backend.
func (b *Backend) SetTitle(string) error { return nil }

// Refresh implements display.Backend.
func (b *Backend) Refresh(*canvas.Canvas) error { return nil }

// PollEvent implements display.Backend.
func (b *Backend) PollEvent(display.EventType, time.Duration) (display.Event, bool) {
	return display.Event{}, false
}

// SetMouse implements display.Backend.
func (b *Backend) SetMouse(bool) {}

// SetCursor implements display.Backend.
func (b *Backend) SetCursor(bool) {}
