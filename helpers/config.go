package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
)

type TankerConfig struct {
	URL     string `json:"url"`
	IDToken string `json:"idToken"`
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

	idConfig := identity.Config{AppID: descriptor.ID, AppSecret: descriptor.PrivateKey}
	return &App{AdminSession, descriptor, config, idConfig}, nil
}

func LoadConfig(configFilePath string, configName string) (*TankerConfig, error) {
	file, err := os.Open(configFilePath)
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
	config, ok := configs[configName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid config name %s", configName)
	}
	return &TankerConfig{URL: config["url"].(string), IDToken: config["idToken"].(string)}, nil
}
