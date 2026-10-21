#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "caca.h"

#define W 16
#define H 16

static unsigned char pixels[W * H * 3];

static void make_pixels(void)
{
    int x, y;
    for(y = 0; y < H; y++)
        for(x = 0; x < W; x++)
        {
            unsigned char r = (unsigned char)(x * 16);
            unsigned char g = (unsigned char)(y * 16);
            unsigned char b = (unsigned char)((x * 37 + y * 53) & 0xff);
            /* libcaca's 24bpp reader treats the buffer as BGR on LE. */
            pixels[(y * W + x) * 3 + 0] = b;
            pixels[(y * W + x) * 3 + 1] = g;
            pixels[(y * W + x) * 3 + 2] = r;
        }
}

static void dump(caca_canvas_t *cv, const char *name)
{
    size_t len = 0;
    void *buf = caca_export_canvas_to_memory(cv, "caca", &len);
    char path[256];
    snprintf(path, sizeof(path), "golden/dither-%s.caca", name);
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

static void run_gamma(const char *name, float gamma)
{
    caca_canvas_t *cv = caca_create_canvas(W, H / 2);
    caca_dither_t *d = caca_create_dither(24, W, H, W * 3,
                                          0xff0000, 0x00ff00, 0x0000ff, 0);
    caca_set_dither_gamma(d, gamma);
    caca_dither_bitmap(cv, 0, 0, W, H / 2, d, pixels);
    dump(cv, name);
    caca_free_dither(d);
    caca_free_canvas(cv);
}

static void run_pix8(const char *name, int custom)
{
    unsigned char p[W * H];
    int x, y;
    for(y = 0; y < H; y++)
        for(x = 0; x < W; x++)
            p[y * W + x] = (unsigned char)((x * 13 + y * 7) & 0xff);

    caca_canvas_t *cv = caca_create_canvas(W, H / 2);
    caca_dither_t *d = caca_create_dither(8, W, H, W, 0, 0, 0, 0);
    if(custom)
    {
        uint32_t r[256], g[256], b[256], a[256];
        for(int i = 0; i < 256; i++)
        {
            r[i] = (uint32_t)((i * 7) & 0xfff);
            g[i] = (uint32_t)((i * 13) & 0xfff);
            b[i] = (uint32_t)((i * 29) & 0xfff);
            a[i] = 0xfff;
        }
        caca_set_dither_palette(d, r, g, b, a);
    }
    caca_dither_bitmap(cv, 0, 0, W, H / 2, d, p);
    dump(cv, name);
    caca_free_dither(d);
    caca_free_canvas(cv);
}

static void run_pix16(const char *name)
{
    uint16_t p[W * H];
    int x, y;
    for(y = 0; y < H; y++)
        for(x = 0; x < W; x++)
        {
            unsigned r = (unsigned)(x * 31 / (W - 1));
            unsigned g = (unsigned)(y * 63 / (H - 1));
            unsigned b = (unsigned)((x * 3 + y * 5) & 31);
            p[y * W + x] = (uint16_t)((r << 11) | (g << 5) | b);
        }
    caca_canvas_t *cv = caca_create_canvas(W, H / 2);
    caca_dither_t *d = caca_create_dither(16, W, H, W * 2,
                                          0xf800, 0x07e0, 0x001f, 0);
    caca_dither_bitmap(cv, 0, 0, W, H / 2, d, p);
    dump(cv, name);
    caca_free_dither(d);
    caca_free_canvas(cv);
}

static void run_pix32(const char *name)
{
    uint32_t p[W * H];
    int x, y;
    for(y = 0; y < H; y++)
        for(x = 0; x < W; x++)
        {
            unsigned r = (unsigned)(x * 16);
            unsigned g = (unsigned)(y * 16);
            unsigned b = (unsigned)((x * 37 + y * 53) & 0xff);
            p[y * W + x] = (0xffu << 24) | (r << 16) | (g << 8) | b;
        }
    caca_canvas_t *cv = caca_create_canvas(W, H / 2);
    caca_dither_t *d = caca_create_dither(32, W, H, W * 4,
                                          0x00ff0000, 0x0000ff00, 0x000000ff,
                                          0xff000000);
    caca_dither_bitmap(cv, 0, 0, W, H / 2, d, p);
    dump(cv, name);
    caca_free_dither(d);
    caca_free_canvas(cv);
}

int main(void)
{
    make_pixels();
    run("default",    "fstein", "ascii",  "full16",   "prefilter");
    run("none",       "none",   "ascii",  "full16",   "prefilter");
    run("ordered2",   "ordered2","ascii", "full16",   "prefilter");
    run("ordered4",   "ordered4","ascii", "full16",   "prefilter");
    run("ordered8",   "ordered8","ascii", "full16",   "prefilter");
    run("shades",     "fstein", "shades", "full16",   "prefilter");
    run("blocks",     "fstein", "blocks", "full16",   "prefilter");
    run("fullgray",   "fstein", "ascii",  "fullgray", "prefilter");
    run("mono",       "fstein", "ascii",  "mono",      "prefilter");
    run("gray",       "fstein", "ascii",  "gray",      "prefilter");
    run("16color",    "fstein", "ascii",  "16",        "prefilter");
    run("full8",      "fstein", "ascii",  "full8",     "prefilter");
    run("noaa",       "fstein", "ascii",  "full16",    "none");
    run_gamma("gamma22", 2.2f);
    run_gamma("gamma05", 0.5f);
    run_pix8("pix8", 0);
    run_pix8("pix8pal", 1);
    run_pix16("pix16");
    run_pix32("pix32");
    return 0;
}
