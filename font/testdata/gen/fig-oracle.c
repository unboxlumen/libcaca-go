#include <stdio.h>
#include <stdlib.h>
#include "caca.h"

static void run(const char *font, const char *name)
{
    caca_canvas_t *cv = caca_create_canvas(0, 0);
    if(caca_canvas_set_figfont(cv, font) < 0) { fputs("set_figfont failed\n", stderr); exit(1); }
    caca_set_figfont_width(cv, 1000);
    caca_put_figchar(cv, 'A');
    caca_put_figchar(cv, 'B');
    caca_flush_figlet(cv);
    size_t len = 0;
    void *buf = caca_export_canvas_to_memory(cv, "caca", &len);
    char path[256];
    snprintf(path, sizeof(path), "golden/%s", name);
    FILE *f = fopen(path, "wb");
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
    caca_free_canvas(cv);
}

static void run_text(const char *font, const char *name, int width, const char *text)
{
    caca_canvas_t *cv = caca_create_canvas(0, 0);
    if(caca_canvas_set_figfont(cv, font) < 0) { fputs("set_figfont failed\n", stderr); exit(1); }
    caca_set_figfont_width(cv, width);
    for(const char *p = text; *p; p++)
        caca_put_figchar(cv, (unsigned char)*p);
    caca_flush_figlet(cv);
    size_t len = 0;
    void *buf = caca_export_canvas_to_memory(cv, "caca", &len);
    char path[256];
    snprintf(path, sizeof(path), "golden/%s", name);
    FILE *f = fopen(path, "wb");
    fwrite(buf, 1, len, f);
    fclose(f);
    free(buf);
    caca_free_canvas(cv);
}

int main(void)
{
    run("figfonts/fig-full.flf", "fig-full.caca");
    run("figfonts/fig-kern.flf", "fig-kern.caca");
    run("figfonts/fig-smush.flf", "fig-smush.caca");
    run_text("figfonts/fig-full.flf", "fig-newline.caca", 1000, "AB\nAB");
    run_text("figfonts/fig-full.flf", "fig-wrap.caca", 3, "ABA");
    return 0;
}
