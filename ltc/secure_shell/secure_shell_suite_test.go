package secure_shell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSecureShell(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecureShell Suite")
}
