package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogsCommandFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logs CommandFactory Suite")
}
