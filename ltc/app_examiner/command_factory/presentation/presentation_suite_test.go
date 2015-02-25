package presentation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPresentation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Presentation Suite")
}
