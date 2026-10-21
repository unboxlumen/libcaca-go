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

// Package dither turns a bitmap into libcaca character cells.
//
// The pipeline and its option names follow libcaca 0.99.beta20: a source
// rectangle is averaged (or point-sampled), mapped to a 16-colour palette,
// then to a glyph from a small ramp, with optional error diffusion. The
// original C defaults are reproduced: prefilter antialiasing, full16 colours,
// ASCII glyphs and Floyd-Steinberg dithering.
package dither

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/unboxlumen/libcaca-go/canvas"
)

// Antialias selects how source pixels are reduced to one cell.
type Antialias int

const (
	AntialiasNone Antialias = iota
	AntialiasPrefilter
)

// ColorMode selects the destination colour model.
type ColorMode int

const (
	Mono ColorMode = iota
	Gray
	ANSI8
	ANSI16
	FullGray
	Full8
	Full16
)

// Charset selects the glyph ramp.
type Charset int

const (
	ASCII Charset = iota
	Shades
	Blocks
)

// Algorithm selects the dithering algorithm.
type Algorithm int

const (
	NoDither Algorithm = iota
	Ordered2
	Ordered4
	Ordered8
	Random
	FloydSteinberg
)

// Options configures a Dither.
type Options struct {
	Antialias  Antialias
	Color      ColorMode
	Charset    Charset
	Algorithm  Algorithm
	Gamma      float64
	Brightness float64
	Contrast   float64
	Invert     bool
	// Seed drives the random algorithm. Zero selects a fixed default seed, so
	// output is reproducible across runs (unlike libcaca's pid/ticks seed).
	Seed uint64
}

// DefaultOptions returns libcaca's defaults.
func DefaultOptions() Options {
	return Options{
		Antialias:  AntialiasPrefilter,
		Color:      Full16,
		Charset:    ASCII,
		Algorithm:  FloydSteinberg,
		Gamma:      1,
		Brightness: 1,
		Contrast:   1,
	}
}

// ParseAntialias parses libcaca's antialias names.
func ParseAntialias(s string) (Antialias, error) {
	switch s {
	case "none":
		return AntialiasNone, nil
	case "prefilter", "default":
		return AntialiasPrefilter, nil
	}
	return 0, fmt.Errorf("dither: unknown antialias %q", s)
}

// ParseColor parses libcaca's colour mode names.
func ParseColor(s string) (ColorMode, error) {
	switch s {
	case "mono":
		return Mono, nil
	case "gray":
		return Gray, nil
	case "8":
		return ANSI8, nil
	case "16":
		return ANSI16, nil
	case "fullgray":
		return FullGray, nil
	case "full8":
		return Full8, nil
	case "full16", "default":
		return Full16, nil
	}
	return 0, fmt.Errorf("dither: unknown colour mode %q", s)
}

// ParseCharset parses libcaca's glyph set names.
func ParseCharset(s string) (Charset, error) {
	switch s {
	case "ascii", "default":
		return ASCII, nil
	case "shades":
		return Shades, nil
	case "blocks":
		return Blocks, nil
	}
	return 0, fmt.Errorf("dither: unknown charset %q", s)
}

// ParseAlgorithm parses libcaca's algorithm names.
func ParseAlgorithm(s string) (Algorithm, error) {
	switch s {
	case "none":
		return NoDither, nil
	case "ordered2":
		return Ordered2, nil
	case "ordered4":
		return Ordered4, nil
	case "ordered8":
		return Ordered8, nil
	case "random":
		return Random, nil
	case "fstein", "default":
		return FloydSteinberg, nil
	}
	return 0, fmt.Errorf("dither: unknown algorithm %q", s)
}

// ErrEmpty is returned when the source image has no pixels.
var ErrEmpty = errors.New("dither: empty image")

// ErrInvalidFormat is returned for an invalid packed-pixel format.
var ErrInvalidFormat = errors.New("dither: invalid pixel format")

// ErrPalette is returned when a palette does not have 256 entries.
var ErrPalette = errors.New("dither: palette must have 256 entries")

