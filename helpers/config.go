package helpers

import (
	"fmt"
	"os"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
)

type ServerConfig struct {
	AdminURL string
	IDToken  string
	URL      string
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

func safeGetEnv(key string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("%s is not set\n", key))
	}
	return v
}

func NewApp(testConfig TestConfig) (*App, error) {
	AdminSession, err := core.NewAdmin(testConfig.Server.AdminURL, testConfig.Server.IDToken)
	if err != nil {
		return nil, err
	}
	descriptor, err := AdminSession.NewApp("sdk-go-test")
	if err != nil {
		return nil, err
	}

	idConfig := identity.Config{AppID: descriptor.ID, AppSecret: descriptor.PrivateKey}
	return &App{AdminSession, descriptor, testConfig.Server, testConfig.Oidc, idConfig}, nil
}

func LoadConfig() (*TestConfig, error) {
	serverConfig := ServerConfig{
		AdminURL: safeGetEnv("TANKER_ADMIND_URL"),
		IDToken:  safeGetEnv("TANKER_ID_TOKEN"),
		URL:      safeGetEnv("TANKER_TRUSTCHAIND_URL"),
	}
	users := map[string]UserConfig{
		"martine": UserConfig{
			Email:        safeGetEnv("TANKER_OIDC_MARTINE_EMAIL"),
			RefreshToken: safeGetEnv("TANKER_OIDC_MARTINE_REFRESH_TOKEN"),
		},
		"kevin": UserConfig{
			Email:        safeGetEnv("TANKER_OIDC_KEVIN_EMAIL"),
			RefreshToken: safeGetEnv("TANKER_OIDC_KEVIN_REFRESH_TOKEN"),
		},
	}

	oidc := OidcConfig{
		ClientSecret: safeGetEnv("TANKER_OIDC_CLIENT_SECRET"),
		ClientId:     safeGetEnv("TANKER_OIDC_CLIENT_ID"),
		Provider:     safeGetEnv("TANKER_OIDC_PROVIDER"),
		Users:        users,
	}
	return &TestConfig{serverConfig, oidc}, nil
}
