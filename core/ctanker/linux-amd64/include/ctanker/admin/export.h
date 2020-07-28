
#ifndef TANKER_ADMIN_C_EXPORT_H
#define TANKER_ADMIN_C_EXPORT_H

#ifdef TANKER_ADMIN_C_STATIC_DEFINE
#  define TANKER_ADMIN_C_EXPORT
#  define TANKER_ADMIN_C_NO_EXPORT
#else
#  ifndef TANKER_ADMIN_C_EXPORT
#    ifdef tanker_admin_c_EXPORTS
        /* We are building this library */
#      define TANKER_ADMIN_C_EXPORT 
#    else
        /* We are using this library */
#      define TANKER_ADMIN_C_EXPORT 
#    endif
#  endif

#  ifndef TANKER_ADMIN_C_NO_EXPORT
#    define TANKER_ADMIN_C_NO_EXPORT 
#  endif
#endif

#ifndef TANKER_ADMIN_C_DEPRECATED
#  define TANKER_ADMIN_C_DEPRECATED __attribute__ ((__deprecated__))
#endif

#ifndef TANKER_ADMIN_C_DEPRECATED_EXPORT
#  define TANKER_ADMIN_C_DEPRECATED_EXPORT TANKER_ADMIN_C_EXPORT TANKER_ADMIN_C_DEPRECATED
#endif

#ifndef TANKER_ADMIN_C_DEPRECATED_NO_EXPORT
#  define TANKER_ADMIN_C_DEPRECATED_NO_EXPORT TANKER_ADMIN_C_NO_EXPORT TANKER_ADMIN_C_DEPRECATED
#endif

#if 0 /* DEFINE_NO_DEPRECATED */
#  ifndef TANKER_ADMIN_C_NO_DEPRECATED
#    define TANKER_ADMIN_C_NO_DEPRECATED
#  endif
#endif

#endif /* TANKER_ADMIN_C_EXPORT_H */