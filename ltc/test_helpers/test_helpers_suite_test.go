package test_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestHelpers Suite")
}
