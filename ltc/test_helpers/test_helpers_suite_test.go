package test_helpers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestHelpers Suite")
}
