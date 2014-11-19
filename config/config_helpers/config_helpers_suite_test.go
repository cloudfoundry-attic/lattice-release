package config_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfigHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConfigHelpers Suite")
}
