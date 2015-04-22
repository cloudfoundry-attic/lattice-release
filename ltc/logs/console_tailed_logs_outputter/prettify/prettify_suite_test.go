package prettify_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPrettify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prettify Suite")
}
