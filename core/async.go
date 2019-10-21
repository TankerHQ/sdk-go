package core

/*
#include <ctanker.h>

void* tanker_then_handler_proxy(tanker_future_t*, void *v);

static void _tanker_future_then(tanker_future_t *fut, void* user_data) {
	tanker_future_t *thenFut = tanker_future_then(fut, tanker_then_handler_proxy, user_data);
	tanker_future_destroy(fut);
	tanker_future_destroy(thenFut);
}
*/
import "C"

import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

type futureResult struct {
	result unsafe.Pointer
	err    error
}

type resultChan chan futureResult

//export tanker_then_handler_proxy
func tanker_then_handler_proxy(fut *C.tanker_future_t, v unsafe.Pointer) unsafe.Pointer {
	tan := (gopointer.Restore(v)).(*resultChan)
	gopointer.Unref(v)
	err := C.tanker_future_get_error(fut)
	if err != nil {
		terror := newError(ErrorCode(err.code), C.GoString(err.message))
		*tan <- futureResult{err: terror}
		return nil
	}
	*tan <- futureResult{result: C.tanker_future_get_voidptr(fut)}
	return nil
}

//await kind of awaits for a tanker_future_t to complete, this is dark magic, beware.
func await(future *C.tanker_future_t) (unsafe.Pointer, error) {
	tan := make(resultChan)

	C._tanker_future_then(future, gopointer.Save(&tan))
	result := <-tan
	if result.err != nil {
		return nil, result.err
	}
	return result.result, nil
}
