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
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func fuzzImport(t *testing.T, data []byte, format string) {
	t.Helper()
	c := canvas.MustNew(0, 0)
	_, _ = Import(c, data, format)
}

func FuzzImportCaca(f *testing.F) {
	f.Add([]byte{0xca, 0xca, 'C', 'V', 0, 0, 0, 48, 0, 0, 0, 8, 0, 1, 0, 0, 0, 1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "caca") })
}

func FuzzImportANSI(f *testing.F) {
	f.Add([]byte("\x1b[31mX"))
	f.Add([]byte("\x1b[2J\x1b[H hello"))
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "ansi") })
}

func FuzzImportUTF8(f *testing.F) {
	f.Add([]byte("hello\nworld"))
	f.Add([]byte("\x1b[1;34m蓝色\x1b[0m"))
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "utf8") })
}

func FuzzImportText(f *testing.F) {
	f.Add([]byte("hello\nworld"))
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "text") })
}

func FuzzImportBIN(f *testing.F) {
	f.Add(make([]byte, 320))
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "bin") })
}

func FuzzImportAuto(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte("\x1b[31mX"))
	f.Fuzz(func(t *testing.T, data []byte) { fuzzImport(t, data, "") })
}
