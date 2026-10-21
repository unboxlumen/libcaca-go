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

package dither

import (
	"image"
	"image/color"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func benchImage(w, h int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: uint8((x + y) * 255 / (w + h)),
				A: 255,
			})
		}
	}
	return img
}

func BenchmarkDither80x25(b *testing.B) {
	img := benchImage(640, 480)
	d, err := New(img, DefaultOptions())
	if err != nil {
		b.Fatal(err)
	}
	c := canvas.MustNew(80, 25)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Draw(c, 0, 0, 80, 25); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDitherNoPrefilter(b *testing.B) {
	img := benchImage(640, 480)
	opts := DefaultOptions()
	opts.Antialias = AntialiasNone
	d, err := New(img, opts)
	if err != nil {
		b.Fatal(err)
	}
	c := canvas.MustNew(80, 25)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Draw(c, 0, 0, 80, 25); err != nil {
			b.Fatal(err)
		}
	}
}