type srcPixel struct {
	r, g, b, a int32
}

type source struct {
	w, h int
	pix  []srcPixel
}

// PixelFormat describes a packed pixel buffer, mirroring caca_create_dither.
type PixelFormat struct {
	BPP           int
	Width, Height int
	Pitch         int
	RMask, GMask  uint32
	BMask, AMask  uint32
}

// Dither converts a bitmap into canvas cells.
type Dither struct {
	src      source
	opts     Options
	glyphs   []rune
	gamma    [4096]int32
	rng      *rand.Rand
	hasAlpha bool

	// Packed-source state, retained so SetPalette can rebuild the cache.
	packed bool
	pf     PixelFormat
	raw    []byte
	rRight int
	rLeft  int
	gRight int
	gLeft  int
	bRight int
	bLeft  int
	aRight int
	aLeft  int
	palR   [256]int32
	palG   [256]int32
	palB   [256]int32
	palA   [256]int32
}

// New prepares a dither for img. The image is converted and gamma-corrected
// once, so Draw can be called repeatedly.
func New(img image.Image, opts Options) (*Dither, error) {
	if img == nil {
		return nil, ErrEmpty
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, ErrEmpty
	}

	d := &Dither{opts: opts, rng: newRNG(opts.Seed), hasAlpha: true}
	d.glyphs = glyphsFor(opts.Charset)
	d.buildGamma()

	d.src.w, d.src.h = w, h
	d.src.pix = make([]srcPixel, w*h)

	switch im := img.(type) {
	case *image.NRGBA:
		for y := 0; y < h; y++ {
			row := im.PixOffset(b.Min.X, b.Min.Y+y)
			for x := 0; x < w; x++ {
				i := row + x*4
				d.src.pix[y*w+x] = d.makePixel(im.Pix[i], im.Pix[i+1], im.Pix[i+2], im.Pix[i+3])
			}
		}
	case *image.RGBA:
		for y := 0; y < h; y++ {
			row := im.PixOffset(b.Min.X, b.Min.Y+y)
			for x := 0; x < w; x++ {
				i := row + x*4
				a := im.Pix[i+3]
				r, g, bb := im.Pix[i], im.Pix[i+1], im.Pix[i+2]
				if a != 0 && a != 0xff {
					r = uint8(uint32(r) * 0xff / uint32(a))
					g = uint8(uint32(g) * 0xff / uint32(a))
					bb = uint8(uint32(bb) * 0xff / uint32(a))
				}
				d.src.pix[y*w+x] = d.makePixel(r, g, bb, a)
			}
		}
	case *image.Gray:
		for y := 0; y < h; y++ {
			row := im.PixOffset(b.Min.X, b.Min.Y+y)
			for x := 0; x < w; x++ {
				v := im.Pix[row+x]
				d.src.pix[y*w+x] = d.makePixel(v, v, v, 0xff)
			}
		}
	default:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
				d.src.pix[y*w+x] = d.makePixel(n.R, n.G, n.B, n.A)
			}
		}
	}
	return d, nil
}

func newRNG(seed uint64) *rand.Rand {
	if seed == 0 {
		seed = 1
	}
	return rand.New(rand.NewPCG(seed, seed*0x9e3779b97f4a7c15+1))
}

func (d *Dither) makePixel(r, g, b, a uint8) srcPixel {
	return srcPixel{
		r: d.gamma[int(r)<<4],
		g: d.gamma[int(g)<<4],
		b: d.gamma[int(b)<<4],
		a: int32(a) << 4,
	}
}

func (d *Dither) buildGamma() {
	g := float32(d.opts.Gamma)
	if g < 0 {
		d.opts.Invert = true
		g = -g
	}
	if g == 0 {
		g = 1
	}
	for i := 0; i < 4096; i++ {
		if g == 1 {
			d.gamma[i] = int32(i)
			continue
		}
		d.gamma[i] = int32(4096.0 * gammaPow(float32(i)/4096.0, 1.0/g))
	}
}

