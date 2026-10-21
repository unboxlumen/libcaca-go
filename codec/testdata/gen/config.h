#ifndef CONFIG_H
#define CONFIG_H
#include <errno.h>
#include <string.h>
#define PACKAGE_VERSION "0.99.beta20"
#define HAVE_STRCASECMP 1
#define HAVE_GETENV 1
#define HAVE_UNISTD_H 1
#define HAVE_VSNPRINTF 1
#define HAVE_SNPRINTF 1
#ifndef strcasecmp
#define strcasecmp _stricmp
#endif
#ifndef strncasecmp
#define strncasecmp _strnicmp
#endif
#endif
