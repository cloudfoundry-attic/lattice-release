package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppRunnerCommandFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppRunner CommandFactory Suite")
}
