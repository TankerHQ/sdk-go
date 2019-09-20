package core

/*
#include <stdlib.h>
#include <ctanker/admin.h>
*/
import "C" //nolint

import (
	"unsafe"
)

//Admin .
type Admin struct {
	admin *C.tanker_admin_t
}

//AppDescriptor .
type AppDescriptor struct {
	Name       string
	ID         string
	PrivateKey string
	PublicKey  string
}

// CreateAdmin creates an admin object
func CreateAdmin(URL string, IDToken string) (*Admin, error) {
	url := C.CString(URL)
	token := C.CString(IDToken)
	defer C.free(unsafe.Pointer(url))
	defer C.free(unsafe.Pointer(token))
	result, err := Await(C.tanker_admin_connect(url, token))
	if err != nil {
		return nil, err
	}
	admin := (*C.tanker_admin_t)(result)
	that := &Admin{
		admin,
	}
	return that, nil
}

//CreateApp creates an aplication on the tanker server
func (adm Admin) CreateApp(Name string) (*AppDescriptor, error) {
	name := C.CString(Name)
	defer C.free(unsafe.Pointer(name))
	result, err := Await(C.tanker_admin_create_app(adm.admin, name))
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

//DeleteApp destroys the application on the Tanker servers.
func (adm Admin) DeleteApp(AppID string) error {
	appID := C.CString(AppID)
	defer C.free(unsafe.Pointer(appID))
	_, err := Await(C.tanker_admin_delete_app(adm.admin, appID))
	if err != nil {
		return err
	}
	return nil
}

//GetVerificationCode retrieve the verificaton on a test trustchain
func (adm *Admin) GetVerificationCode(AppID string, Email string) (*string, error) {
	appID := C.CString(AppID)
	email := C.CString(Email)
	defer C.free(unsafe.Pointer(appID))
	defer C.free(unsafe.Pointer(email))
	result, err := Await(C.tanker_admin_get_verification_code(adm.admin, appID, email))
	defer C.free(result)
	if err != nil {
		return nil, err
	}
	code := C.GoString((*C.char)(result))
	return &code, nil
}

// Destroy an Admin
func (adm Admin) Destroy() {
	_, _ = Await(C.tanker_admin_destroy(adm.admin))
}
