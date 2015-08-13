package logemitter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEmitter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Log Emitter Suite")
}
