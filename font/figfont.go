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
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// Figfont is a loaded FIGlet/TOIlet font.
type Figfont struct {
	Hardblank  rune
	Height     int
	Baseline   int
	MaxLength  int
	OldLayout  int
	FullLayout int
	PrintDir   int
	Glyphs     map[rune][]string
	// TermWidth is the wrap column used by Render. The zero value disables
	// wrapping; C's caca_canvas_set_figfont defaults to 80.
	TermWidth int

	hmode      hmode
	hsmushRule int
}

type hmode int

const (
	hDefault hmode = iota
	hKern
	hSmush
	hNone
	hOverlap
)

// ParseFigfont reads a .flf font.
func ParseFigfont(r io.Reader) (*Figfont, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	if !sc.Scan() {
		return nil, fmt.Errorf("figfont: empty input")
	}
	header := strings.TrimRight(sc.Text(), "\r")
	fields := strings.Fields(header)
	if len(fields) < 6 || !strings.HasPrefix(fields[0], "flf2a") || len(fields[0]) < 6 {
		return nil, fmt.Errorf("figfont: bad signature")
	}
	f := &Figfont{Hardblank: rune(fields[0][5]), Glyphs: make(map[rune][]string), TermWidth: 80}
	var err error
	if f.Height, err = atoiField(fields, 1); err != nil {
		return nil, err
	}
	if f.Baseline, err = atoiField(fields, 2); err != nil {
		return nil, err
	}
	if f.MaxLength, err = atoiField(fields, 3); err != nil {
		return nil, err
	}
	if f.OldLayout, err = atoiField(fields, 4); err != nil {
		return nil, err
	}
	commentLines, err := atoiField(fields, 5)
	if err != nil {
		return nil, err
	}
	if len(fields) > 6 {
		f.PrintDir, _ = strconv.Atoi(fields[6])
	}
	if len(fields) > 7 {
		f.FullLayout, _ = strconv.Atoi(fields[7])
	}
	if f.Height <= 0 || f.Height > 1024 || f.MaxLength < 0 || f.MaxLength > 4096 ||
		commentLines < 0 || commentLines > 1<<20 {
		return nil, fmt.Errorf("figfont: invalid header dimensions")
	}
	f.updateLayout()

	for i := 0; i < commentLines; i++ {
		if !sc.Scan() {
			return nil, fmt.Errorf("figfont: truncated comments")
		}
	}

	readGlyph := func() ([]string, error) {
		rows := make([]string, f.Height)
		var endmark rune
		for i := 0; i < f.Height; i++ {
			if !sc.Scan() {
				return nil, fmt.Errorf("figfont: truncated glyph")
			}
			line := strings.TrimRight(sc.Text(), "\r")
			if i == 0 {
				rs := []rune(line)
				if len(rs) == 0 {
					return nil, fmt.Errorf("figfont: empty glyph row")
				}
				endmark = rs[len(rs)-1]
			}
			line = strings.TrimRight(line, string(endmark))
			// libcaca normalises the hardblank to U+00A0 so the smushing
			// rules see a single sentinel.
			if f.Hardblank != 0 && strings.ContainsRune(line, f.Hardblank) {
				line = strings.Map(func(r rune) rune {
					if r == f.Hardblank {
						return 0xa0
					}
					return r
				}, line)
			}
			rows[i] = line
		}
		return rows, nil
	}

	for ch := rune(32); ch <= 126; ch++ {
		rows, err := readGlyph()
		if err != nil {
			return nil, err
		}
		f.Glyphs[ch] = rows
	}

	// Classic fonts carry these seven German characters after ASCII.
	for _, ch := range []rune{196, 214, 220, 228, 246, 252, 223} {
		rows, err := readGlyph()
		if err != nil {
			break
		}
		f.Glyphs[ch] = rows
	}

	// Code-tagged glyphs.
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var code rune
		if strings.HasPrefix(line, "0x") || strings.HasPrefix(line, "0X") {
			v, err := strconv.ParseInt(line[2:], 16, 32)
			if err != nil {
				break
			}
			code = rune(v)
		} else {
			v, err := strconv.Atoi(line)
			if err != nil {
				break
			}
			code = rune(v)
		}
		rows, err := readGlyph()
		if err != nil {
			break
		}
		f.Glyphs[code] = rows
	}
	return f, nil
}

