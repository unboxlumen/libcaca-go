#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

static void run(const char *name, const char *format, const unsigned char *data, size_t len)
{
    caca_canvas_t *cv = caca_create_canvas(0, 0);
    ssize_t r = caca_import_canvas_from_memory(cv, data, len, format);
    if(r < 0) { fprintf(stderr, "import failed: %s\n", name); exit(1); }
    size_t outlen = 0;
    void *buf = caca_export_canvas_to_memory(cv, "caca", &outlen);
    char path[256];
    snprintf(path, sizeof(path), "golden/%s", name);
    FILE *f = fopen(path, "wb");
    fwrite(buf, 1, outlen, f);
    fclose(f);
    free(buf);
    caca_free_canvas(cv);
}

#define RUN(name, fmt, lit) run(name, fmt, (const unsigned char *)lit, sizeof(lit)-1)

int main(void)
{
    RUN("ansi-sgr.caca", "ansi", "\x1b[31mred\x1b[0m normal");
    RUN("ansi-bold.caca", "ansi", "\x1b[1;34mbold blue\x1b[0m");
    RUN("ansi-reverse.caca", "ansi", "\x1b[7mreverse\x1b[0m");
    RUN("ansi-conceal.caca", "ansi", "\x1b[8msecret\x1b[0m");
    RUN("ansi-erase.caca", "ansi", "\x1b[2J\x1b[Hhome");
    RUN("ansi-cup.caca", "ansi", "\x1b[5;10Hpos");
    RUN("ansi-cr.caca", "ansi", "line1\nline2\rX");
    RUN("ansi-tab.caca", "ansi", "a\tb\bc");
    RUN("ansi-dch.caca", "ansi", "abcde\r\x1b[2P");
    RUN("ansi-ech.caca", "ansi", "abcde\r\x1b[2X");
    RUN("ansi-osc.caca", "ansi", "\x1b]0;title\aX");
    RUN("ansi-sauce.caca", "ansi", "AB\x1aSAUCE00");
    RUN("ansi-cp437.caca", "ansi", "\xb0\xb1\xb2");
    RUN("ansi-frames.caca", "ansi", "A\f\nB");
    RUN("utf8-wide.caca", "utf8", "\x1b[32m\xe6\xb1\x89\xe5\xad\x97\x1b[0m");
    return 0;
}
