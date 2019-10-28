package core

/*
#include <stdlib.h>
#include <ctanker/admin.h>
*/
import "C" //nolint

import (
	"unsafe"
)

// Admin allows you to create, destroy Application.
type Admin struct {
	admin *C.tanker_admin_t
}

// AppDescriptor contains properties of a Tanker application.
type AppDescriptor struct {
	Name       string
	ID         string
	PrivateKey string
	PublicKey  string
}

// NewAdmin creates a new admin session.
func NewAdmin(URL string, IDToken string) (*Admin, error) {
	url := C.CString(URL)
	token := C.CString(IDToken)
	defer C.free(unsafe.Pointer(url))
	defer C.free(unsafe.Pointer(token))
	result, err := await(C.tanker_admin_connect(url, token))
	if err != nil {
		return nil, err
	}
	admin := (*C.tanker_admin_t)(result)
	that := &Admin{
		admin,
	}
	return that, nil
}

// NewApp creates a Tanker application on the Tanker server.
func (adm Admin) NewApp(Name string) (*AppDescriptor, error) {
	name := C.CString(Name)
	defer C.free(unsafe.Pointer(name))
	result, err := await(C.tanker_admin_create_app(adm.admin, name))
	if err != nil {
		return nil, err
	}
	app := (*C.tanker_app_descriptor_t)(result)
	that := &AppDescriptor{
		Name:       C.GoString(app.name),
		ID:         C.GoString(app.id),
		PrivateKey: C.GoString(app.private_key),
		PublicKey:  C.GoString(app.public_key),
	}
	C.tanker_admin_app_descriptor_free(app)
	return that, nil
}

// DeleteApp destroys the application on the Tanker server.
func (adm Admin) DeleteApp(AppID string) error {
	appID := C.CString(AppID)
	defer C.free(unsafe.Pointer(appID))
	_, err := await(C.tanker_admin_delete_app(adm.admin, appID))
	if err != nil {
		return err
	}
	return nil
}

// GetVerificationCode retrieves the verificaton code on a test app. The email provided must the
// same as the one in the ProvisionalIdentity you want the verification code for. The Tanker application
// must be a test application.
func (adm *Admin) GetVerificationCode(AppID string, Email string) (*string, error) {
	appID := C.CString(AppID)
	email := C.CString(Email)
	defer C.free(unsafe.Pointer(appID))
	defer C.free(unsafe.Pointer(email))
	result, err := await(C.tanker_admin_get_verification_code(adm.admin, appID, email))
	defer C.free(result)
	if err != nil {
		return nil, err
	}
	code := C.GoString((*C.char)(result))
	return &code, nil
}

// Update updates a Tanker application's settings.
func (adm Admin) Update(AppID string, OidcClientId string, OidcProvider string) error {
	appID := C.CString(AppID)
	oidcClientId := C.CString(OidcClientId)
	oidcProvider := C.CString(OidcProvider)
	defer func() {
		C.free(unsafe.Pointer(appID))
		C.free(unsafe.Pointer(oidcClientId))
		C.free(unsafe.Pointer(oidcProvider))
	}()
	_, err := await(C.tanker_admin_app_update(adm.admin, appID, oidcClientId, oidcProvider))
	return err
}

// Destroy destroys this Admin session.
func (adm Admin) Destroy() {
	_, _ = await(C.tanker_admin_destroy(adm.admin))
}
