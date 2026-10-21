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

// Command demo draws an animated scene on a tcell terminal: a bouncing ball,
// a colour strip and a status line. It is the nfilesart counterpart of
// libcaca's cacademo, and the Go equivalent of the demo shipped by the
// unrelated gocaca project.
//
// Usage:
//
//	go run ./cmd/demo
//
// Press q, Q or Ctrl-C to quit. When stdout is not a terminal, the command
// writes one still frame as UTF-8 with ANSI escapes and exits, so it can be
// piped or used in tests.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/unboxlumen/libcaca-go/backend/tcell"
	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/display"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func run() error {
	c, err := canvas.New(60, 18)
	if err != nil {
		return err
	}

	backend := tcell.New()
	d, err := display.New(backend, c)
	if err != nil {
		// No usable terminal: draw one frame and exit.
		_ = backend.Close()
		draw(c, 8, 5, 0, 60, 18)
		return printFrame(c)
	}
	defer d.Close()

	w, h := d.Size()
	if w > 0 && h > 0 {
		if err := c.Resize(w, h); err != nil {
			return err
		}
	}
	_ = d.SetTitle("nfilesart demo")
	d.SetCursor(false)

	bx, by := 10, 5
	dx, dy := 1, 1
	for frame := 0; ; frame++ {
		w, h := c.Size()
		bx, by, dx, dy = advance(bx, by, dx, dy, w, h)
		draw(c, bx, by, frame, w, h)
		if err := d.Refresh(); err != nil {
			return err
		}

		ev, ok := d.PollEvent(display.EventAny, 50*time.Millisecond)
		if !ok {
			continue
		}
		switch ev.Type {
		case display.EventKeyPress:
			if ev.Key == 'q' || ev.Key == 'Q' || ev.Key == display.KeyCtrlC {
				return nil
			}
		case display.EventQuit:
			return nil
		case display.EventResize:
			if ev.W > 0 && ev.H > 0 {
				if err := c.Resize(ev.W, ev.H); err != nil {
					return err
				}
				bx = min(bx, ev.W-3)
				by = min(by, ev.H-3)
			}
		}
	}
}

// advance moves the ball and bounces it off the inner border.
func advance(bx, by, dx, dy, w, h int) (int, int, int, int) {
	if w < 5 || h < 5 {
		return 1, 1, dx, dy
	}
	bx += dx
	by += dy
	if bx <= 2 || bx >= w-3 {
		dx = -dx
		bx += 2 * dx
	}
	if by <= 2 || by >= h-3 {
		dy = -dy
		by += 2 * dy
	}
	return bx, by, dx, dy
}

func draw(c *canvas.Canvas, bx, by, frame, w, h int) {
	_ = c.SetColor(canvas.LightGray, canvas.Black)
	c.Clear()
	c.DrawThinBox(0, 0, w, h)

	// Title bar.
	_ = c.SetColor(canvas.White, canvas.Blue)
	c.FillBox(1, 0, maxInt(0, w-2), 1, ' ')
	c.PutString(2, 0, " nfilesart demo ")

	if w > 4 && h > 6 {
		// An ANSI colour strip.
		for i := 0; i < 8 && i < w-2; i++ {
			_ = c.SetColor(canvas.Color(i+8), canvas.Black)
			c.Put(i+1, 2, '█')
		}
		// The bouncing ball.
		_ = c.SetColor(canvas.Yellow, canvas.Black)
		c.FillEllipse(bx, by, 2, 1, '.')
		c.Put(bx, by, 'O')

		_ = c.SetColor(canvas.Green, canvas.Black)
		c.PutString(2, h-3, fmt.Sprintf("frame %d · q/Ctrl-C to quit · resize me", frame))
		_ = c.SetColor(canvas.DarkGray, canvas.Black)
		c.PutString(2, h-2, "WTFPL v2 · pure Go · no cgo")
	}
}

// printFrame writes one UTF-8 ANSI frame to stdout.
func printFrame(c *canvas.Canvas) error {
	data, err := codec.Export(c, "utf8")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
