package core

/*
#include <ctanker.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// CreateGroup creates a Tanker group. The group will be created with the user's PublicIdentities provided.
// This function succeeds or fails completely, e.g. if a PublicIdentity is invalid, no group is created.
// On success, the created group ID is returned.
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

// UpdateGroupMembers updates the members of a group. The new group members will automatically
// get access to all resources previously shared with the group.
func (t *Tanker) UpdateGroupMembers(groupID string, publicIdentitiesToAdd []string) error {
	nbIDs := len(publicIdentitiesToAdd)
	cgroupID := C.CString(groupID)
	ids := toCArray(publicIdentitiesToAdd)
	defer freeCArray(ids, nbIDs)
	defer C.free(unsafe.Pointer(cgroupID))
	_, err := await(C.tanker_update_group_members(t.instance, cgroupID, ids, C.uint64_t(nbIDs)))
	return err
}
