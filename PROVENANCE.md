# Provenance

nfilesart (`github.com/unboxlumen/libcaca-go`) is released under the **WTFPL
v2** (see [LICENSE](LICENSE)).

It is a Go port of the **core** of libcaca — the `caca/` library sources at
<https://github.com/cacalabs/libcaca>, reference commit
`de48e1a0888caabdc6276afdea46e9096d0adb05` (after `v0.99.beta20`, the version
this port targets).

Every line of upstream code that was used comes from files carrying the
WTFPL v2 header. Nothing else from the upstream tree was copied.

## Ported from libcaca (all WTFPL v2)

| nfilesart package | Reference C sources |
|---|---|
| `canvas/` | `caca/attr.c`, `caca/canvas.c`, `caca/string.c`, `caca/frame.c`, `caca/dirty.c`, `caca/line.c`, `caca/box.c`, `caca/conic.c`, `caca/triangle.c`, `caca/transform.c` |
| `codec/` | `caca/codec/export.c`, `caca/codec/import.c`, `caca/codec/text.c` |
| `dither/` | `caca/dither.c` |
| `font/` | `caca/font.c`, `caca/figfont.c` |
| `charset/` | `caca/charset.c` |
| `display/`, `backend/` | the driver interface of `caca/graphics.c` and `caca/event.c`; the `null` backend mirrors `caca/driver/null.c` |

The following data was transcribed from those files: the attribute encoding
and the 16/14-bit ANSI palettes, the dithering RGB palette and the ordered
dither matrices, the glyph ramps, the fullwidth character ranges, the CP437
table, and the character transform tables.

`font/figfont.go` additionally carries over the layout rules that
`caca/figfont.c` itself marks as coming from TOIlet (`main.c`, `render.c`,
`figlet.c`). TOIlet is Sam Hocevar's WTFPL project, so no new license terms
are introduced.

## Deliberately excluded

These parts of the upstream tree are **not** used by nfilesart:

- `kernel/multiboot.h` — GPL-2.0-or-later
- `src/aafire.c` — LGPL-2.0-or-later
- `examples/demo0.c` — LGPL-2.0-or-later
- `examples/conio.c`, `examples/conio-snake.cpp` — unrelated demo code
  (the former has no license header, the latter is ISC)
- `caca/mono9.data`, `caca/monobold12.data` — generated font blobs with no
  stated license. The default font is Go Mono (`golang.org/x/image/font/gofont`),
  which is BSD-3-Clause. The `.caf` and `.flf` files under `font/testdata/`
  are hand-written synthetic fixtures, not upstream font data.
- the platform drivers (`x11.c`, `gl.c`, `ncurses.c`, `slang.c`, `cocoa.m`,
  `win32.c`, `conio.c`, `vga.c`, `raw.c`) and the `caca0`/`caca_conio`
  compatibility layers — their behaviour is replaced by the `display.Backend`
  interface and the `tcell` backend.

## Upstream licensing note

Upstream `COPYING` is the WTFPL v2, and every file under `caca/` carries the
WTFPL header. Two other upstream artifacts say otherwise but concern code
that is not part of the core: `libcaca.spec` still says `License: GPL` (stale
metadata from the 0.9 era), and `doc/cacafire.1` states that the `cacafire`
demo is LGPL. Neither affects the `caca/` sources used here.

## Not related to nfilesart

`gocaca/` in the parent repository is a **separate third-party Go project**,
not part of libcaca and not part of nfilesart. It shares the idea of porting
libcaca to Go but was written independently. nfilesart does not import,
vendor or otherwise depend on it.
