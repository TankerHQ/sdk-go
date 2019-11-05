// Bindings for the Tanker Core SDK
package core

/*
#include <stdlib.h>
#include <ctanker.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
) //nolint

// Status represents the Tanker current status.
type Status uint32

const (
	StatusStopped Status = iota
	StatusReady
	StatusIdentityRegistrationNeeded
	StatusIdentityVerificationNeeded
)

// EncryptOptions contains user and group recipients to share with during an @Encrypt()
type EncryptOptions struct {
	// Recipients is a list of the public identities of each recipient to share with
	Recipients []string
	// Groups is a list of group ids to share with
	Groups []string
}

// unsafeANSIToString transforms a *C.char to a GoString. The *C.char is free'd.
func unsafeANSIToString(pointer unsafe.Pointer) string {
	charP := (*C.char)(unsafe.Pointer(pointer))
	str := C.GoString(charP)
	C.tanker_free_buffer(unsafe.Pointer(charP))
	return str
}

// Tanker represents a Tanker instance.
type Tanker struct {
	instance *C.tanker_t
}

// initializeTanker initializes the native library.
// Must be called a least once before any Tanker operations.
// This functions is called each time a Tanker instance is created.
func initializeTanker() {
	C.tanker_init()
}

// EventHandler defines the function object type used by RegisterEventHandler().
type EventHandler func()

// EventType represents the type of event one can register and be notified of.
type EventType uint32

const (
	EventSessionClosed EventType = C.TANKER_EVENT_SESSION_CLOSED
	EventDeviceRevoked EventType = C.TANKER_EVENT_DEVICE_REVOKED
)

// RegisterEventHandler registers an event handler for the given EventType.
func (t *Tanker) RegisterEventHandler(event EventType, handler EventHandler) error {
	panic("Not Implemented")
	return nil
}

// DeviceDescription contains the id of a device and whether this device has been revoked.
type DeviceDescription struct {
	DeviceID  string
	IsRevoked bool
}

// Version returns the current version of this SDK.
func Version() string {
	currentVersion := "2.2.1-beta1"
	return currentVersion
}

// nativeVersion returns the native version currently used by this SDK.
func nativeVersion() string {
	return C.GoString(C.tanker_version_string())
}

// TankerOptions defines the options needed to create a new Tanker
// instance with NewTanker().
type TankerOptions struct {
	// The Application ID you want to use.
	AppID string
	// An existing filesystem path to store persistent user data.
	WritablePath string
	// The url of the Tanker service. Should be left to nil.
	Url *string
}

// NewTanker creates a new a Tanker instance.
//  session, err := core.NewTanker(core.TankerOptions{"<your app ID>", "/home/user/.config/fancyname/", nil})
func NewTanker(options TankerOptions) (*Tanker, error) {
	initializeTanker()

	cappID := C.CString(options.AppID)
	url := (*C.char)(unsafe.Pointer(uintptr(0)))
	if options.Url != nil {
		url = C.CString(*options.Url)
	}
	cwritablePath := C.CString(options.WritablePath)
	sdkgo := C.CString("sdk-go")
	version := C.CString(Version())
	defer func() {
		C.free(unsafe.Pointer(cappID))
		C.free(unsafe.Pointer(url))
		C.free(unsafe.Pointer(cwritablePath))
		C.free(unsafe.Pointer(sdkgo))
		C.free(unsafe.Pointer(version))
	}()
	this := Tanker{}
	coptions := &C.tanker_options_t{
		version:       2,
		app_id:        cappID,
		url:           url,
		writable_path: cwritablePath,
		sdk_type:      sdkgo,
		sdk_version:   version,
	}
	result, err := await(C.tanker_create(coptions))
	if err != nil {
		return nil, err
	}
	this.instance = (*C.tanker_t)(result)

	return &this, nil
}

// Destroy destroys this Tanker instance. This functions performs
// internal resources cleanups and calls Stop() if necessary.
// No further operations is possible on this instance after calling Destroy(),
// you'll need to create a new one.
func (t *Tanker) Destroy() error {
	_, err := await(C.tanker_destroy(t.instance))
	return err
}

// Start starts a new Tanker session and returns a status.
//
//  User := app.AuthenticatedUser(id, password)
//  status, err := tanker.Start(user.TankerIdentity)
//  if err == nil && status == core.StatusReady {
//	 	// Let's encrypt, share and decrypts data!
//  }
// The Tanker status must be StatusStopped before calling Start().
func (t *Tanker) Start(identity string) (Status, error) {
	cidentity := C.CString(identity)
	result, err := await(C.tanker_start(t.instance, cidentity))
	defer C.free(unsafe.Pointer(cidentity))
	if err != nil {
		return StatusStopped, err
	}
	status := (Status)((uintptr)(result))
	return status, nil
}

// Stop stops the current Tanker Session. This session can either
// be destroyed with Destroy() or be restarted with Start().
func (t *Tanker) Stop() error {
	_, err := await(C.tanker_stop(t.instance))
	return err
}

// GetStatus retrieves the current Tanker session status.
func (t *Tanker) GetStatus() Status {
	return Status(C.tanker_status(t.instance))
}

// GetDeviceID retrieves the current Tanker device's ID. Each device
// has its own ID and can be identified as such.
func (t *Tanker) GetDeviceID() (*string, error) {
	result, err := await(C.tanker_device_id(t.instance))
	if err != nil {
		return nil, err
	}
	ID := unsafeANSIToString(result)
	return &ID, nil
}

// Encrypt encrypts the passed []byte and returns the result. To share the resulting
// encrypted resource with either or both individuals and groups, fill the EncryptOptions parameter.
func (t *Tanker) Encrypt(clearData []byte, options *EncryptOptions) ([]byte, error) {
	if clearData == nil {
		return nil, newError(ErrorInvalidArgument, "clearData must not be nil")
	}
	var cClearData unsafe.Pointer
	if len(clearData) == 0 {
		cClearData = C.CBytes(clearData)
	} else {
		cClearData = unsafe.Pointer(&clearData[0])
	}
	encryptedSize := C.tanker_encrypted_size(C.uint64_t(len(clearData)))

	encryptedData := make([]byte, encryptedSize)
	var coptions *C.tanker_encrypt_options_t = nil
	if options != nil {
		coptions = convertOptions(*options)
		defer freeCArray(coptions.recipient_public_identities, len(options.Recipients))
		defer freeCArray(coptions.recipient_gids, len(options.Groups))
	}
	_, err := await(
		C.tanker_encrypt(
			t.instance,
			(*C.uint8_t)(unsafe.Pointer(&encryptedData[0])),
			(*C.uint8_t)(cClearData),
			C.uint64_t(len(clearData)),
			coptions,
		),
	)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}

// Decrypt decrypts the pass encrypted resource and return the original clear data.
func (t *Tanker) Decrypt(encryptedData []byte) ([]byte, error) {
	if encryptedData == nil || len(encryptedData) == 0 {
		return nil, newError(ErrorInvalidArgument, "encryptedData must not be nil")
	}
	cencrypted := (*C.uint8_t)(unsafe.Pointer(&encryptedData[0]))
	cdecryptedSize, err := await(C.tanker_decrypted_size(cencrypted, C.uint64_t(len(encryptedData))))
	if err != nil {
		return nil, err
	}
	decryptedSize := uint64((uintptr)(cdecryptedSize))

	clearData := make([]byte, decryptedSize)
	_, err = await(
		C.tanker_decrypt(
			t.instance,
			(*C.uint8_t)(unsafe.Pointer(&clearData[0])),
			(*C.uint8_t)(unsafe.Pointer(&encryptedData[0])),
			C.uint64_t(len(encryptedData)),
		),
	)
	if err != nil {
		return nil, err
	}
	return clearData, nil
}

// GetResourceId retrieves an encrypted resource's ID.
// The resource ID can be pass to a call to Share().
func (t *Tanker) GetResourceId(encryptedData []byte) (*string, error) {
	if encryptedData == nil || len(encryptedData) == 0 {
		return nil, newError(ErrorInvalidArgument, "encryptedData must not be nil")
	}
	result, err := await(C.tanker_get_resource_id((*C.uchar)(unsafe.Pointer(&encryptedData[0])), C.uint64_t(len(encryptedData))))
	if err != nil {
		return nil, err
	}
	resourceID := unsafeANSIToString(result)
	return &resourceID, nil
}

// Share shares a list of resource to a list of recipients and/or groups
// This function either fully succeeds or fails. In case of failure,
// nothing is share with any recipient or group.
func (t *Tanker) Share(resourceIDs []string, recipients []string, groups []string) error {
	if resourceIDs == nil || len(resourceIDs) == 0 {
		return fmt.Errorf("ResourceIDs must not be nil nor empty")
	}
	cresourceIds := toCArray(resourceIDs)
	crecipients := toCArray(recipients)
	cgroups := toCArray(groups)

	_, err := await(
		C.tanker_share(
			t.instance,
			crecipients,
			C.uint64_t(len(recipients)),
			cgroups,
			C.uint64_t(len(groups)),
			cresourceIds,
			C.uint64_t(len(resourceIDs)),
		),
	)
	return err
}

// GetDeviceList retrieves the user's device list.
// The current Tanker status must be StatusReady.
func (t *Tanker) GetDeviceList() (goDevices []DeviceDescription, err error) {
	cresult, err := await(C.tanker_get_device_list(t.instance))
	if err != nil {
		return
	}
	cdeviceList := (*C.tanker_device_list_t)(cresult)
	count := (int)(cdeviceList.count)
	goDevices = make([]DeviceDescription, 0, count)
	for i := 0; i < count; i++ {
		cdevice := (*C.tanker_device_list_elem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cdeviceList.devices)) + (unsafe.Sizeof(*cdeviceList.devices) * uintptr(i))))
		goDevices = append(goDevices, DeviceDescription{DeviceID: C.GoString(cdevice.device_id), IsRevoked: bool(cdevice.is_revoked)})
	}
	C.tanker_free_device_list(cdeviceList)
	return
}

// RevokeDevice revokes one of the user's devices.
func (t *Tanker) RevokeDevice(deviceID string) (err error) {
	cdeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cdeviceID))
	_, err = await(C.tanker_revoke_device(t.instance, cdeviceID))
	return
}
