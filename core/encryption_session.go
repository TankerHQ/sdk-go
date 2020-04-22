package core

/*
#include <stdlib.h>
#include <ctanker.h>
*/
import "C"
import (
	"unsafe"
)

// Represents an EncryptionSession instance.
type EncryptionSession struct {
	instance *C.tanker_encryption_session_t
}

// Destroy destroys the session, internal resource cleanup is performed
// this object is no longer usable.
func (s *EncryptionSession) Destroy() {
	_, _ = await(C.tanker_encryption_session_close(s.instance))
}

// GetResourceId retrieves the session resource's ID.
// The resource ID can be passed to a call to Share().
func (s *EncryptionSession) GetResourceId() string {
	result, _ := await(C.tanker_encryption_session_get_resource_id(s.instance))
	return unsafeANSIToString(result)
}

// Encrypts the passed []byte with the session and returns the result.
func (s *EncryptionSession) Encrypt(clearData []byte) ([]byte, error) {
	if clearData == nil {
		return nil, newError(ErrorInvalidArgument, "clearData must not be nil")
	}
	var cClearData unsafe.Pointer
	if len(clearData) == 0 {
		cClearData = C.CBytes(clearData)
	} else {
		cClearData = unsafe.Pointer(&clearData[0])
	}
	encryptedSize := C.tanker_encryption_session_encrypted_size(C.uint64_t(len(clearData)))

	encryptedData := make([]byte, encryptedSize)

	_, err := await(
		C.tanker_encryption_session_encrypt(
			s.instance,
			(*C.uint8_t)(unsafe.Pointer(&encryptedData[0])),
			(*C.uint8_t)(cClearData),
			C.uint64_t(len(clearData)),
		),
	)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}
