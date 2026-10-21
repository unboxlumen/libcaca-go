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

package codec

import (
	"fmt"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func benchCanvas() *canvas.Canvas {
	c := canvas.MustNew(80, 25)
	for y := 0; y < 25; y++ {
		if err := c.SetColor(canvas.Color(y%16), canvas.Color((y+3)%16)); err != nil {
			return c
		}
		for x := 0; x < 80; x++ {
			c.Put(x, y, rune('!'+((x+y)%90)))
		}
	}
	return c
}

func BenchmarkExportANSI(b *testing.B) {
	c := benchCanvas()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Export(c, "ansi"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExportUTF8(b *testing.B) {
	c := benchCanvas()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Export(c, "utf8"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExportCaca(b *testing.B) {
	c := benchCanvas()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Export(c, "caca"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImportCaca(b *testing.B) {
	c := benchCanvas()
	data, err := Export(c, "caca")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := canvas.MustNew(0, 0)
		if _, err := Import(dst, data, "caca"); err != nil {
			b.Fatal(err)
		}
	}
}

func ExampleExport() {
	c := canvas.MustNew(2, 1)
	c.Put(0, 0, 'A')
	c.Put(1, 0, 'B')
	out, _ := Export(c, "utf8")
	fmt.Print(string(out))
	// Output:
	// AB
}
