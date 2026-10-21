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

package font

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func FuzzParseFigfont(f *testing.F) {
	f.Add([]byte("flf2a$ 2 1 2 -1 0 0 0 0\n"))
	f.Add([]byte("flf2a$ 1 1 1 0 0 0 0 0\n  @\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseFigfont(bytes.NewReader(data))
	})
}

func FuzzLoadBitmapFont(f *testing.F) {
	for _, bpp := range []string{"1", "2", "4", "8"} {
		if b, err := os.ReadFile(filepath.Join("testdata", "fonts", "bpp"+bpp+".caf")); err == nil {
			f.Add(b)
		}
	}
	f.Add([]byte{0xca, 0xca, 'F', 'T'})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = LoadBitmapFont(data)
	})
}
