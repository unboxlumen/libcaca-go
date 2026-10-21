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

// Package tcell implements the display.Backend contract on top of tcell.
//
// tcell is pure Go, so this backend keeps the whole art stack free of cgo and
// works on Windows, Linux and macOS.
package tcell

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/display"
)

// Backend is a tcell-backed display.
type Backend struct {
	screen tcell.Screen
	events chan tcell.Event
	done   chan struct{}
	once   sync.Once

	lastButtons tcell.ButtonMask
}

// New returns an uninitialised tcell backend.
func New() *Backend { return &Backend{} }

// Init implements display.Backend.
func (b *Backend) Init() error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := s.Init(); err != nil {
		return err
	}
	b.screen = s
	b.events = make(chan tcell.Event, 32)
	b.done = make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			if ev == nil {
				return
			}
			select {
			case b.events <- ev:
			case <-b.done:
				return
			}
		}
	}()
	return nil
}

// Close implements display.Backend.
func (b *Backend) Close() error {
	b.once.Do(func() {
		if b.done != nil {
			close(b.done)
		}
		if b.screen != nil {
			b.screen.Fini()
		}
	})
	return nil
}

// Size implements display.Backend.
func (b *Backend) Size() (int, int) {
	if b.screen == nil {
		return 80, 25
	}
	return b.screen.Size()
}

// SetTitle implements display.Backend.
func (b *Backend) SetTitle(title string) error {
	b.screen.SetTitle(title)
	return nil
}

// Refresh implements display.Backend.
func (b *Backend) Refresh(c *canvas.Canvas) error {
	if b.screen == nil {
		return nil
	}
	b.screen.Clear()
	w, h := b.screen.Size()
	cw, ch := c.Size()
	if cw > w {
		cw = w
	}
	if ch > h {
		ch = h
	}
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			cell := c.CellAt(x, y)
			if cell.Ch == canvas.MagicFullwidth {
				continue
			}
			b.screen.SetContent(x, y, cell.Ch, nil, styleFor(cell.A))
		}
	}
	b.screen.Show()
	return nil
}

// PollEvent implements display.Backend.
func (b *Backend) PollEvent(mask display.EventType, timeout time.Duration) (display.Event, bool) {
	if timeout == 0 {
		select {
		case ev := <-b.events:
			e, ok := b.convert(ev)
			if ok && e.Type&mask != 0 {
				return e, true
			}
			return display.Event{}, false
		default:
			return display.Event{}, false
		}
	}

	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}
	for {
		select {
		case ev := <-b.events:
			e, ok := b.convert(ev)
			if ok && e.Type&mask != 0 {
				return e, true
			}
		case <-timer:
			return display.Event{}, false
		}
	}
}

// SetMouse implements display.Backend.
func (b *Backend) SetMouse(enabled bool) {
	if b.screen == nil {
		return
	}
	if enabled {
		b.screen.EnableMouse()
	} else {
		b.screen.DisableMouse()
	}
}

// SetCursor implements display.Backend.
func (b *Backend) SetCursor(enabled bool) {
	if b.screen == nil {
		return
	}
	if enabled {
		b.screen.ShowCursor(0, 0)
	} else {
		b.screen.HideCursor()
	}
}

func (b *Backend) convert(ev tcell.Event) (display.Event, bool) {
	switch e := ev.(type) {
	case *tcell.EventKey:
		return convertKey(e)
	case *tcell.EventMouse:
		x, y := e.Position()
		buttons := e.Buttons()
		out := display.Event{X: x, Y: y}
		switch {
		case buttons != 0 && b.lastButtons == 0:
			out.Type = display.EventMousePress
			out.Button = buttonNumber(buttons)
		case buttons == 0 && b.lastButtons != 0:
			out.Type = display.EventMouseRelease
			out.Button = buttonNumber(b.lastButtons)
		case buttons != 0:
			out.Type = display.EventMouseMotion
			out.Button = buttonNumber(buttons)
		default:
			out.Type = display.EventMouseMotion
		}
		b.lastButtons = buttons
		return out, true
	case *tcell.EventResize:
		w, h := e.Size()
		return display.Event{Type: display.EventResize, W: w, H: h}, true
	}
	return display.Event{}, false
}

func buttonNumber(mask tcell.ButtonMask) int {
	switch {
	case mask&tcell.Button1 != 0:
		return 1
	case mask&tcell.Button2 != 0:
		return 2
	case mask&tcell.Button3 != 0:
		return 3
	}
	return 0
}

