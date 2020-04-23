package core_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"
)

var (
	Config  helpers.TestConfig
	TestApp *helpers.App
)

var _ = BeforeSuite(func() {
	Config, err := helpers.LoadConfig()
	if err != nil {
		Fail(err.Error())
	}
	TestApp, err = helpers.NewApp(*Config)
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
	core.SetLogHandler(func(record core.LogRecord) {
		fmt.Printf("[%c]{%s}'%s+%d': %s\n", record.Level, record.Category, record.File, record.Line, record.Message)
	})
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Test Suite")
}
