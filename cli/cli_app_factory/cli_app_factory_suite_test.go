package cli_app_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCliAppFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CliAppFactory Suite")
}
