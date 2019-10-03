
#ifndef CTANKER_EXPORT_H
#define CTANKER_EXPORT_H

#ifdef CTANKER_STATIC_DEFINE
#  define CTANKER_EXPORT
#  define CTANKER_NO_EXPORT
#else
#  ifndef CTANKER_EXPORT
#    ifdef ctanker_EXPORTS
        /* We are building this library */
#      define CTANKER_EXPORT 
#    else
        /* We are using this library */
#      define CTANKER_EXPORT 
#    endif
#  endif

#  ifndef CTANKER_NO_EXPORT
#    define CTANKER_NO_EXPORT 
#  endif
#endif

#ifndef CTANKER_DEPRECATED
#  define CTANKER_DEPRECATED 
#endif

#ifndef CTANKER_DEPRECATED_EXPORT
#  define CTANKER_DEPRECATED_EXPORT CTANKER_EXPORT CTANKER_DEPRECATED
#endif

#ifndef CTANKER_DEPRECATED_NO_EXPORT
#  define CTANKER_DEPRECATED_NO_EXPORT CTANKER_NO_EXPORT CTANKER_DEPRECATED
#endif

#if 0 /* DEFINE_NO_DEPRECATED */
#  ifndef CTANKER_NO_DEPRECATED
#    define CTANKER_NO_DEPRECATED
#  endif
#endif

#endif /* CTANKER_EXPORT_H */
