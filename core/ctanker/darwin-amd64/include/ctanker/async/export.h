
#ifndef TANKER_ASYNC_EXPORT_H
#define TANKER_ASYNC_EXPORT_H

#ifdef TANKER_ASYNC_STATIC_DEFINE
#  define TANKER_ASYNC_EXPORT
#  define TANKER_ASYNC_NO_EXPORT
#else
#  ifndef TANKER_ASYNC_EXPORT
#    ifdef tanker_async_EXPORTS
        /* We are building this library */
#      define TANKER_ASYNC_EXPORT 
#    else
        /* We are using this library */
#      define TANKER_ASYNC_EXPORT 
#    endif
#  endif

#  ifndef TANKER_ASYNC_NO_EXPORT
#    define TANKER_ASYNC_NO_EXPORT 
#  endif
#endif

#ifndef TANKER_ASYNC_DEPRECATED
#  define TANKER_ASYNC_DEPRECATED __attribute__ ((__deprecated__))
#endif

#ifndef TANKER_ASYNC_DEPRECATED_EXPORT
#  define TANKER_ASYNC_DEPRECATED_EXPORT TANKER_ASYNC_EXPORT TANKER_ASYNC_DEPRECATED
#endif

#ifndef TANKER_ASYNC_DEPRECATED_NO_EXPORT
#  define TANKER_ASYNC_DEPRECATED_NO_EXPORT TANKER_ASYNC_NO_EXPORT TANKER_ASYNC_DEPRECATED
#endif

#if 0 /* DEFINE_NO_DEPRECATED */
#  ifndef TANKER_ASYNC_NO_DEPRECATED
#    define TANKER_ASYNC_NO_DEPRECATED
#  endif
#endif

#endif /* TANKER_ASYNC_EXPORT_H */