// gammaPow mirrors libcaca's polynomial gammapow() when the x87 inline asm
// path is unavailable. Using float32 here is deliberate: the table is built
// from float operations in C, and matching the rounding makes the dither
// output byte-identical.
func gammaPow(x, y float32) float32 {
	if x == 0 {
		if y == 0 {
			return 1
		}
		return 0
	}
	t := (x - 1) / (x + 1)
	t2 := t * t
	r := t
	tmp := t
	for i := 3; i < 20; i += 2 {
		r *= t2
		tmp += r / float32(i)
	}
	tmp = -y * 2 * tmp
	r = tmp
	t = tmp
	tmp = 1 + t
	for i := 2; i < 16; i++ {
		r = r * t / float32(i)
		tmp += r
	}
	return 1 / tmp
}

func glyphsFor(cs Charset) []rune {
	switch cs {
	case Shades:
		return []rune{' ', 0xb7, 0x2591, 0x2592, '?'}
	case Blocks:
		return []rune{' ', 0x2598, 0x259a, '?'}
	default:
		return []rune{' ', '.', ':', ';', 't', '%', 'S', 'X', '@', '8', '?'}
	}
}

var rgbPalette = [16][3]int32{
	{0x0, 0x0, 0x0}, {0x0, 0x0, 0x7ff}, {0x0, 0x7ff, 0x0}, {0x0, 0x7ff, 0x7ff},
	{0x7ff, 0x0, 0x0}, {0x7ff, 0x0, 0x7ff}, {0x7ff, 0x7ff, 0x0}, {0xaaa, 0xaaa, 0xaaa},
	{0x555, 0x555, 0x555}, {0x000, 0x000, 0xfff}, {0x000, 0xfff, 0x000}, {0x000, 0xfff, 0xfff},
	{0xfff, 0x000, 0x000}, {0xfff, 0x000, 0xfff}, {0xfff, 0xfff, 0x000}, {0xfff, 0xfff, 0xfff},
}

var ordered2 = [...]int{0x00, 0x80, 0xc0, 0x40}

var ordered4 = [...]int{
	0x00, 0x80, 0x20, 0xa0,
	0xc0, 0x40, 0xe0, 0x60,
	0x30, 0xb0, 0x10, 0x90,
	0xf0, 0x70, 0xd0, 0x50,
}

var ordered8 = [...]int{
	0x00, 0x80, 0x20, 0xa0, 0x08, 0x88, 0x28, 0xa8,
	0xc0, 0x40, 0xe0, 0x60, 0xc8, 0x48, 0xe8, 0x68,
	0x30, 0xb0, 0x10, 0x90, 0x38, 0xb8, 0x18, 0x98,
	0xf0, 0x70, 0xd0, 0x50, 0xf8, 0x78, 0xd8, 0x58,
	0x0c, 0x8c, 0x2c, 0xac, 0x04, 0x84, 0x24, 0xa4,
	0xcc, 0x4c, 0xec, 0x6c, 0xc4, 0x44, 0xe4, 0x64,
	0x3c, 0xbc, 0x1c, 0x9c, 0x34, 0xb4, 0x14, 0x94,
	0xfc, 0x7c, 0xdc, 0x5c, 0xf4, 0x74, 0xd4, 0x54,
}

