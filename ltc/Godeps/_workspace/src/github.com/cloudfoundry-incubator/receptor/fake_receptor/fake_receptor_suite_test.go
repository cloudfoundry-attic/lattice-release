package fake_receptor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFakeReceptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FakeReceptor Suite")
}
