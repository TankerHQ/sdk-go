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

type Status uint32

const (
	TankerStatusStopped Status = iota
	TankerStatusReady
	TankerStatusIdentityRegistrationNeeded
	TankerStatusIdentityVerificationNeeded
)

type EncryptOptions struct {
	Recipients []string
	Groups     []string
}

//unsafeANSIToString transforms a *C.char to a GoString. The *C.char is free'd.
func unsafeANSIToString(pointer unsafe.Pointer) string {
	charP := (*C.char)(unsafe.Pointer(pointer))
	str := C.GoString(charP)
	C.tanker_free_buffer(unsafe.Pointer(charP))
	return str
}

// Tanker object
type Tanker struct {
	instance *C.tanker_t
}

func initializeTanker() {
	C.tanker_init()
}

type EventHandler func()
type EventType uint32

const (
	EventSessionClosed EventType = C.TANKER_EVENT_SESSION_CLOSED
	EventDeviceRevoked EventType = C.TANKER_EVENT_DEVICE_REVOKED
)

func (t *Tanker) ConnectEvent(event EventType, handler EventHandler) error {
	panic("Not Implemented")
	return nil
}

type DeviceDescription struct {
	DeviceID  string
	IsRevoked bool
}

func Version() string {
	currentVersion := "dev"
	return currentVersion
}

func NativeVersion() string {
	return C.GoString(C.tanker_version_string())
}

// CreateTanker Instance
func CreateTanker(appID string, Url string, writablePath string) (*Tanker, error) {
	initializeTanker()

	cappID := C.CString(appID)
	url := C.CString(Url)
	cwritablePath := C.CString(writablePath)
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
	result, err := Await(C.tanker_create(coptions))
	if err != nil {
		return nil, err
	}
	this.instance = (*C.tanker_t)(result)

	return &this, nil
}

//Destroy destroys this tanker object
func (t *Tanker) Destroy() error {
	_, err := Await(C.tanker_destroy(t.instance))
	return err
}

func (t *Tanker) Start(identity string) (Status, error) {
	cidentity := C.CString(identity)
	result, err := Await(C.tanker_start(t.instance, cidentity))
	defer C.free(unsafe.Pointer(cidentity))
	if err != nil {
		return TankerStatusStopped, err
	}
	status := (Status)((uintptr)(result))
	return status, nil
}

func (t *Tanker) Stop() error {
	_, err := Await(C.tanker_stop(t.instance))
	return err
}

func (t *Tanker) GetStatus() Status {
	return Status(C.tanker_status(t.instance))
}

func (t *Tanker) GetDeviceID() (*string, error) {
	result, err := Await(C.tanker_device_id(t.instance))
	if err != nil {
		return nil, err
	}
	ID := unsafeANSIToString(result)
	return &ID, nil
}

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
	_, err := Await(
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

func (t *Tanker) Decrypt(encryptedData []byte) ([]byte, error) {
	if encryptedData == nil || len(encryptedData) == 0 {
		return nil, newError(ErrorInvalidArgument, "encryptedData must not be nil")
	}
	cencrypted := (*C.uint8_t)(unsafe.Pointer(&encryptedData[0]))
	cdecryptedSize, err := Await(C.tanker_decrypted_size(cencrypted, C.uint64_t(len(encryptedData))))
	if err != nil {
		return nil, err
	}
	decryptedSize := uint64((uintptr)(cdecryptedSize))

	clearData := make([]byte, decryptedSize)
	_, err = Await(
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

func (t *Tanker) GetResourceId(encryptedData []byte) (*string, error) {
	if encryptedData == nil || len(encryptedData) == 0 {
		return nil, newError(ErrorInvalidArgument, "encryptedData must not be nil")
	}
	result, err := Await(C.tanker_get_resource_id((*C.uchar)(unsafe.Pointer(&encryptedData[0])), C.uint64_t(len(encryptedData))))
	if err != nil {
		return nil, err
	}
	resourceID := unsafeANSIToString(result)
	return &resourceID, nil
}

func (t *Tanker) Share(resourceIDs []string, recipients []string, groups []string) error {
	if resourceIDs == nil || len(resourceIDs) == 0 {
		return fmt.Errorf("ResourceIDs must not be nil nor empty")
	}
	cresourceIds := toCArray(resourceIDs)
	crecipients := toCArray(recipients)
	cgroups := toCArray(groups)

	_, err := Await(
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

func (t *Tanker) GetDeviceList() (goDevices []DeviceDescription, err error) {
	cresult, err := Await(C.tanker_get_device_list(t.instance))
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

func (t *Tanker) RevokeDevice(deviceID string) (err error) {
	cdeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cdeviceID))
	_, err = Await(C.tanker_revoke_device(t.instance, cdeviceID))
	return
}
