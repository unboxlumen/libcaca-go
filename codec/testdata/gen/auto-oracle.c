#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

static void dump(caca_canvas_t *cv, const char *name)
{
    size_t len = 0;
    void *buf = caca_export_canvas_to_memory(cv, "caca", &len);
    char path[256];
    snprintf(path, sizeof(path), "golden/%s", name);
    FILE *f = fopen(path, "wb");
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
}

int main(void)
{
    caca_canvas_t *cv;

    cv = caca_create_canvas(0, 0);
    caca_import_canvas_from_memory(cv, "hello", 5, "");
    dump(cv, "auto-text.caca");
    caca_free_canvas(cv);

    cv = caca_create_canvas(0, 0);
    caca_import_canvas_from_memory(cv, "\x1b[31mred", 8, "");
    dump(cv, "auto-ansi.caca");
    caca_free_canvas(cv);

    cv = caca_create_canvas(0, 0);
    {
        unsigned char bin[320];
        for(int j = 0; j < 320; j += 2) { bin[j] = ' '; bin[j+1] = 0x1f; }
        caca_import_canvas_from_memory(cv, bin, sizeof(bin), "");
    }
    dump(cv, "auto-bin.caca");
    caca_free_canvas(cv);

    cv = caca_create_canvas(6, 1);
    caca_set_color_ansi(cv, CACA_RED, CACA_BLUE);
    caca_put_str(cv, 0, 0, "native");
    size_t len = 0;
    void *blob = caca_export_canvas_to_memory(cv, "caca", &len);
    caca_free_canvas(cv);
    cv = caca_create_canvas(0, 0);
    caca_import_canvas_from_memory(cv, blob, len, "");
    free(blob);
    /* libcaca leaves the canvas' original frame in place during native
       import. Normalise it away so the fixture compares the imported frame. */
    if(caca_get_frame_count(cv) > 1)
    {
        caca_set_frame(cv, 0);
        caca_free_frame(cv, caca_get_frame_count(cv) - 1);
    }
    dump(cv, "auto-caca.caca");
    caca_free_canvas(cv);
    return 0;
}
