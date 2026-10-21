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

import "testing"

func TestNewAndResize(t *testing.T) {
	c, err := New(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := c.Size(); w != 0 || h != 0 {
		t.Fatalf("size = %dx%d, want 0x0", w, h)
	}

	for _, tc := range []struct{ w, h int }{{1, 1}, {1234, 1001}, {0, 0}} {
		if err := c.Resize(tc.w, tc.h); err != nil {
			t.Fatalf("resize %dx%d: %v", tc.w, tc.h, err)
		}
		if w, h := c.Size(); w != tc.w || h != tc.h {
			t.Fatalf("size = %dx%d, want %dx%d", w, h, tc.w, tc.h)
		}
	}

	for _, tc := range []struct{ w, h int }{{-1, 50}, {50, -1}, {-1, -1}} {
		if err := c.Resize(tc.w, tc.h); err == nil {
			t.Fatalf("resize %dx%d: expected error", tc.w, tc.h)
		}
	}
}

func TestPutGet(t *testing.T) {
	c := MustNew(0, 0)
	if got := c.Get(0, 0); got != ' ' {
		t.Fatalf("get on empty canvas = %q", got)
	}
	c.Put(0, 0, 'x')
	if got := c.Get(0, 0); got != ' ' {
		t.Fatalf("put outside canvas leaked: %q", got)
	}

	c.Resize(1, 1)
	if got := c.Get(0, 0); got != ' ' {
		t.Fatalf("resized cell = %q, want space", got)
	}
	c.Put(0, 0, 'x')
	if got := c.Get(0, 0); got != 'x' {
		t.Fatalf("get = %q, want x", got)
	}
	c.Put(0, 0, 'y')
	if got := c.Get(0, 0); got != 'y' {
		t.Fatalf("get = %q, want y", got)
	}

	c.Resize(1000, 1000)
	if got := c.Get(999, 999); got != ' ' {
		t.Fatalf("new cell = %q, want space", got)
	}
	c.Put(999, 999, 'z')
	if got := c.Get(999, 999); got != 'z' {
		t.Fatalf("get = %q, want z", got)
	}
}

func TestFullwidthPut(t *testing.T) {
	c := MustNew(10, 2)
	if n := c.Put(7, 0, 0x2f06); n != 2 {
		t.Fatalf("put width = %d, want 2", n)
	}
	if got := c.Get(8, 0); got != MagicFullwidth {
		t.Fatalf("right half = %#x, want magic", got)
	}

	// Overwrite the right half: the left half becomes a space.
	c.Put(8, 0, 'x')
	if got := c.Get(7, 0); got != ' ' {
		t.Fatalf("left half = %q, want space", got)
	}
	if got := c.Get(8, 0); got != 'x' {
		t.Fatalf("right half = %q, want x", got)
	}

	// A fullwidth character at the last column degrades to a space.
	if n := c.Put(9, 1, 0x2f06); n != 2 {
		t.Fatalf("edge width = %d, want 2", n)
	}
	if got := c.Get(9, 1); got != ' ' {
		t.Fatalf("edge cell = %q, want space", got)
	}
}

func TestFlip(t *testing.T) {
	c := MustNew(3, 2)
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'H')
	c.Put(2, 0, 'M')
	c.Put(0, 1, '(')
	c.Flip()

	want := []rune{'M', 'H', 'A', ' ', ' ', ')'}
	for i, w := range want {
		if got := c.Cells()[i].Ch; got != w {
			t.Fatalf("cell %d = %q, want %q", i, got, w)
		}
	}

	c.Flip()
	want = []rune{'A', 'H', 'M', '(', ' ', ' '}
	for i, w := range want {
		if got := c.Cells()[i].Ch; got != w {
			t.Fatalf("after second flip cell %d = %q, want %q", i, got, w)
		}
	}
}

func TestFlopRotateInvolutions(t *testing.T) {
	orig := MustNew(7, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 7; x++ {
			orig.Put(x, y, rune('a'+x+y*7))
		}
	}

	c := clone(orig)
	c.Flop()
	c.Flop()
	assertCells(t, c, orig)

	c = clone(orig)
	c.Rotate180()
	c.Rotate180()
	assertCells(t, c, orig)

	c = clone(orig)
	for i := 0; i < 4; i++ {
		c.RotateLeft()
	}
	// Rotating left halves an odd width (rounded up), so four rotations
	// do not restore a 7x5 canvas; it stabilises at 8x5.
	assertSize(t, c, MustNew(8, 5))
}

