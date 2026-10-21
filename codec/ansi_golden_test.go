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
	"bytes"
	"os"
	"testing"

	"github.com/unboxlumen/libcaca-go/canvas"
)

func TestGoldenANSIImport(t *testing.T) {
	cases := []struct {
		name, format string
		data         []byte
	}{
		{"ansi-sgr", "ansi", []byte("\x1b[31mred\x1b[0m normal")},
		{"ansi-bold", "ansi", []byte("\x1b[1;34mbold blue\x1b[0m")},
		{"ansi-reverse", "ansi", []byte("\x1b[7mreverse\x1b[0m")},
		{"ansi-conceal", "ansi", []byte("\x1b[8msecret\x1b[0m")},
		{"ansi-erase", "ansi", []byte("\x1b[2J\x1b[Hhome")},
		{"ansi-cup", "ansi", []byte("\x1b[5;10Hpos")},
		{"ansi-cr", "ansi", []byte("line1\nline2\rX")},
		{"ansi-tab", "ansi", []byte("a\tb\bc")},
		{"ansi-dch", "ansi", []byte("abcde\r\x1b[2P")},
		{"ansi-ech", "ansi", []byte("abcde\r\x1b[2X")},
		{"ansi-osc", "ansi", []byte("\x1b]0;title\aX")},
		{"ansi-sauce", "ansi", []byte("AB\x1aSAUCE00")},
		{"ansi-cp437", "ansi", []byte("\xb0\xb1\xb2")},
		{"ansi-frames", "ansi", []byte("A\f\nB")},
		{"utf8-wide", "utf8", []byte("\x1b[32m汉字\x1b[0m")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := canvas.MustNew(0, 0)
			if _, err := Import(c, tc.data, tc.format); err != nil {
				t.Fatal(err)
			}
			got, err := Export(c, "caca")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("testdata/golden/" + tc.name + ".caca")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs\n%s", tc.name, diff(got, want))
			}
		})
	}
}
