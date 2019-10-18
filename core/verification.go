package core

/*
#include <ctanker.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

type VerificationMethodType uint32

const (
	VerificationMethodTypeEmail           VerificationMethodType = C.TANKER_VERIFICATION_METHOD_EMAIL
	VerificationMethodTypePassphrase      VerificationMethodType = C.TANKER_VERIFICATION_METHOD_PASSPHRASE
	VerificationMethodTypeVerificationKey VerificationMethodType = C.TANKER_VERIFICATION_METHOD_VERIFICATION_KEY
	VerificationMethodTypeOidcIdToken     VerificationMethodType = C.TANKER_VERIFICATION_METHOD_OIDC_ID_TOKEN
)

type VerificationMethod struct {
	Type  VerificationMethodType
	Email *string
}

type AttachResult struct {
	Status Status
	Method *VerificationMethod
}

type EmailVerification struct {
	Email            string
	VerificationCode string
}

type PassphraseVerification struct {
	Passphrase string
}

type KeyVerification struct {
	Key string
}

type OidcVerification struct {
	OidcIdToken string
}

func convertVerificationToTanker(verif interface{}) *C.tanker_verification_t {
	result := &C.tanker_verification_t{
		version: 3,
	}

	switch t := verif.(type) {
	case EmailVerification:
		result.verification_method_type = C.TANKER_VERIFICATION_METHOD_EMAIL
		result.email_verification = C.tanker_email_verification_t{
			version:           1,
			email:             C.CString(t.Email),
			verification_code: C.CString(t.VerificationCode),
		}
	case PassphraseVerification:
		result.verification_method_type = C.TANKER_VERIFICATION_METHOD_PASSPHRASE
		result.passphrase = C.CString(t.Passphrase)
	case KeyVerification:
		result.verification_method_type = C.TANKER_VERIFICATION_METHOD_VERIFICATION_KEY
		result.verification_key = C.CString(t.Key)
	case OidcVerification:
		result.verification_method_type = C.TANKER_VERIFICATION_METHOD_OIDC_ID_TOKEN
		result.oidc_id_token = C.CString(t.OidcIdToken)
	}
	return result
}

func freeVerif(verif *C.tanker_verification_t) {
	switch verif.verification_method_type {
	case C.TANKER_VERIFICATION_METHOD_EMAIL:
		C.free(unsafe.Pointer(verif.email_verification.email))
		C.free(unsafe.Pointer(verif.email_verification.verification_code))
	case C.TANKER_VERIFICATION_METHOD_PASSPHRASE:
		C.free(unsafe.Pointer(verif.passphrase))
	case C.TANKER_VERIFICATION_METHOD_VERIFICATION_KEY:
		C.free(unsafe.Pointer(verif.verification_key))
	case C.TANKER_VERIFICATION_METHOD_OIDC_ID_TOKEN:
		C.free(unsafe.Pointer(verif.oidc_id_token))
	}
}

func (t *Tanker) RegisterIdentity(verification interface{}) error {
	cverif := convertVerificationToTanker(verification)
	defer freeVerif(cverif)

	_, err := Await(C.tanker_register_identity(t.instance, cverif))
	return err
}

func (t *Tanker) VerifyIdentity(verification interface{}) error {
	cverif := convertVerificationToTanker(verification)
	defer freeVerif(cverif)
	_, err := Await(C.tanker_verify_identity(t.instance, cverif))
	return err
}

func (t *Tanker) SetVerificationMethod(verification interface{}) error {
	cverif := convertVerificationToTanker(verification)
	defer freeVerif(cverif)
	_, err := Await(C.tanker_set_verification_method(t.instance, cverif))
	return err
}

func convertVerificationMethodToTanker(cmethod *C.tanker_verification_method_t) *VerificationMethod {
	if cmethod == nil {
		return nil
	}
	var email *string
	if cmethod.verification_method_type == C.TANKER_VERIFICATION_METHOD_EMAIL {
		dummy := C.GoString(cmethod.email)
		email = &dummy
	}
	return &VerificationMethod{
		Type:  VerificationMethodType(cmethod.verification_method_type),
		Email: email,
	}
}

func (t *Tanker) GetVerificationMethods() ([]VerificationMethod, error) {
	result, err := Await(C.tanker_get_verification_methods(t.instance))
	if err != nil {
		return nil, err
	}
	methodList := (*C.tanker_verification_method_list_t)(unsafe.Pointer(result))
	count := (int)(methodList.count)
	goMethods := make([]VerificationMethod, 0, count)
	for i := 0; i < count; i++ {
		cmethod := (*C.tanker_verification_method_t)(unsafe.Pointer(uintptr(unsafe.Pointer(methodList.methods)) + (unsafe.Sizeof(*methodList.methods) * uintptr(i))))
		goMethods = append(goMethods, *convertVerificationMethodToTanker(cmethod))
	}
	C.tanker_free_verification_method_list(methodList)
	return goMethods, nil
}

func (t *Tanker) AttachProvisionalIdentity(provisionalIdentity string) (*AttachResult, error) {
	cidentity := C.CString(provisionalIdentity)
	defer C.free(unsafe.Pointer(cidentity))
	result, err := Await(C.tanker_attach_provisional_identity(t.instance, cidentity))
	if err != nil {
		return nil, err
	}
	cresult := (*C.tanker_attach_result_t)(result)
	defer C.tanker_free_attach_result(cresult)
	attachResult := &AttachResult{
		Status: Status(cresult.status),
		Method: convertVerificationMethodToTanker(cresult.method),
	}

	return attachResult, err
}

func (t *Tanker) VerifyProvisionalIdentity(verification interface{}) error {
	_, err := Await(C.tanker_verify_provisional_identity(t.instance, convertVerificationToTanker(verification)))
	return err
}

func (t *Tanker) GenerateVerificationKey() (*string, error) {
	result, err := Await(C.tanker_generate_verification_key(t.instance))
	if err != nil {
		return nil, err
	}
	key := unsafeANSIToString(result)
	return &key, nil
}
