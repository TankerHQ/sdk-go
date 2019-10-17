package core_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"
)

var (
	tankerConfigFilePath = os.Getenv("TANKER_CONFIG_FILEPATH")
	tankerConfigName     = os.Getenv("TANKER_CONFIG_NAME")
	Config               helpers.TestConfig
	TestApp              *helpers.App
)

var _ = BeforeSuite(func() {
	config, err := helpers.LoadConfig(tankerConfigFilePath, tankerConfigName)
	if err != nil {
		Fail(err.Error())
	}
	TestApp, err = helpers.CreateApp(config.Server)
	if err != nil {
		Fail(err.Error())
	}
})

var _ = AfterSuite(func() {
	err := TestApp.Destroy()
	if err != nil {
		Fail(err.Error())
	}
})

func TestSDK(t *testing.T) {
	if len(tankerConfigFilePath) == 0 || len(tankerConfigName) == 0 {
		panic("Tanker test config is invalid")
	}
	core.SetLogHandler(func(record core.LogRecord) {
		fmt.Printf("[%c]{%s}'%s+%d': %s\n", record.Level, record.Category, record.File, record.Line, record.Message)
	})
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Test Suite")
}
