package setup_cli_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSetupCliHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SetupCliHelpers Suite")
}
