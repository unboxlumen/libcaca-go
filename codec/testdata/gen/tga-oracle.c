#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

/* Same scene as oracle.c's build(), so the TGA golden exercises the default
 * bitmap font together with colours, styles, fullwidth and box drawing. */
static void build(caca_canvas_t *cv)
{
    caca_set_color_ansi(cv, CACA_LIGHTGRAY, CACA_BLACK);
    caca_fill_box(cv, 0, 0, 12, 4, ' ');
    caca_put_str(cv, 0, 0, "libcaca");
    caca_set_color_ansi(cv, CACA_RED, CACA_BLUE);
    caca_put_str(cv, 0, 1, "Go port");
    caca_set_color_ansi(cv, CACA_YELLOW, CACA_DARKGRAY);
    caca_put_char(cv, 9, 2, 0x2f06); /* fullwidth ⼆ */
    caca_set_attr(cv, CACA_BOLD | CACA_UNDERLINE);
    caca_put_str(cv, 0, 3, "styles");
    caca_draw_box(cv, 7, 0, 5, 4, '#');
}

int main(int argc, char **argv)
{
    const char *dir = argc > 1 ? argv[1] : "golden";
    caca_canvas_t *cv = caca_create_canvas(12, 4);
    size_t len = 0;
    void *buf;
    char path[512];
    FILE *f;

    if(!cv) { fprintf(stderr, "canvas failed\n"); return 1; }
    build(cv);

    buf = caca_export_canvas_to_memory(cv, "tga", &len);
    if(!buf) { fprintf(stderr, "export tga failed\n"); return 1; }
    snprintf(path, sizeof(path), "%s/base.tga", dir);
    f = fopen(path, "wb");
    if(!f) { perror(path); return 1; }
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
    caca_free_canvas(cv);
    return 0;
}
