package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
)

type ServerConfig struct {
	URL     string `json:"url"`
	IDToken string `json:"idToken"`
}

type UserConfig struct {
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
}

type OidcConfig struct {
	ClientSecret string                `json:"clientSecret"`
	ClientId     string                `json:"clientId"`
	Provider     string                `json:"provider"`
	Users        map[string]UserConfig `json:"users"`
}

type TestConfig struct {
	Server ServerConfig
	Oidc   OidcConfig
}

func CreateApp(config ServerConfig) (*App, error) {
	AdminSession, err := core.CreateAdmin(config.URL, config.IDToken)
	if err != nil {
		return nil, err
	}
	descriptor, err := AdminSession.CreateApp("sdk-go-test")
	if err != nil {
		return nil, err
	}

	idConfig := identity.Config{AppID: descriptor.ID, AppSecret: descriptor.PrivateKey}
	return &App{AdminSession, descriptor, config, idConfig}, nil
}

func LoadConfig(configFilePath string, configName string) (*TestConfig, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var configs map[string]json.RawMessage
	err = json.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}
	var serverConfig ServerConfig
	if err = json.Unmarshal(configs[configName], &serverConfig); err != nil {
		return nil, fmt.Errorf("Invalid config name %s", configName)
	}
	var oidcs map[string]json.RawMessage
	if err = json.Unmarshal(configs["oidc"], &oidcs); err != nil {
		return nil, fmt.Errorf("No valid oidc config found")
	}
	var oidc OidcConfig
	if err = json.Unmarshal(oidcs["googleAuth"], &oidc); err != nil {
		return nil, fmt.Errorf("Invalid oidc config name %s", "googleAuth")
	}
	return &TestConfig{serverConfig, oidc}, nil
}
