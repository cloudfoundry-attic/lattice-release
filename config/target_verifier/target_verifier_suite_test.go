package target_verifier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTargetVerifier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TargetVerifier Suite")
}
