package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestS3downloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3deleter Suite")
}
