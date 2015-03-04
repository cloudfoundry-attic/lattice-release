package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTee2metron(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tee2metron Suite")
}