func convertKey(ev *tcell.EventKey) (display.Event, bool) {
	out := display.Event{Type: display.EventKeyPress}
	switch ev.Key() {
	case tcell.KeyRune:
		out.Rune = ev.Rune()
		if ev.Rune() >= 0 && ev.Rune() < 0x80 {
			out.Key = display.Key(ev.Rune())
		}
	case tcell.KeyUp:
		out.Key = display.KeyUp
	case tcell.KeyDown:
		out.Key = display.KeyDown
	case tcell.KeyLeft:
		out.Key = display.KeyLeft
	case tcell.KeyRight:
		out.Key = display.KeyRight
	case tcell.KeyInsert:
		out.Key = display.KeyInsert
	case tcell.KeyDelete:
		out.Key = display.KeyDelete
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		out.Key = display.KeyBackspace
	case tcell.KeyHome:
		out.Key = display.KeyHome
	case tcell.KeyEnd:
		out.Key = display.KeyEnd
	case tcell.KeyPgUp:
		out.Key = display.KeyPageUp
	case tcell.KeyPgDn:
		out.Key = display.KeyPageDown
	case tcell.KeyEsc:
		out.Key = display.KeyEscape
	case tcell.KeyEnter:
		out.Key = display.KeyReturn
	case tcell.KeyTab:
		out.Key = display.KeyTab
	case tcell.KeyCtrlA:
		out.Key = display.KeyCtrlA
	case tcell.KeyCtrlB:
		out.Key = display.KeyCtrlB
	case tcell.KeyCtrlC:
		out.Key = display.KeyCtrlC
	case tcell.KeyCtrlD:
		out.Key = display.KeyCtrlD
	case tcell.KeyCtrlE:
		out.Key = display.KeyCtrlE
	case tcell.KeyCtrlF:
		out.Key = display.KeyCtrlF
	case tcell.KeyCtrlG:
		out.Key = display.KeyCtrlG
	case tcell.KeyCtrlH:
		out.Key = display.KeyBackspace
	case tcell.KeyCtrlJ:
		out.Key = display.KeyCtrlJ
	case tcell.KeyCtrlK:
		out.Key = display.KeyCtrlK
	case tcell.KeyCtrlL:
		out.Key = display.KeyCtrlL
	case tcell.KeyCtrlN:
		out.Key = display.KeyCtrlN
	case tcell.KeyCtrlO:
		out.Key = display.KeyCtrlO
	case tcell.KeyCtrlP:
		out.Key = display.KeyCtrlP
	case tcell.KeyCtrlQ:
		out.Key = display.KeyCtrlQ
	case tcell.KeyCtrlR:
		out.Key = display.KeyCtrlR
	case tcell.KeyCtrlS:
		out.Key = display.KeyPause // libcaca's 0x13 is Pause (Ctrl-S)
	case tcell.KeyCtrlT:
		out.Key = display.KeyCtrlT
	case tcell.KeyCtrlU:
		out.Key = display.KeyCtrlU
	case tcell.KeyCtrlV:
		out.Key = display.KeyCtrlV
	case tcell.KeyCtrlW:
		out.Key = display.KeyCtrlW
	case tcell.KeyCtrlX:
		out.Key = display.KeyCtrlX
	case tcell.KeyCtrlY:
		out.Key = display.KeyCtrlY
	case tcell.KeyCtrlZ:
		out.Key = display.KeyCtrlZ
	case tcell.KeyF1:
		out.Key = display.KeyF1
	case tcell.KeyF2:
		out.Key = display.KeyF2
	case tcell.KeyF3:
		out.Key = display.KeyF3
	case tcell.KeyF4:
		out.Key = display.KeyF4
	case tcell.KeyF5:
		out.Key = display.KeyF5
	case tcell.KeyF6:
		out.Key = display.KeyF6
	case tcell.KeyF7:
		out.Key = display.KeyF7
	case tcell.KeyF8:
		out.Key = display.KeyF8
	case tcell.KeyF9:
		out.Key = display.KeyF9
	case tcell.KeyF10:
		out.Key = display.KeyF10
	case tcell.KeyF11:
		out.Key = display.KeyF11
	case tcell.KeyF12:
		out.Key = display.KeyF12
	default:
		return display.Event{}, false
	}
	return out, true
}

func styleFor(a canvas.Attr) tcell.Style {
	st := tcell.StyleDefault
	if c := a.FG(); c == canvas.Default || c == canvas.Transparent {
		st = st.Foreground(tcell.ColorDefault)
	} else {
		rgb := a.RGB24FG()
		st = st.Foreground(tcell.NewRGBColor(int32(rgb>>16)&0xff, int32(rgb>>8)&0xff, int32(rgb)&0xff))
	}
	if c := a.BG(); c == canvas.Default || c == canvas.Transparent {
		st = st.Background(tcell.ColorDefault)
	} else {
		rgb := a.RGB24BG()
		st = st.Background(tcell.NewRGBColor(int32(rgb>>16)&0xff, int32(rgb>>8)&0xff, int32(rgb)&0xff))
	}
	sty := a.Styles()
	if sty&canvas.Bold != 0 {
		st = st.Bold(true)
	}
	if sty&canvas.Italics != 0 {
		st = st.Italic(true)
	}
	if sty&canvas.Underline != 0 {
		st = st.Underline(true)
	}
	if sty&canvas.Blink != 0 {
		st = st.Blink(true)
	}
	return st
}
