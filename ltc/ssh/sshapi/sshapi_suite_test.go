package sshapi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSshapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSHAPI Suite")
}
