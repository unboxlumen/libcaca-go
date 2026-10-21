# ANSI import fixtures

`../golden/ansi-*.caca` and `../golden/utf8-wide.caca` are produced by
`caca_import_canvas_from_memory()` in C libcaca 0.99.beta20. They cover SGR
colours/bold/reverse/conceal, erase, cursor positioning, CR/tab/backspace,
DCH/ECH, OSC, SAUCE, CP437, the `\f\n` frame separator and UTF-8 fullwidth
input.

`ansi-oracle.c` contains the exact inputs. Compile it against a libcaca
checkout the same way as `oracle.c` and copy `golden/ansi-*.caca` plus
`golden/utf8-wide.caca` here.
