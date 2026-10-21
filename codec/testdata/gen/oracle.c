#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

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

static void dump(caca_canvas_t *cv, char const *format, char const *name)
{
    size_t len = 0;
    void *buf = caca_export_canvas_to_memory(cv, format, &len);
    if(!buf) { fprintf(stderr, "export %s failed\n", format); exit(1); }
    char path[512];
    snprintf(path, sizeof(path), "golden/%s", name);
    FILE *f = fopen(path, "wb");
    if(!f) { perror(path); exit(1); }
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
}

static void fresh(caca_canvas_t **cv)
{
    *cv = caca_create_canvas(12, 4);
    if(!*cv) { fprintf(stderr, "canvas failed\n"); exit(1); }
    build(*cv);
}

int main(void)
{
    caca_canvas_t *cv;

    fresh(&cv);
    dump(cv, "ansi", "base.ansi");
    dump(cv, "utf8", "base.utf8");
    dump(cv, "utf8cr", "base.utf8cr");
    dump(cv, "html", "base.html");
    dump(cv, "html3", "base.html3");
    dump(cv, "bbfr", "base.bbfr");
    dump(cv, "irc", "base.irc");
    dump(cv, "ps", "base.ps");
    dump(cv, "svg", "base.svg");
    dump(cv, "troff", "base.troff");
    dump(cv, "caca", "base.caca");
    caca_free_canvas(cv);

    fresh(&cv); caca_flip(cv);        dump(cv, "caca", "flip.caca");        caca_free_canvas(cv);
    fresh(&cv); caca_flop(cv);        dump(cv, "caca", "flop.caca");        caca_free_canvas(cv);
    fresh(&cv); caca_rotate_180(cv);  dump(cv, "caca", "rot180.caca");      caca_free_canvas(cv);
    fresh(&cv); caca_rotate_left(cv); dump(cv, "caca", "rotleft.caca");     caca_free_canvas(cv);
    fresh(&cv); caca_rotate_right(cv);dump(cv, "caca", "rotright.caca");    caca_free_canvas(cv);
    fresh(&cv); caca_stretch_left(cv);dump(cv, "caca", "stretchleft.caca"); caca_free_canvas(cv);
    fresh(&cv); caca_stretch_right(cv);dump(cv,"caca", "stretchright.caca");caca_free_canvas(cv);

    fresh(&cv); caca_set_canvas_boundaries(cv, 3, 1, 7, 3);  dump(cv, "caca", "bounds-crop.caca");   caca_free_canvas(cv);
    fresh(&cv); caca_set_canvas_boundaries(cv, -2, -1, 16, 7);dump(cv, "caca", "bounds-expand.caca"); caca_free_canvas(cv);
    fresh(&cv); caca_set_canvas_boundaries(cv, 2, -2, 10, 8); dump(cv, "caca", "bounds-mixed.caca");  caca_free_canvas(cv);

    fresh(&cv);
    {
        char const *txt = "AB\nCD";
        caca_import_area_from_memory(cv, 2, 1, txt, strlen(txt), "text");
        dump(cv, "caca", "import-area.caca");
    }
    caca_free_canvas(cv);

    fresh(&cv);
    {
        size_t len = 0;
        void *buf = caca_export_area_to_memory(cv, 2, 1, 6, 3, "caca", &len);
        FILE *f = fopen("golden/area.caca", "wb");
        fwrite(buf, 1, len, f);
        fclose(f);
        free(buf);
    }
    caca_free_canvas(cv);
    return 0;
}
