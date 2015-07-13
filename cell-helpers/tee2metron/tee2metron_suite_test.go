package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTee2metron(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tee2metron Suite")
}
