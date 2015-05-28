package setup_cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSetupCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SetupCli Suite")
}
