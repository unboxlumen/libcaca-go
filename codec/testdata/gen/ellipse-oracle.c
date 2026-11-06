#include <stdio.h>
#include <stdlib.h>
#include "caca.h"

/* Exercises the thin and filled ellipse rasterisers, which share the same
 * Bresenham second-phase initialisation in caca/conic.c. */
int main(int argc, char **argv)
{
    const char *dir = argc > 1 ? argv[1] : "golden";
    caca_canvas_t *cv = caca_create_canvas(20, 10);
    size_t len = 0;
    void *buf;
    char path[512];
    FILE *f;

    if(!cv) { fprintf(stderr, "canvas failed\n"); return 1; }
    caca_set_color_ansi(cv, CACA_LIGHTGRAY, CACA_BLUE);
    caca_fill_box(cv, 0, 0, 20, 10, '.');
    caca_set_color_ansi(cv, CACA_YELLOW, CACA_BLACK);
    caca_draw_thin_ellipse(cv, 5, 5, 4, 3);
    caca_set_color_ansi(cv, CACA_RED, CACA_DARKGRAY);
    caca_fill_ellipse(cv, 14, 5, 4, 3, '@');

    buf = caca_export_canvas_to_memory(cv, "caca", &len);
    if(!buf) { fprintf(stderr, "export caca failed\n"); return 1; }
    snprintf(path, sizeof(path), "%s/ellipse.caca", dir);
    f = fopen(path, "wb");
    if(!f) { perror(path); return 1; }
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
    caca_free_canvas(cv);
    return 0;
}
