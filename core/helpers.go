package core

/*
#include <stdlib.h>
#include <ctanker.h>
*/
import "C"
import "unsafe"

func toCArray(goStrings []string) **C.char {
	l := len(goStrings)
	if len(goStrings) == 0 {
		return nil
	}
	res := (**C.char)(unsafe.Pointer(C.malloc(C.size_t(uintptr(l) * unsafe.Sizeof(uintptr(0))))))
	for i := 0; i < l; i++ {
		*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(res)) + unsafe.Sizeof(uintptr(0))*uintptr(i))) = C.CString(goStrings[i])
	}
	return res
}

func freeCArray(array **C.char, size int) {
	for i := 0; i < size; i++ {
		C.free(unsafe.Pointer(*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(array)) + unsafe.Sizeof(uintptr(0))*uintptr(i)))))
	}
	C.free(unsafe.Pointer(array))
}

func convertOptions(options EncryptOptions) *C.tanker_encrypt_options_t {
	return &C.tanker_encrypt_options_t{
		version:                        2,
		recipient_public_identities:    toCArray(options.Recipients),
		nb_recipient_public_identities: C.uint32_t(len(options.Recipients)),
		recipient_gids:                 toCArray(options.Groups),
		nb_recipient_gids:              C.uint32_t(len(options.Groups)),
	}
}
