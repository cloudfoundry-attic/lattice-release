package exit_handler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestExitHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExitHandler Suite")
}
