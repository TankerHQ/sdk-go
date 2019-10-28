package helpers

import "crypto/rand"

func RandomBytes(size int) []byte {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil
	}
	return bytes
}
