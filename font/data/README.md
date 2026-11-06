# Embedded bitmap fonts

`mono9.bin` and `monobold12.bin` are the byte arrays that
`caca/font.c` compiles in through `mono9.data` and `monobold12.data`.

- Source: `caca/mono9.data`, `caca/monobold12.data` at upstream commit
  `de48e1a0888caabdc6276afdea46e9096d0adb05` (libcaca 0.99.beta20).
- Generator: `tools/makefont.c`.
- Licence: WTFPL v2, the same as the rest of the `caca/` tree.

To regenerate, compile a tiny program that includes the upstream `.data`
file and writes `mono9_data`/`mono9_size` (or `monobold12_data`/
`monobold12_size`) to stdout, then redirect it here. The files are embedded
with `//go:embed` in `../builtin.go`.