// Draw renders the source bitmap into the rectangle (x,y,w,h) of c.
func (d *Dither) Draw(c *canvas.Canvas, x, y, w, h int) error {
	if w <= 0 || h <= 0 || d.src.w == 0 || d.src.h == 0 {
		return nil
	}

	saved := c.CurAttr()
	defer c.SetAttr(saved)

	x1, x2 := x, x+w-1
	y1, y2 := y, y+h-1
	dstW, dstH := w, h
	srcW, srcH := d.src.w, d.src.h
	dchmax := len(d.glyphs)
	fullGray := d.opts.Color == FullGray
	full16 := d.opts.Color == Full16 || fullGray

	fsLen := x2 + 1
	if c.Width() <= x2 {
		fsLen = c.Width() + 1
	}
	if fsLen < 2 {
		fsLen = 2
	}
	fsR := make([]int32, fsLen+2)
	fsG := make([]int32, fsLen+2)
	fsB := make([]int32, fsLen+2)

	startY := max(y1, 0)
	endY := min(y2, c.Height()-1)
	startX := max(x1, 0)
	endX := min(x2, c.Width()-1)

	for yy := startY; yy <= endY; yy++ {
		var remainR, remainG, remainB int32
		for xx := startX; xx <= endX; xx++ {
			var r, g, b, a int32

			fromx := int(int64(xx-x1) * int64(srcW) / int64(dstW))
			fromy := int(int64(yy-y1) * int64(srcH) / int64(dstH))
			tox := int(int64(xx-x1+1) * int64(srcW) / int64(dstW))
			toy := int(int64(yy-y1+1) * int64(srcH) / int64(dstH))
			if tox == fromx {
				tox++
			}
			if toy == fromy {
				toy++
			}

			if d.opts.Antialias == AntialiasPrefilter {
				var dots int64
				for myx := max(fromx, 0); myx < min(tox, srcW); myx++ {
					for myy := max(fromy, 0); myy < min(toy, srcH); myy++ {
						p := d.src.pix[myy*srcW+myx]
						r += p.r
						g += p.g
						b += p.b
						a += p.a
						dots++
					}
				}
				if dots == 0 {
					dots = 1
				}
				r /= int32(dots)
				g /= int32(dots)
				b /= int32(dots)
				a /= int32(dots)
			} else {
				mx := (fromx + tox) / 2
				my := (fromy + toy) / 2
				mx = min(max(mx, 0), srcW-1)
				my = min(max(my, 0), srcH-1)
				p := d.src.pix[my*srcW+mx]
				r, g, b, a = p.r, p.g, p.b, p.a
			}

			if fullGray {
				gray := (3*r + 4*g + b + 4) / 8
				r, g, b = gray, gray, gray
			}

			if d.hasAlpha && a < 0x800 {
				remainR, remainG, remainB = 0, 0, 0
				fsR[xx+1] = 0
				fsG[xx+1] = 0
				fsB[xx+1] = 0
				continue
			}

			if d.opts.Algorithm == FloydSteinberg {
				r += remainR
				g += remainG
				b += remainB
			} else {
				n := int32(d.noise(xx, yy, startX)) - 0x80
				r += n * 4
				g += n * 4
				b += n * 4
			}

			// Nearest background colour.
			outbg, distmin := 0, int32(math.MaxInt32)
			for i := 0; i < 16; i++ {
				if fullGray && (rgbPalette[i][0] != rgbPalette[i][1] || rgbPalette[i][0] != rgbPalette[i][2]) {
					continue
				}
				dist := sq(r-rgbPalette[i][0]) + sq(g-rgbPalette[i][1]) + sq(b-rgbPalette[i][2])
				if dist < distmin {
					outbg = i
					distmin = dist
				}
			}
			bgR := rgbPalette[outbg][0]
			bgG := rgbPalette[outbg][1]
			bgB := rgbPalette[outbg][2]

			var outfg, ch int
			var outch rune
			var errR, errG, errB int32

			if full16 {
				outfg, distmin = 0, int32(math.MaxInt32)
				for i := 0; i < 16; i++ {
					if i == outbg {
						continue
					}
					if fullGray && (rgbPalette[i][0] != rgbPalette[i][1] || rgbPalette[i][0] != rgbPalette[i][2]) {
						continue
					}
					dist := sq(r-rgbPalette[i][0]) + sq(g-rgbPalette[i][1]) + sq(b-rgbPalette[i][2])
					if dist < distmin {
						outfg = i
						distmin = dist
					}
				}
				fgR := rgbPalette[outfg][0]
				fgG := rgbPalette[outfg][1]
				fgB := rgbPalette[outfg][2]

				distmin = int32(math.MaxInt32)
				den := int32(2*dchmax - 1)
				for i := 0; i < dchmax-1; i++ {
					newr := int32(i)*fgR + (den-int32(i))*bgR
					newg := int32(i)*fgG + (den-int32(i))*bgG
					newb := int32(i)*fgB + (den-int32(i))*bgB
					dist := abs32(r*den-newr) + abs32(g*den-newg) + abs32(b*den-newb)
					if dist < distmin {
						ch = i
						distmin = dist
					}
				}
				outch = d.glyphs[ch]
				if d.opts.Algorithm == FloydSteinberg {
					errR = r - (fgR*int32(ch)+bgR*(den-int32(ch)))/den
					errG = g - (fgG*int32(ch)+bgG*(den-int32(ch)))/den
					errB = b - (fgB*int32(ch)+bgB*(den-int32(ch)))/den
				}
			} else {
				lum := max32(r, max32(g, b))
				outfg = outbg
				outbg = int(canvas.Black)
				ch = int(lum * int32(dchmax) / 0x1000)
				if ch < 0 {
					ch = 0
				} else if ch > dchmax-1 {
					ch = dchmax - 1
				}
				outch = d.glyphs[ch]
				if d.opts.Algorithm == FloydSteinberg {
					errR = r - bgR*int32(ch)/int32(dchmax-1)
					errG = g - bgG*int32(ch)/int32(dchmax-1)
					errB = b - bgB*int32(ch)/int32(dchmax-1)
				}
			}

			if d.opts.Algorithm == FloydSteinberg {
				remainR = fsR[xx+2] + 7*errR/16
				remainG = fsG[xx+2] + 7*errG/16
				remainB = fsB[xx+2] + 7*errB/16
				fsR[xx] += 3 * errR / 16
				fsG[xx] += 3 * errG / 16
				fsB[xx] += 3 * errB / 16
				fsR[xx+1] = 5 * errR / 16
				fsG[xx+1] = 5 * errG / 16
				fsB[xx+1] = 5 * errB / 16
				fsR[xx+2] = errR / 16
				fsG[xx+2] = errG / 16
				fsB[xx+2] = errB / 16
			}

			if d.opts.Invert {
				outfg = 15 - outfg
				outbg = 15 - outbg
			}

			_ = c.SetColor(canvas.Color(outfg), canvas.Color(outbg))
			c.Put(xx, yy, outch)
		}
	}
	return nil
}