// LoadFigfont reads a .flf file from disk.
func LoadFigfont(path string) (*Figfont, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return ParseFigfont(fh)
}

func atoiField(fields []string, i int) (int, error) {
	if i >= len(fields) {
		return 0, fmt.Errorf("figfont: missing header field %d", i)
	}
	v, err := strconv.Atoi(fields[i])
	if err != nil {
		return 0, fmt.Errorf("figfont: header field %d: %w", i, err)
	}
	return v, nil
}

func (f *Figfont) updateLayout() {
	if f.FullLayout&0x3f != 0 {
		f.hsmushRule = f.FullLayout & 0x3f
	} else if f.OldLayout > 0 {
		f.hsmushRule = f.OldLayout
	}

	switch {
	case f.OldLayout == -1:
		f.hmode = hNone
	case f.OldLayout == 0 && f.FullLayout&0xc0 == 0x40:
		f.hmode = hKern
	case f.OldLayout&0x3f != 0 && f.FullLayout&0x3f != 0 && f.FullLayout&0x80 != 0:
		f.hmode = hSmush
		f.hsmushRule = f.FullLayout & 0x3f
	case f.OldLayout == 0 && f.FullLayout&0xbf == 0x80:
		f.hmode = hSmush
		f.hsmushRule = 0x3f
	default:
		f.hmode = hOverlap
	}
}

// SetWidth changes the wrap column used by Render. A zero or negative value
// disables wrapping.
func (f *Figfont) SetWidth(w int) { f.TermWidth = w }

// SetSmush overrides the layout mode: "default" recomputes it from the font
// header; "kern", "smush", "none" and "overlap" force that mode. Unknown
// values fall back to "default", like caca_set_figfont_smush.
func (f *Figfont) SetSmush(mode string) {
	f.updateLayout()
	switch strings.ToLower(mode) {
	case "kern":
		f.hmode = hKern
	case "smush":
		f.hmode = hSmush
	case "none":
		f.hmode = hNone
	case "overlap":
		f.hmode = hOverlap
	}
}

// Render lays out text and returns the resulting rows.
func (f *Figfont) Render(text string) []string {
	return f.renderRows(text)
}

// Draw renders text into c at (x,y) using the canvas' current attribute.
func (f *Figfont) Draw(c *canvas.Canvas, x, y int, text string) {
	for j, row := range f.renderRows(text) {
		for i, ch := range row {
			if ch == 0xa0 {
				ch = ' '
			}
			c.Put(x+i, y+j, ch)
		}
	}
}

func (f *Figfont) renderRows(text string) []string {
	if f.Height <= 0 {
		return nil
	}
	rows := make([][]rune, f.Height)
	x, y := 0, 0

	for _, ch := range text {
		switch ch {
		case '\r':
			continue
		case '\n':
			x = 0
			y += f.Height
			rows = ensureRows(rows, y+f.Height)
			continue
		}

		glyph, ok := f.Glyphs[ch]
		if !ok {
			glyph = f.Glyphs['?']
			if glyph == nil {
				continue
			}
		}
		w := glyphWidth(glyph)
		if w == 0 {
			continue
		}

		// C wraps only when a glyph is already on the line.
		if f.TermWidth > 0 && x != 0 && x+w > f.TermWidth {
			x = 0
			y += f.Height
			rows = ensureRows(rows, y+f.Height)
		}

		overlap := f.overlap(rows, y, x, glyph, w)
		x = f.renderGlyph(rows, y, x, glyph, w, overlap)
	}

	maxW := 0
	for j := range rows {
		if len(rows[j]) > maxW {
			maxW = len(rows[j])
		}
	}
	out := make([]string, len(rows))
	for j := range rows {
		if len(rows[j]) < maxW {
			rows[j] = append(rows[j], make([]rune, maxW-len(rows[j]))...)
		}
		for i, ch := range rows[j] {
			if ch == 0xa0 {
				rows[j][i] = ' '
			}
		}
		out[j] = string(rows[j])
	}
	return out
}

func glyphWidth(glyph []string) int {
	w := 0
	for _, row := range glyph {
		if n := len([]rune(row)); n > w {
			w = n
		}
	}
	return w
}

