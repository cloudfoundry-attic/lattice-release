package instrumented_round_tripper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInstrumentedRoundTripper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstrumentedRoundTripper Suite")
}
