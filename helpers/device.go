package helpers

import (
	"os"

	"github.com/TankerHQ/sdk-go/v2/core"
)

type Device struct {
	User
	Path string
}

func (device Device) Destroy() error {
	return os.RemoveAll(device.Path)
}

func (device Device) CreateSession() (*core.Tanker, error) {
	return core.CreateTanker(device.AppID, device.Url, device.Path)
}

func (device Device) Start() (*core.Tanker, error) {
	tanker, err := device.CreateSession()
	if err != nil {
		return nil, err
	}
	_, err = StartTankerSession(tanker, device.Identity)
	return tanker, err
}

func StartTankerSession(tanker *core.Tanker, identity string) (status core.Status, err error) {
	status, err = tanker.Start(identity)
	if err != nil {
		return
	}
	switch status {
	case core.TankerStatusIdentityVerificationNeeded:
		err = tanker.VerifyIdentity(core.PassphraseVerification{Passphrase: "multipass"})
	case core.TankerStatusIdentityRegistrationNeeded:
		err = tanker.RegisterIdentity(core.PassphraseVerification{Passphrase: "multipass"})
	}
	return
}
