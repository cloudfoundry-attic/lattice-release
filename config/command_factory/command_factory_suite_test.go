package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfigCommandFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config CommandFactory Suite")
}