func TestResizePreservesContent(t *testing.T) {
	c := MustNew(4, 3)
	c.Put(0, 0, 'a')
	c.Put(3, 2, 'z')
	if err := c.Resize(6, 5); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("kept cell = %q", got)
	}
	if got := c.Get(3, 2); got != 'z' {
		t.Fatalf("kept cell = %q", got)
	}
	if got := c.Get(5, 4); got != ' ' {
		t.Fatalf("new cell = %q", got)
	}

	if err := c.Resize(2, 2); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("kept cell after shrink = %q", got)
	}
}

func TestFrames(t *testing.T) {
	c := MustNew(2, 2)
	c.Put(0, 0, 'a')
	if c.FrameCount() != 1 {
		t.Fatalf("frames = %d", c.FrameCount())
	}
	if err := c.CreateFrame(1); err != nil {
		t.Fatal(err)
	}
	if c.FrameCount() != 2 {
		t.Fatalf("frames = %d", c.FrameCount())
	}
	if err := c.SetFrame(1); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("new frame cell = %q, want the current frame's copy", got)
	}
	c.Put(0, 0, 'b')
	if err := c.SetFrame(0); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("frame 0 cell = %q after editing frame 1", got)
	}
	if err := c.SetFrame(0); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("frame 0 cell = %q", got)
	}
	if c.FrameName() != "frame#00000000" {
		t.Fatalf("frame name = %q", c.FrameName())
	}
	c.SetFrameName("first")
	if c.FrameName() != "first" {
		t.Fatalf("frame name = %q", c.FrameName())
	}
	if err := c.FreeFrame(1); err != nil {
		t.Fatal(err)
	}
	if c.FrameCount() != 1 {
		t.Fatalf("frames = %d", c.FrameCount())
	}
}

func TestBlit(t *testing.T) {
	dst := MustNew(5, 3)
	src := MustNew(2, 2)
	src.Put(0, 0, 'a')
	src.Put(1, 1, 'b')
	if err := dst.Blit(1, 1, src, nil); err != nil {
		t.Fatal(err)
	}
	if got := dst.Get(1, 1); got != 'a' {
		t.Fatalf("blit cell = %q", got)
	}
	if got := dst.Get(2, 2); got != 'b' {
		t.Fatalf("blit cell = %q", got)
	}

	dst2 := MustNew(5, 3)
	mask := MustNew(2, 2)
	mask.Put(0, 0, 'x')
	if err := dst2.Blit(0, 0, src, mask); err != nil {
		t.Fatal(err)
	}
	if got := dst2.Get(0, 0); got != 'a' {
		t.Fatalf("masked blit cell = %q", got)
	}
	if got := dst2.Get(1, 1); got != ' ' {
		t.Fatalf("mask skipped cell = %q", got)
	}
}

func TestAttrConversions(t *testing.T) {
	a := ColorAttr(Red, Blue)
	if got := a.FG(); got != Red {
		t.Fatalf("fg = %d", got)
	}
	if got := a.BG(); got != Blue {
		t.Fatalf("bg = %d", got)
	}
	if got := a.ANSI(); got != uint8(Red)|uint8(Blue)<<4 {
		t.Fatalf("ansi = %#x", got)
	}
	if got := a.RGB12FG(); got != ansiTab16[Red]&0x0fff {
		t.Fatalf("rgb12 fg = %#x", got)
	}

	d := DefaultAttr()
	if d.FG() != Default {
		t.Fatalf("default fg = %d", d.FG())
	}
	if d.BG() != Transparent {
		t.Fatalf("default bg = %d", d.BG())
	}
	if d != DefaultAttr() {
		t.Fatal("default attr mismatch")
	}
}

func TestDirtyBasics(t *testing.T) {
	c := MustNew(80, 25)
	if got := c.DirtyRectCount(); got != 1 {
		t.Fatalf("initial dirty count = %d", got)
	}
	r := c.DirtyRects()[0]
	if r != (Rect{0, 0, 80, 25}) {
		t.Fatalf("initial dirty = %+v", r)
	}
	c.ClearDirtyRects()
	if c.DirtyRectCount() != 0 {
		t.Fatal("dirty list not cleared")
	}

	c.Put(7, 3, 'x')
	if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{7, 3, 1, 1}) {
		t.Fatalf("put dirty = %+v", got)
	}

	c.ClearDirtyRects()
	c.Put(7, 3, 'x')
	if c.DirtyRectCount() != 0 {
		t.Fatalf("same char dirty = %+v", c.DirtyRects())
	}
}

