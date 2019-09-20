package core_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/core"
	uuid "github.com/satori/go.uuid"
)

var (
	TankerIDToken = os.Getenv("TANKER_TOKEN")
	TankerUrl     = os.Getenv("TANKER_URL")
)

type App struct {
	AdminSession *core.Admin
	Descriptor   *core.AppDescriptor
}

var TestApp *App

func CreateApp() (*App, error) {
	AdminSession, err := core.CreateAdmin(TankerUrl, TankerIDToken)
	if err != nil {
		return nil, err
	}
	descriptor, err := AdminSession.CreateApp("sdk-go-test")
	if err != nil {
		return nil, err
	}
	return &App{AdminSession, descriptor}, nil
}

func (app *App) GetVerificationCode(email string) (*string, error) {
	return app.AdminSession.GetVerificationCode(app.Descriptor.ID, email)
}

func (app *App) Destroy() error {
	err := app.AdminSession.DeleteApp(app.Descriptor.ID)
	if err != nil {
		return err
	}
	app.AdminSession.Destroy()
	return nil
}

var IdConfig identity.Config

type User struct {
	AppID          string
	Url            string
	UserID         string
	Identity       string
	PublicIdentity string
}

func (app App) CreateUser() User {
	userID := uuid.NewV4().String()
	userIdentity, _ := identity.Create(IdConfig, userID)
	publicIdentity, _ := identity.GetPublicIdentity(*userIdentity)
	return User{
		AppID:          app.Descriptor.ID,
		Url:            TankerUrl,
		UserID:         userID,
		Identity:       *userIdentity,
		PublicIdentity: *publicIdentity,
	}
}

type Device struct {
	User
	Path string
}

func (user User) CreateDevice() (*Device, error) {
	dir, err := ioutil.TempDir("", user.UserID+"-")
	if err != nil {
		return nil, err
	}
	return &Device{user, dir}, nil
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
		err = tanker.VerifyIdentity(core.PassphraseVerification{"multipass"})
	case core.TankerStatusIdentityRegistrationNeeded:
		err = tanker.RegisterIdentity(core.PassphraseVerification{"multipass"})
	}
	return
}

var _ = BeforeSuite(func() {
	var err error
	TestApp, err = CreateApp()
	if err != nil {
		Fail(err.Error())
	}
	IdConfig = identity.Config{AppID: TestApp.Descriptor.ID, AppSecret: TestApp.Descriptor.PrivateKey}
})

var _ = AfterSuite(func() {
	err := TestApp.Destroy()
	if err != nil {
		Fail(err.Error())
	}
})

func TestSDK(t *testing.T) {
	p := func(record core.LogRecord) {
		fmt.Printf("[%c]{%s}'%s+%d': %s\n", record.Level, record.Category, record.File, record.Line, record.Message)
	}
	if len(TankerIDToken) == 0 || len(TankerUrl) == 0 {
		panic("TANKER_TOKEN or TANKER_URL not set")
	}
	core.SetLogHandler(p)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Test Suite")
}