func (d *Dither) noise(x, y, startX int) int {
	switch d.opts.Algorithm {
	case Ordered2:
		return ordered2[(y%2)*2+(x-startX)%2]
	case Ordered4:
		return ordered4[(y%4)*4+(x-startX)%4]
	case Ordered8:
		return ordered8[(y%8)*8+(x-startX)%8]
	case Random:
		return d.rng.IntN(0x100)
	default:
		return 0x80
	}
}

func sq(v int32) int32 { return v * v }

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NewPixels prepares a dither from a packed pixel buffer. bpp must be 8, 16,
// 24 or 32; for 8bpp the palette defaults to grayscale and can be replaced
// with SetPalette. Masks follow caca_create_dither().
func NewPixels(pixels []byte, pf PixelFormat, opts Options) (*Dither, error) {
	if pf.BPP < 8 || pf.BPP > 32 || pf.BPP%8 != 0 {
		return nil, ErrInvalidFormat
	}
	if pf.Width < 0 || pf.Height < 0 || pf.Pitch < 0 {
		return nil, ErrInvalidFormat
	}
	bytesPer := pf.BPP / 8
	if pf.Pitch == 0 {
		pf.Pitch = bytesPer * pf.Width
	}
	if pf.Width > 0 && pf.Height > 0 {
		need := pf.Pitch*(pf.Height-1) + bytesPer*pf.Width
		if need > len(pixels) {
			return nil, ErrInvalidFormat
		}
	}

	d := &Dither{
		opts:     opts,
		rng:      newRNG(opts.Seed),
		packed:   true,
		pf:       pf,
		raw:      pixels,
		hasAlpha: pf.AMask != 0,
	}
	d.glyphs = glyphsFor(opts.Charset)
	d.buildGamma()

	if pf.BPP == 8 {
		for i := 0; i < 256; i++ {
			v := int32(i * 0xfff / 256)
			d.palR[i], d.palG[i], d.palB[i] = v, v, v
			d.palA[i] = 4095
		}
	} else {
		d.rRight, d.rLeft = maskShift(pf.RMask)
		d.gRight, d.gLeft = maskShift(pf.GMask)
		d.bRight, d.bLeft = maskShift(pf.BMask)
		d.aRight, d.aLeft = maskShift(pf.AMask)
	}
	d.fillPackedSource()
	return d, nil
}

