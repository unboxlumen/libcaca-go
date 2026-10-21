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
	"encoding/binary"
	"image"
	"image/draw"

	"github.com/unboxlumen/libcaca-go/canvas"
	"github.com/unboxlumen/libcaca-go/font"
)

// exportTGA renders the canvas with the default Go Mono font into an
// uncompressed 32-bit TGA image.
func exportTGA(v view, face font.Font) ([]byte, error) {
	f := face
	if f == nil {
		var err error
		f, err = font.Default()
		if err != nil {
			return nil, err
		}
	}
	fw, fh := f.Metrics()
	imgW, imgH := v.w*fw, v.h*fh
	if imgW <= 0 || imgH <= 0 {
		return nil, ErrInvalid
	}

	tmp := canvas.MustNew(v.w, v.h)
	for y := 0; y < v.h; y++ {
		for x := 0; x < v.w; x++ {
			tmp.SetCell(x, y, v.cells[y*v.w+x])
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)
	f.Render(tmp, img)

	var b bytes.Buffer
	header := make([]byte, 18)
	header[2] = 2 // uncompressed truecolour
	binary.LittleEndian.PutUint16(header[12:14], uint16(imgW))
	binary.LittleEndian.PutUint16(header[14:16], uint16(imgH))
	header[16] = 32
	header[17] = 0x28 // top-left origin, 8 alpha bits
	b.Write(header)

	pix := make([]byte, imgW*imgH*4)
	for i := 0; i < imgW*imgH; i++ {
		r := img.Pix[i*4]
		g := img.Pix[i*4+1]
		bl := img.Pix[i*4+2]
		a := img.Pix[i*4+3]
		pix[i*4] = bl
		pix[i*4+1] = g
		pix[i*4+2] = r
		pix[i*4+3] = a
	}
	b.Write(pix)
	return b.Bytes(), nil
}
