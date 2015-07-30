package dav_blob_store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDAVBlobStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DAVBlobStore Suite")
}
