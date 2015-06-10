package blob_store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBlobStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BlobStore Suite")
}
