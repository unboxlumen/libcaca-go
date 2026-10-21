# nfilesart

`nfilesart` is a pure Go colour ASCII-art toolkit. It is a port of the
**core of [libcaca](https://github.com/cacalabs/libcaca)** — the `caca/`
library sources — with no cgo and no C dependency.

The port targets libcaca **0.99.beta20** (upstream commit `de48e1a`) and is
released as **v0.99.20**, keeping the version aligned with the C library it
mirrors. Output is byte-compatible with libcaca for the native, ANSI, UTF-8,
HTML, HTML3, BBCode, IRC, PostScript, SVG and troff formats; the dithering
pipeline reproduces the C implementation's choices exactly.

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

## Status

| Area | State |
|---|---|
| Character grid, attributes, frames, dirty rectangles | complete, byte-compatible |
| Drawing primitives and transforms | complete |
| Dithering (6 algorithms, 7 colour modes, 3 charsets) | complete, golden-tested |
| Import: `caca`, `text`, `ansi`, `utf8`, `bin`, auto-detect | complete |
| Export: `caca`, `ansi`, `utf8`, `utf8cr`, `html`, `html3`, `bbfr`, `irc`, `ps`, `svg`, `tga`, `troff` | complete |
| Bitmap fonts (libcaca `.data`), FIGlet/TOIlet FIGfonts | complete |
| Display backends | `tcell` (pure Go) and `null` |
| Platform drivers, `caca0`/`conio` compatibility layers | not ported |

Differences from the C library are intentional and documented: random
dithering uses a fixed seed so output is reproducible, and the TGA exporter
renders with Go Mono by default instead of libcaca's bundled bitmap font.

## Install

```sh
go get github.com/unboxlumen/libcaca-go@v0.99.20
```

Requires Go 1.26 or newer.

## Packages

| Package | Purpose |
|---|---|
| `canvas` | Character grid: cells, attributes, frames, dirty rectangles, primitives, transforms |
| `dither` | Bitmap → character cells: 6 algorithms, 7 colour modes, 3 glyph ramps |
| `codec` | Import/export in the formats listed above |
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
against hand-written expectations.

## License

**WTFPL v2** — see [LICENSE](LICENSE). Do what the fuck you want to.

All Go code here derives from libcaca's WTFPL-licensed `caca/` sources
(Copyright © Sam Hocevar and contributors, WTFPL v2). The GPL/LGPL parts of
the libcaca tree — `kernel/multiboot.h` (GPL), `src/aafire.c` and
`examples/demo0.c` (LGPL) — were **not** used. See
[PROVENANCE.md](PROVENANCE.md) for a per-package list of the reference
sources and for the licensing notes on the upstream tree.

Every Go file carries an attribution header naming the upstream file and
license.

## Relationship to other projects

- **libcaca** — the C original. nfilesart is an independent reimplementation,
  not a binding; it does not link against or vendor libcaca.
- **`gocaca/`** (in the parent repository) — an unrelated third-party Go port
  of libcaca by a different author. It is not libcaca and it is not part of
  nfilesart; the two share no code.
- **tcell** (MIT) and **golang.org/x/image** (BSD-3-Clause) are the only
  third-party Go dependencies.

## Package documentation

Run `go doc ./...`, or read the rendered docs at
<https://pkg.go.dev/github.com/unboxlumen/libcaca-go>. The umbrella package
comment in [doc.go](doc.go) gives the overall map.

## Versioning

Tags follow Go module semantics (`vX.Y.Z`). `v0.99.20` corresponds to
libcaca 0.99.beta20; patches to this port will increment the patch number
(`v0.99.21`, …) rather than moving to `1.0` until the display-driver surface
reaches parity with the C library.
