#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

#define W 32
#define H 32

static unsigned char pixels[W * H * 3];

static void make(void)
{
    int x, y;
    for(y = 0; y < H; y++)
        for(x = 0; x < W; x++)
        {
            /* high-frequency black/white + saturated primaries: this drives
               Floyd-Steinberg errors far from zero in both directions. */
            unsigned v = ((x ^ y) & 1) ? 255 : 0;
            unsigned r = v, g = v, b = v;
            if((x % 5) == 0) { r = 255; g = 0; b = 0; }
            if((y % 5) == 0) { r = 0; g = 0; b = 255; }
            if((x + y) % 7 == 0) { r = 0; g = 255; b = 0; }
            pixels[(y * W + x) * 3 + 0] = (unsigned char)b;
            pixels[(y * W + x) * 3 + 1] = (unsigned char)g;
            pixels[(y * W + x) * 3 + 2] = (unsigned char)r;
        }
}

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

static void run(const char *name, const char *algo, const char *charset,
                const char *color, const char *aa)
{
    caca_canvas_t *cv = caca_create_canvas(W, H / 2);
    caca_dither_t *d = caca_create_dither(24, W, H, W * 3,
                                          0xff0000, 0x00ff00, 0x0000ff, 0);
    caca_set_dither_algorithm(d, algo);
    caca_set_dither_charset(d, charset);
    caca_set_dither_color(d, color);
    caca_set_dither_antialias(d, aa);
    caca_dither_bitmap(cv, 0, 0, W, H / 2, d, pixels);
    dump(cv, name);
    caca_free_dither(d);
    caca_free_canvas(cv);
}

int main(void)
{
    make();
    run("extreme-fstein.caca", "fstein",   "ascii",  "full16", "prefilter");
    run("extreme-noaa.caca",   "fstein",   "ascii",  "full16", "none");
    run("extreme-ordered4.caca","ordered4","ascii",  "full16", "prefilter");
    run("extreme-shades.caca", "fstein",   "shades", "full16", "prefilter");
    run("extreme-fullgray.caca","fstein",  "ascii",  "fullgray","prefilter");
    return 0;
}