func TestDirtyFullwidth(t *testing.T) {
	c := MustNew(80, 25)
	c.Put(7, 3, 0x2f06)
	c.ClearDirtyRects()
	c.Put(7, 3, 'x')
	if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{7, 3, 2, 1}) {
		t.Fatalf("overwrite left dirty = %+v", got)
	}

	c.Clear()
	c.Put(7, 3, 0x2f06)
	c.ClearDirtyRects()
	c.Put(8, 3, 'x')
	if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{7, 3, 2, 1}) {
		t.Fatalf("overwrite right dirty = %+v", got)
	}

	c.Clear()
	c.Put(7, 3, 0x2f06)
	c.ClearDirtyRects()
	c.Put(6, 3, 0x2f06)
	if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{6, 3, 3, 1}) {
		t.Fatalf("fullwidth over fullwidth dirty = %+v", got)
	}
}

func TestDirtySimplify(t *testing.T) {
	c := MustNew(80, 25)
	c.ClearDirtyRects()
	for i := 0; i < 10; i++ {
		c.Put(7+i, 3, '-')
		if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{7, 3, 1 + i, 1}) {
			t.Fatalf("i=%d dirty = %+v", i, got)
		}
	}

	c.ClearDirtyRects()
	for j := 0; j < 10; j++ {
		c.Put(7, 3+j, '|')
		if got := c.DirtyRects(); len(got) != 1 || got[0] != (Rect{7, 3, 1, 1 + j}) {
			t.Fatalf("j=%d dirty = %+v", j, got)
		}
	}
}

func clone(c *Canvas) *Canvas {
	n := MustNew(c.Width(), c.Height())
	copy(n.Cells(), c.Cells())
	return n
}

func assertCells(t *testing.T, got, want *Canvas) {
	t.Helper()
	if got.Width() != want.Width() || got.Height() != want.Height() {
		t.Fatalf("size = %dx%d, want %dx%d", got.Width(), got.Height(), want.Width(), want.Height())
	}
	for i := range want.Cells() {
		if got.Cells()[i] != want.Cells()[i] {
			t.Fatalf("cell %d = %+v, want %+v", i, got.Cells()[i], want.Cells()[i])
		}
	}
}

func assertSize(t *testing.T, got, want *Canvas) {
	t.Helper()
	if got.Width() != want.Width() || got.Height() != want.Height() {
		t.Fatalf("size = %dx%d, want %dx%d", got.Width(), got.Height(), want.Width(), want.Height())
	}
}

func TestFrameDeleteCurrent(t *testing.T) {
	c := MustNew(2, 1)
	c.Put(0, 0, 'a')
	if err := c.CreateFrame(1); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFrame(1); err != nil {
		t.Fatal(err)
	}
	c.Put(0, 0, 'b')
	if err := c.FreeFrame(1); err != nil {
		t.Fatal(err)
	}
	if c.CurrentFrame() != 0 {
		t.Fatalf("current frame = %d, want 0", c.CurrentFrame())
	}
	if got := c.Get(0, 0); got != 'a' {
		t.Fatalf("cell = %q", got)
	}
}

func TestCreateFrameClamps(t *testing.T) {
	c := MustNew(1, 1)
	if err := c.CreateFrame(-5); err != nil {
		t.Fatal(err)
	}
	if c.CurrentFrame() != 1 {
		t.Fatalf("current frame = %d, want 1 after inserting at 0", c.CurrentFrame())
	}
	if err := c.CreateFrame(99); err != nil {
		t.Fatal(err)
	}
	if c.FrameCount() != 3 {
		t.Fatalf("frames = %d", c.FrameCount())
	}
}

func TestManageCanvas(t *testing.T) {
	c := MustNew(2, 2)
	if err := c.ManageCanvas(func() bool { return false }); err != nil {
		t.Fatal(err)
	}
	if err := c.Resize(3, 3); err != ErrManaged {
		t.Fatalf("resize error = %v, want ErrManaged", err)
	}
	if err := c.UnmanageCanvas(); err != nil {
		t.Fatal(err)
	}
	if err := c.Resize(3, 3); err != nil {
		t.Fatalf("resize after unmanage: %v", err)
	}
	if err := c.UnmanageCanvas(); err != ErrNotManaged {
		t.Fatalf("second unmanage = %v, want ErrNotManaged", err)
	}
}
