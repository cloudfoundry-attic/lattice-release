package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLatticeCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LatticeCli Suite")
}
