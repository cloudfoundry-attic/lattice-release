package keygen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKeygen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keygen Suite")
}
