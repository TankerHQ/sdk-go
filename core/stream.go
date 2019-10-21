package core

import (
	"io"
	"reflect"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

/*
#include <stdlib.h>
#include <ctanker.h>

void gotanker_proxy_input_source_read(uint8_t *buffer, int64_t buffer_size, tanker_stream_read_operation_t *operation, void *additional_data);

static tanker_future_t *gotanker_stream_encrypt(tanker_t *ctanker, void *additional_data, tanker_encrypt_options_t const *options) {
	return tanker_stream_encrypt(ctanker, gotanker_proxy_input_source_read, additional_data, options);
}

static tanker_future_t *gotanker_stream_decrypt(tanker_t *ctanker, void *additional_data) {
	return tanker_stream_decrypt(ctanker, gotanker_proxy_input_source_read, additional_data);
}
*/
import "C"

type streamWrapper struct {
	reader io.Reader
	err    error
}
type OutputStream struct {
	stream   *C.tanker_stream_t
	wrapper  *streamWrapper
	todelete unsafe.Pointer
}

//export gotanker_proxy_input_source_read
func gotanker_proxy_input_source_read(
	buffer *C.uint8_t,
	buffer_size C.int64_t,
	operation *C.tanker_stream_read_operation_t,
	additional_data unsafe.Pointer,
) {
	go func() {
		wrapper := gopointer.Restore(additional_data).(*streamWrapper)
		slice := &reflect.SliceHeader{Data: uintptr(unsafe.Pointer(buffer)), Len: int(buffer_size), Cap: int(buffer_size)}
		rbuf := *(*[]byte)(unsafe.Pointer(slice))
		nb_read, err := wrapper.reader.Read(rbuf)
		if err != nil {
			wrapper.err = err
			C.tanker_stream_read_operation_finish(operation, -1)
		}
		C.tanker_stream_read_operation_finish(operation, C.int64_t(nb_read))
	}()
}

func (s *OutputStream) Read(buffer []byte) (int, error) {
	askedLen := C.int64_t(len(buffer))
	result, err := await(C.tanker_stream_read(s.stream, (*C.uchar)(unsafe.Pointer(&buffer[0])), askedLen))
	nb_read := int((uintptr)(result))
	if err != nil {
		if s.wrapper.err != nil {
			return nb_read, s.wrapper.err
		}
		return nb_read, err
	}
	if nb_read == 0 {
		return 0, io.EOF
	}
	return nb_read, nil
}

func (s *OutputStream) Destroy() {
	_, _ = await(C.tanker_stream_close(s.stream))
	gopointer.Unref(unsafe.Pointer(s.todelete))
}

func (s *OutputStream) GetResourceID() (*string, error) {
	result, err := await(C.tanker_stream_get_resource_id(s.stream))
	if err != nil {
		return nil, err
	}
	streamID := unsafeANSIToString(result)
	return &streamID, nil
}

func (t *Tanker) StreamEncrypt(reader io.Reader, options *EncryptOptions) (*OutputStream, error) {
	var coptions *C.tanker_encrypt_options_t = nil
	if options != nil {
		coptions = convertOptions(*options)
		defer freeCArray(coptions.recipient_public_identities, len(options.Recipients))
		defer freeCArray(coptions.recipient_gids, len(options.Groups))
	}
	wrapper := streamWrapper{
		reader: reader,
		err:    nil,
	}
	wrapped := gopointer.Save(&wrapper)
	result, err := await(C.gotanker_stream_encrypt(t.instance, wrapped, coptions))
	if err != nil {
		return nil, err
	}
	return &OutputStream{
		wrapper:  &wrapper,
		todelete: wrapped,
		stream:   (*C.tanker_stream_t)(unsafe.Pointer(result)),
	}, nil
}

func (t *Tanker) StreamDecrypt(reader io.Reader) (*OutputStream, error) {
	wrapper := streamWrapper{reader: reader, err: nil}
	result, err := await(C.gotanker_stream_decrypt(t.instance, gopointer.Save(&wrapper)))
	if err != nil {
		return nil, err
	}
	return &OutputStream{stream: (*C.tanker_stream_t)(result), wrapper: &wrapper}, nil
}
