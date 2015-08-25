package s3_blob_store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestS3BlobStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3BlobStore Suite")
}
