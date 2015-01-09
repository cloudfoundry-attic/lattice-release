package cursor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCursor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cursor Suite")
}