// SetPalette installs a 256-entry palette. Values are in the 0..4095 range,
// as in caca_set_dither_palette().
func (d *Dither) SetPalette(red, green, blue, alpha []uint16) error {
	if !d.packed || d.pf.BPP != 8 {
		return ErrInvalidFormat
	}
	if len(red) < 256 || len(green) < 256 || len(blue) < 256 || len(alpha) < 256 {
		return ErrPalette
	}
	hasAlpha := false
	for i := 0; i < 256; i++ {
		if red[i]|green[i]|blue[i]|alpha[i] >= 0x1000 {
			return ErrInvalidFormat
		}
	}
	for i := 0; i < 256; i++ {
		d.palR[i] = int32(red[i])
		d.palG[i] = int32(green[i])
		d.palB[i] = int32(blue[i])
		d.palA[i] = int32(alpha[i])
		if alpha[i] != 0 {
			hasAlpha = true
		}
	}
	d.hasAlpha = hasAlpha
	d.fillPackedSource()
	return nil
}

func (d *Dither) fillPackedSource() {
	w, h := d.pf.Width, d.pf.Height
	d.src.w, d.src.h = w, h
	d.src.pix = make([]srcPixel, w*h)
	if w == 0 || h == 0 {
		return
	}
	bytesPer := d.pf.BPP / 8
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := y*d.pf.Pitch + x*bytesPer
			bits := readBits(d.raw, off, bytesPer)
			var p srcPixel
			if d.pf.BPP == 8 {
				idx := int(bits & 0xff)
				a := int32(4095)
				if d.hasAlpha {
					a = d.palA[idx]
				}
				p = srcPixel{r: d.palR[idx], g: d.palG[idx], b: d.palB[idx], a: a}
			} else {
				p.r = d.component(bits, d.pf.RMask, d.rRight, d.rLeft)
				p.g = d.component(bits, d.pf.GMask, d.gRight, d.gLeft)
				p.b = d.component(bits, d.pf.BMask, d.bRight, d.bLeft)
				if d.pf.AMask != 0 {
					p.a = int32(((bits & d.pf.AMask) >> d.aRight) << d.aLeft)
				}
			}
			d.src.pix[y*w+x] = p
		}
	}
}

func (d *Dither) component(bits, mask uint32, right, left int) int32 {
	if mask == 0 {
		return 0
	}
	idx := int(((bits & mask) >> right) << left)
	if idx < 0 {
		idx = 0
	} else if idx > 4095 {
		idx = 4095
	}
	return d.gamma[idx]
}

func readBits(data []byte, off, n int) uint32 {
	switch n {
	case 1:
		return uint32(data[off])
	case 2:
		return uint32(binary.LittleEndian.Uint16(data[off : off+2]))
	case 3:
		return uint32(data[off]) | uint32(data[off+1])<<8 | uint32(data[off+2])<<16
	case 4:
		return binary.LittleEndian.Uint32(data[off : off+4])
	}
	return 0
}

// maskShift converts a color mask to the same (right, left) shifts used by
// libcaca: right is the trailing-zero count, left is 12 minus the bit count.
func maskShift(mask uint32) (right, left int) {
	if mask == 0 {
		return 0, 0
	}
	for mask&1 == 0 {
		mask >>= 1
		right++
	}
	ones := 0
	for mask&1 != 0 {
		mask >>= 1
		ones++
	}
	return right, 12 - ones
}
