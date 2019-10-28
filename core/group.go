package core

/*
#include <ctanker.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

func (t *Tanker) CreateGroup(publicIdentities []string) (*string, error) {
	nbIDs := len(publicIdentities)
	ids := toCArray(publicIdentities)
	defer freeCArray(ids, nbIDs)
	result, err := await(C.tanker_create_group(t.instance, ids, C.uint64_t(nbIDs)))
	if err != nil {
		return nil, err
	}
	groupID := unsafeANSIToString(result)
	return &groupID, nil
}

func (t *Tanker) UpdateGroupMembers(groupID string, publicIdentitiesToAdd []string) error {
	nbIDs := len(publicIdentitiesToAdd)
	cgroupID := C.CString(groupID)
	ids := toCArray(publicIdentitiesToAdd)
	defer freeCArray(ids, nbIDs)
	defer C.free(unsafe.Pointer(cgroupID))
	_, err := await(C.tanker_update_group_members(t.instance, cgroupID, ids, C.uint64_t(nbIDs)))
	return err
}
