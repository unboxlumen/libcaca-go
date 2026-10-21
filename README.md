# nfilesart

`nfilesart` is a pure Go colour ASCII-art toolkit. It is a port of the
**core of [libcaca](https://github.com/cacalabs/libcaca)** — the
`caca/` library sources — with no cgo and no C dependency. Output is
byte-compatible with libcaca 0.99.beta20 for the native, ANSI, UTF-8, HTML,
HTML3, BBCode, IRC, PostScript, SVG and troff formats, and the dithering
pipeline reproduces the C implementation's choices.

```go
import (
    "github.com/unboxlumen/libcaca-go/canvas"
    "github.com/unboxlumen/libcaca-go/codec"
    "github.com/unboxlumen/libcaca-go/dither"
)

d, _ := dither.New(img, dither.DefaultOptions())
c, _ := canvas.New(80, 24)
_ = d.Draw(c, 0, 0, 80, 24)
ansi, _ := codec.Export(c, "ansi")
```

## License

**WTFPL v2** — see [LICENSE](LICENSE). Do what the fuck you want to.

All Go code here derives from libcaca's WTFPL-licensed `caca/` sources
(Copyright © Sam Hocevar and contributors, WTFPL v2). The GPL/LGPL parts of
the libcaca tree — `kernel/multiboot.h` (GPL), `src/aafire.c` and
`examples/demo0.c` (LGPL) — were **not** used. See
[PROVENANCE.md](PROVENANCE.md) for a per-package list of the reference
sources and for the licensing notes on the upstream tree.

Every Go file carries the attribution header:

```
nfilesart is a Go port of the core of libcaca, the WTFPL-licensed
caca/ library sources by Sam Hocevar ...
```

## Packages

| Package | Purpose |
|---|---|
| `canvas` | Character grid: cells, attributes, frames, dirty rectangles, primitives, transforms |
| `dither` | Bitmap → character cells: 6 algorithms, 7 colour modes, 3 glyph ramps |
| `codec` | Import/export: `caca`, `ansi`, `utf8`, `utf8cr`, `html`, `html3`, `bbfr`, `irc`, `ps`, `svg`, `tga`, `troff`, plus `text` and `bin` import |
| `font` | Go Mono by default, libcaca native bitmap fonts, FIGlet/TOIlet FIGfonts, canvas rendering |
| `charset` | Fullwidth classification, CP437, ASCII |
| `display` | Backend contract and the libcaca event/key model |
| `backend/tcell` | Pure Go terminal backend (keyboard, SGR mouse, resize) |
| `backend/null` | Headless backend |
| `preview` | Image → character art → bitmap, for GUI panes |

## Demos

Two runnable commands live under `cmd/`:

```sh
# Animated terminal scene. Press q / Ctrl-C to quit.
go run ./cmd/demo

# Image to character art. PNG, JPEG and GIF.
go run ./cmd/img2txt -W 100 -c blocks photo.jpg
go run ./cmd/img2txt -c shades -f ansi -o art.ansi photo.jpg
```

`cmd/img2txt` mirrors libcaca's `img2txt` flags: `-W`, `-H`, `-x`, `-y`,
`-c` (charset), `-color`, `-d` (algorithm), `-A` (antialias), `-g` (gamma),
`-f` (output format) and `-o` (output file).

Both commands degrade gracefully: when stdout is not a terminal, `demo`
prints one frame as UTF-8 and exits, and `img2txt` writes the requested
format to stdout or to `-o`.

## Testing

```sh
go test ./...
go test -cover ./...
go test -run Fuzz -fuzz FuzzImportCaca ./codec   # fuzzing targets
go test -bench . ./dither ./codec
```

83 golden fixtures are checked in under `*/testdata/golden/`. They were
produced by the C library using the small programs in `*/testdata/gen/`, so
the tests compare nfilesart's bytes against libcaca's own output rather than
against a hand-written expectation.

## Relationship to other projects

- **libcaca** — the C original. nfilesart is an independent reimplementation,
  not a binding; it does not link against or vendor libcaca.
- **`gocaca/`** (in the parent repository) — an unrelated third-party Go port
  of libcaca by a different author. It is not libcaca and it is not part of
  nfilesart; the two share no code.
- **tcell** (MIT) and **golang.org/x/image** (BSD-3-Clause) are the only
  third-party Go dependencies.

## Package documentation

Run `go doc ./...` or `go doc github.com/unboxlumen/libcaca-go/canvas`, or
read the rendered docs at <https://pkg.go.dev/github.com/unboxlumen/libcaca-go>.
The umbrella package comment in [doc.go](doc.go) gives the overall map.