// overlap is a direct port of libcaca's per-row overlap computation.
func ensureRows(rows [][]rune, n int) [][]rune {
	for len(rows) < n {
		rows = append(rows, nil)
	}
	return rows
}

func (f *Figfont) overlap(rows [][]rune, y, x int, glyph []string, w int) int {
	if f.hmode == hNone {
		return 0
	}
	overlap := w
	for j := 0; j < len(glyph) && y+j < len(rows); j++ {
		g := []rune(glyph[j])
		xright := 0
		for xright < overlap && xright < len(g) && g[xright] == ' ' {
			xright++
		}
		xleft := 0
		for xright+xleft < overlap && xleft < x {
			if rows[y+j][x-1-xleft] != ' ' {
				break
			}
			xleft++
		}
		if f.hmode == hOverlap && xleft < x {
			xleft++
		}
		if f.hmode == hSmush && xleft < x {
			left := rows[y+j][x-1-xleft]
			right := runeAt(g, xright)
			if hsmush(left, right, f.hsmushRule) != 0 {
				xleft++
			}
		}
		if xleft+xright < overlap {
			overlap = xleft + xright
		}
	}
	return overlap
}

func (f *Figfont) renderGlyph(rows [][]rune, y, x int, glyph []string, w, overlap int) int {
	newX := x + w - overlap
	if newX < 0 {
		newX = 0
	}
	for j := 0; j < len(glyph) && y+j < len(rows); j++ {
		if len(rows[y+j]) < x+w {
			rows[y+j] = append(rows[y+j], make([]rune, x+w-len(rows[y+j]))...)
		}
		g := []rune(glyph[j])
		for i := 0; i < w; i++ {
			ch2 := runeAt(g, i)
			if ch2 == ' ' {
				continue
			}
			col := x - overlap + i
			if col < 0 || col >= newX {
				continue
			}
			ch1 := rows[y+j][col]
			if ch1 == ' ' || f.hmode != hSmush {
				rows[y+j][col] = ch2
			} else {
				rows[y+j][col] = hsmush(ch1, ch2, f.hsmushRule)
			}
		}
		if len(rows[y+j]) > newX {
			rows[y+j] = rows[y+j][:newX]
		} else if len(rows[y+j]) < newX {
			rows[y+j] = append(rows[y+j], make([]rune, newX-len(rows[y+j]))...)
		}
	}
	return newX
}

func runeAt(s []rune, i int) rune {
	if i < 0 || i >= len(s) {
		return ' '
	}
	return s[i]
}

// hsmush is a direct port of libcaca's horizontal smushing rules.
func hsmush(ch1, ch2 rune, rule int) rune {
	if rule&0x01 != 0 && ch1 == ch2 && ch1 != 0xa0 {
		return ch2
	}
	if ch1 < 0x80 && ch2 < 0x80 {
		const charlist = "|/\\[]{}()<>"

		if rule&0x02 != 0 {
			if ch1 == '_' && strings.ContainsRune(charlist, ch2) {
				return ch2
			}
			if ch2 == '_' && strings.ContainsRune(charlist, ch1) {
				return ch1
			}
		}
		if rule&0x04 != 0 {
			i1 := strings.IndexRune(charlist, ch1)
			i2 := strings.IndexRune(charlist, ch2)
			if i1 >= 0 && i2 >= 0 {
				cl1 := (i1 + 1) / 2
				cl2 := (i2 + 1) / 2
				if cl1 < cl2 {
					return ch2
				}
				if cl1 > cl2 {
					return ch1
				}
			}
		}
		if rule&0x08 != 0 {
			s := uint16(ch1) + uint16(ch2)
			p := uint16(ch1) * uint16(ch2)
			if p == 15375 || p == 8463 || (p == 1640 && s == 81) {
				return '|'
			}
		}
		if rule&0x10 != 0 {
			switch (uint16(ch1) << 8) | uint16(ch2) {
			case 0x2f5c:
				return '|'
			case 0x5c2f:
				return 'Y'
			case 0x3e3c:
				return 'X'
			}
		}
		if rule&0x20 != 0 && ch1 == ch2 && ch1 == 0xa0 {
			return 0xa0
		}
	}
	return 0
}
