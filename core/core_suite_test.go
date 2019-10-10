package core_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
	uuid "github.com/satori/go.uuid"
)

type TankerConfig struct {
	URL     string `json:"url"`
	IDToken string `json:"idToken"`
}

var (
	tankerConfigFilePath = os.Getenv("TANKER_CONFIG_FILEPATH")
	tankerConfigName     = os.Getenv("TANKER_CONFIG_NAME")
	Config               TankerConfig
	TestApp              *App
)

type App struct {
	AdminSession *core.Admin
	Descriptor   *core.AppDescriptor
	Config       TankerConfig
}

func CreateApp(config TankerConfig) (*App, error) {
	AdminSession, err := core.CreateAdmin(config.URL, config.IDToken)
	if err != nil {
		return nil, err
	}
	descriptor, err := AdminSession.CreateApp("sdk-go-test")
	if err != nil {
		return nil, err
	}
	return &App{AdminSession, descriptor, config}, nil
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
		Url:            app.Config.URL,
		UserID:         userID,
		Identity:       *userIdentity,
		PublicIdentity: *publicIdentity,
	}
}

type Device struct {
	User
	Path string
}

func loadConfig() (*TankerConfig, error) {
	file, err := os.Open(tankerConfigFilePath)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var configs map[string]interface{}
	err = json.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}
	config, ok := configs[tankerConfigName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid config name %s", tankerConfigName)
	}
	return &TankerConfig{URL: config["url"].(string), IDToken: config["idToken"].(string)}, nil
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
	config, err := loadConfig()
	if err != nil {
		Fail(err.Error())
	}
	TestApp, err = CreateApp(*config)
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
	if len(tankerConfigFilePath) == 0 || len(tankerConfigName) == 0 {
		panic("Tanker test config is invalid")
	}
	core.SetLogHandler(p)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Test Suite")
}
