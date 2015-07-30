package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDavTool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DavTool Suite")
}
