package cf_ignore_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCfIgnore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CFIgnore Suite")
}
