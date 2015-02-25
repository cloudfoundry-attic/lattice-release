package colors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestColors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Colors Suite")
}
