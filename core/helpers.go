package core

/*
#include <stdlib.h>
#include <ctanker.h>
*/
import "C"
import "unsafe"

func toCArray(goStrings []string) **C.char {
	l := len(goStrings)
	if l == 0 {
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

func convertEncryptionOptions(options EncryptionOptions) *C.tanker_encrypt_options_t {
	return &C.tanker_encrypt_options_t{
		version:           3,
		share_with_users:  toCArray(options.ShareWithUsers),
		nb_users:          C.uint32_t(len(options.ShareWithUsers)),
		share_with_groups: toCArray(options.ShareWithGroups),
		nb_groups:         C.uint32_t(len(options.ShareWithGroups)),
		share_with_self:   C.bool(options.ShareWithSelf),
	}
}

func convertSharingOptions(options SharingOptions) *C.tanker_sharing_options_t {
	return &C.tanker_sharing_options_t{
		version:           1,
		share_with_users:  toCArray(options.ShareWithUsers),
		nb_users:          C.uint32_t(len(options.ShareWithUsers)),
		share_with_groups: toCArray(options.ShareWithGroups),
		nb_groups:         C.uint32_t(len(options.ShareWithGroups)),
	}
}
