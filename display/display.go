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

// Package display defines the backend contract used to present a character
// canvas and receive input events.
package display

import (
	"errors"
	"time"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// EventType is a libcaca event type bit.
type EventType uint32

const (
	EventNone         EventType = 0x0000
	EventKeyPress     EventType = 0x0001
	EventKeyRelease   EventType = 0x0002
	EventMousePress   EventType = 0x0004
	EventMouseRelease EventType = 0x0008
	EventMouseMotion  EventType = 0x0010
	EventResize       EventType = 0x0020
	EventQuit         EventType = 0x0040
	EventAny          EventType = 0xffff
)

// Key mirrors libcaca's caca_key values so applications ported from C keep
// their numeric constants.
type Key int

const (
	KeyUnknown   Key = 0x00
	KeyCtrlA     Key = 0x01
	KeyCtrlB     Key = 0x02
	KeyCtrlC     Key = 0x03
	KeyCtrlD     Key = 0x04
	KeyCtrlE     Key = 0x05
	KeyCtrlF     Key = 0x06
	KeyCtrlG     Key = 0x07
	KeyBackspace Key = 0x08
	KeyTab       Key = 0x09
	KeyCtrlJ     Key = 0x0a
	KeyCtrlK     Key = 0x0b
	KeyCtrlL     Key = 0x0c
	KeyReturn    Key = 0x0d
	KeyCtrlN     Key = 0x0e
	KeyCtrlO     Key = 0x0f
	KeyCtrlP     Key = 0x10
	KeyCtrlQ     Key = 0x11
	KeyCtrlR     Key = 0x12
	KeyPause     Key = 0x13
	KeyCtrlT     Key = 0x14
	KeyCtrlU     Key = 0x15
	KeyCtrlV     Key = 0x16
	KeyCtrlW     Key = 0x17
	KeyCtrlX     Key = 0x18
	KeyCtrlY     Key = 0x19
	KeyCtrlZ     Key = 0x1a
	KeyEscape    Key = 0x1b
	KeyDelete    Key = 0x7f

	KeyUp       Key = 0x111
	KeyDown     Key = 0x112
	KeyLeft     Key = 0x113
	KeyRight    Key = 0x114
	KeyInsert   Key = 0x115
	KeyHome     Key = 0x116
	KeyEnd      Key = 0x117
	KeyPageUp   Key = 0x118
	KeyPageDown Key = 0x119

	KeyF1  Key = 0x11a
	KeyF2  Key = 0x11b
	KeyF3  Key = 0x11c
	KeyF4  Key = 0x11d
	KeyF5  Key = 0x11e
	KeyF6  Key = 0x11f
	KeyF7  Key = 0x120
	KeyF8  Key = 0x121
	KeyF9  Key = 0x122
	KeyF10 Key = 0x123
	KeyF11 Key = 0x124
	KeyF12 Key = 0x125
	KeyF13 Key = 0x126
	KeyF14 Key = 0x127
	KeyF15 Key = 0x128
)

// Event is one input or window event.
type Event struct {
	Type   EventType
	Key    Key
	Rune   rune
	X, Y   int
	Button int
	W, H   int
}

// Backend is the display-driver contract. It is deliberately shaped after
// libcaca's driver vtable.
type Backend interface {
	Init() error
	Close() error
	Size() (w, h int)
	SetTitle(title string) error
	Refresh(c *canvas.Canvas) error
	PollEvent(mask EventType, timeout time.Duration) (Event, bool)
	SetMouse(enabled bool)
	SetCursor(enabled bool)
}

// ErrNoBackend is returned when a nil backend is passed to New.
var ErrNoBackend = errors.New("display: nil backend")

// Display couples a canvas with a backend.
type Display struct {
	backend Backend
	canvas  *canvas.Canvas
}

// New initialises backend and binds it to c. The canvas is not resized; the
// caller can query the backend size and resize it.
func New(backend Backend, c *canvas.Canvas) (*Display, error) {
	if backend == nil {
		return nil, ErrNoBackend
	}
	if err := backend.Init(); err != nil {
		return nil, err
	}
	return &Display{backend: backend, canvas: c}, nil
}

// Canvas returns the bound canvas.
func (d *Display) Canvas() *canvas.Canvas { return d.canvas }

// Size returns the backend's display size in cells.
func (d *Display) Size() (w, h int) { return d.backend.Size() }

// SetTitle changes the window title.
func (d *Display) SetTitle(title string) error { return d.backend.SetTitle(title) }

// Refresh presents the canvas.
func (d *Display) Refresh() error { return d.backend.Refresh(d.canvas) }

// PollEvent waits for an event matching mask. A zero timeout polls once.
func (d *Display) PollEvent(mask EventType, timeout time.Duration) (Event, bool) {
	return d.backend.PollEvent(mask, timeout)
}

// SetMouse enables or disables mouse reporting.
func (d *Display) SetMouse(enabled bool) { d.backend.SetMouse(enabled) }

// SetCursor shows or hides the cursor.
func (d *Display) SetCursor(enabled bool) { d.backend.SetCursor(enabled) }

// Close releases the backend.
func (d *Display) Close() error { return d.backend.Close() }
