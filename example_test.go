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

package art_test

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/dither"
)

// Example draws on a canvas and round-trips it through libcaca's native
// format. The result is byte-compatible with what the C library would write.
func Example() {
	c := canvas.MustNew(4, 1)
	c.PutString(0, 0, "acdc")

	data, err := codec.Export(c, "caca")
	if err != nil {
		panic(err)
	}

	back := canvas.MustNew(0, 0)
	if _, err := codec.Import(back, data, "caca"); err != nil {
		panic(err)
	}

	var s strings.Builder
	for x := 0; x < back.Width(); x++ {
		s.WriteRune(back.Get(x, 0))
	}
	fmt.Println(s.String())
	// Output: acdc
}

// Example_dither shows the image pipeline: a bitmap is reduced to one
// attribute pair and one glyph per character cell, then exported as ANSI art.
func Example_dither() {
	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 32), uint8(y * 64), 128, 255})
		}
	}

	d, err := dither.New(img, dither.DefaultOptions())
	if err != nil {
		panic(err)
	}

	c := canvas.MustNew(8, 4)
	if err := d.Draw(c, 0, 0, 8, 4); err != nil {
		panic(err)
	}

	ansi, err := codec.Export(c, "ansi")
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(string(ansi), "\x1b["))
	fmt.Println(c.Width(), c.Height())
	// Output:
	// true
	// 8 4
}
