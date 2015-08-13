package envelope_extensions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEnvelopeExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnvelopeExtensions Suite")
}
