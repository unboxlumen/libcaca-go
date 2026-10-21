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

// Command img2txt converts an image into colour character art. It mirrors
// libcaca's img2txt utility: the image is dithered onto a character canvas
// whose size is derived from the source aspect ratio, then exported.
//
// Usage:
//
//	img2txt [options] image.png
//	img2txt -W 100 -c blocks -f ansi -o art.ansi photo.jpg
//
// PNG, JPEG and GIF are supported. Without -o the result is written to
// stdout; formats other than "utf8" also work when stdout is a pipe.
package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/codec"
	"github.com/unboxlumen/libcaca-go/dither"
)

func main() {
	var (
		width     = flag.Int("W", 80, "output width in character cells")
		height    = flag.Int("H", 0, "output height in cells (0 = from aspect ratio)")
		xoffset   = flag.Int("x", 0, "left margin in cells")
		yoffset   = flag.Int("y", 0, "top margin in cells")
		charset   = flag.String("c", "ascii", "character set: ascii, shades or blocks")
		colorMode = flag.String("color", "full16", "colour mode: mono, gray, 8, 16, fullgray, full8, full16")
		algo      = flag.String("d", "fstein", "dithering: none, ordered2, ordered4, ordered8, random, fstein")
		antialias = flag.String("A", "prefilter", "antialiasing: none or prefilter")
		gamma     = flag.Float64("g", 1.0, "gamma correction")
		format    = flag.String("f", "utf8", "output format: "+strings.Join(codec.Formats(), ", "))
		output    = flag.String("o", "", "write to this file instead of stdout")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] image.png|jpg|gif\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), options{
		width: *width, height: *height, x: *xoffset, y: *yoffset,
		charset: *charset, colorMode: *colorMode, algo: *algo,
		antialias: *antialias, gamma: *gamma, format: *format, output: *output,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "img2txt:", err)
		os.Exit(1)
	}
}

type options struct {
	width, height int
	x, y          int
	charset       string
	colorMode     string
	algo          string
	antialias     string
	gamma         float64
	format        string
	output        string
}

func run(path string, o options) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	opts, err := ditherOptions(o)
	if err != nil {
		return err
	}
	d, err := dither.New(img, opts)
	if err != nil {
		return err
	}

	b := img.Bounds()
	dw, dh := o.width, o.height
	if dw <= 0 && dh <= 0 {
		dw = 80
	}
	if dw <= 0 {
		// Terminal cells are about twice as tall as they are wide.
		dw = dh * b.Dx() * 2 / maxInt(b.Dy(), 1)
	}
	if dh <= 0 {
		dh = dw * b.Dy() / maxInt(2*b.Dx(), 1)
	}
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}

	c, err := canvas.New(dw+2*o.x, dh+2*o.y)
	if err != nil {
		return err
	}
	if err := d.Draw(c, o.x, o.y, dw, dh); err != nil {
		return err
	}

	data, err := codec.Export(c, o.format)
	if err != nil {
		return err
	}
	if o.output != "" {
		return os.WriteFile(o.output, data, 0o644)
	}
	_, err = os.Stdout.Write(data)
	return err
}

func ditherOptions(o options) (dither.Options, error) {
	opts := dither.DefaultOptions()
	var err error
	if opts.Charset, err = dither.ParseCharset(o.charset); err != nil {
		return opts, err
	}
	if opts.Color, err = dither.ParseColor(o.colorMode); err != nil {
		return opts, err
	}
	if opts.Algorithm, err = dither.ParseAlgorithm(o.algo); err != nil {
		return opts, err
	}
	if opts.Antialias, err = dither.ParseAntialias(o.antialias); err != nil {
		return opts, err
	}
	opts.Gamma = o.gamma
	return opts, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
