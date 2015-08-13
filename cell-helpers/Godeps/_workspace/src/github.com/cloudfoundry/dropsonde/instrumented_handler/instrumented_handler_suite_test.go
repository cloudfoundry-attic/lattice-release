package instrumented_handler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInstrumentedHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstrumentedHandler Suite")
}
