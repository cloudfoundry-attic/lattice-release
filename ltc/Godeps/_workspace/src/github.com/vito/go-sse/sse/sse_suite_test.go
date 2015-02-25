package sse_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSSE(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSE Suite")
}
