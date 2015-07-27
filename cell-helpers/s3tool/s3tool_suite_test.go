package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestS3Tool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3Tool Suite")
}
