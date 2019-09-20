package core

/*
#include <ctanker.h>

void* tanker_then_handler_proxy(tanker_future_t*, void *v);

static tanker_future_t *_tanker_future_then(tanker_future_t *fut, void* user_data) {
	tanker_future_t *thenFut = tanker_future_then(fut, tanker_then_handler_proxy, user_data);
	tanker_future_destroy(fut);
	tanker_future_destroy(thenFut);
}
*/
import "C"

import (
	"errors"
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
		*tan <- futureResult{err: errors.New(C.GoString(err.message))}
		return nil
	}
	*tan <- futureResult{result: C.tanker_future_get_voidptr(fut)}
	return nil
}

//Await kind of awaits for a tanker_future_t to complete, this is dark magic, beware.
func Await(future *C.tanker_future_t) (unsafe.Pointer, error) {
	tan := make(resultChan)

	C._tanker_future_then(future, gopointer.Save(&tan))
	result := <-tan
	if result.err != nil {
		return nil, result.err
	}
	return result.result, nil
}
