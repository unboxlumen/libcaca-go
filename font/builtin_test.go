// Copyright (c) 2026 The nfilesart Authors.
//
// nfilesart is a Go port of the core of libcaca, the WTFPL-licensed
// caca/ library sources by Sam Hocevar (libcaca 0.99.beta20,
// https://github.com/cacalabs/libcaca, commit de48e1a0888caabdc6276afdea46e9096d0adb05).
// Only the WTFPL parts of libcaca were used; see PROVENANCE.md.
//
// This program is free software. It comes without any warranty, to
// the extent permitted by applicable law. You can redistribute it and/or
// modify it under the terms of the Do What The Fuck You Want To Public
// License, Version 2, as published by Sam Hocevar. See LICENSE.

package font

import (
	"errors"
	"testing"
)

func TestBuiltinFonts(t *testing.T) {
	cases := []struct {
		name string
		w, h int
	}{
		{"Monospace 9", 7, 15},
		{"Monospace Bold 12", 10, 19},
		{"monospace 9", 7, 15},
	}
	for _, tc := range cases {
		f, err := LoadBuiltin(tc.name)
		if err != nil {
			t.Fatalf("LoadBuiltin(%q): %v", tc.name, err)
		}
		w, h := f.Metrics()
		if w != tc.w || h != tc.h {
			t.Errorf("LoadBuiltin(%q) metrics = %dx%d, want %dx%d", tc.name, w, h, tc.w, tc.h)
		}
	}
	if _, err := LoadBuiltin("no such font"); !errors.Is(err, ErrBuiltinFont) {
		t.Errorf("unknown font error = %v, want ErrBuiltinFont", err)
	}
}
