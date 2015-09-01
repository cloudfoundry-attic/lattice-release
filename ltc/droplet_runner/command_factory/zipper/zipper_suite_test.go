package zipper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZipper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zipper Suite")
}
