package blob_store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("BlobStoreManager", func() {
	Describe(".New", func() {
		var blobStore blob_store.BlobStore

		Context("when a dav blob store is targeted", func() {
			BeforeEach(func() {
				config := config.New(nil)
				config.SetBlobStore("some-host", "some-port", "some-user", "some-password")
				blobStore = blob_store.New(config)
			})

			It("returns a new DavBlobStore object", func() {
				davBlobStore, ok := blobStore.(*dav_blob_store.BlobStore)
				Expect(ok).To(BeTrue())
				Expect(davBlobStore.URL.String()).To(Equal("http://some-user:some-password@some-host:some-port"))
			})
		})

		Context("when an s3 blob store is targeted", func() {
			BeforeEach(func() {
				config := config.New(nil)
				config.SetS3BlobStore("", "", "some-bucket-name", "")
				blobStore = blob_store.New(config)
			})

			It("returns a new S3BlobStore object", func() {
				s3BlobStore, ok := blobStore.(*s3_blob_store.BlobStore)
				Expect(ok).To(BeTrue())
				Expect(s3BlobStore.Bucket).To(Equal("some-bucket-name"))
			})
		})
	})

	Describe(".NewVerifier", func() {
		var verifier blob_store.Verifier

		Context("when a dav blob store is targeted", func() {
			BeforeEach(func() {
				config := config.New(nil)
				config.SetBlobStore("some-host", "some-port", "some-user", "some-password")
				verifier = blob_store.NewVerifier(config)
			})

			It("returns a new DavBlobStore Verifier", func() {
				_, ok := verifier.(dav_blob_store.Verifier)
				Expect(ok).To(BeTrue())
			})
		})

		Context("when an s3 blob store is targeted", func() {
			BeforeEach(func() {
				config := config.New(nil)
				config.SetS3BlobStore("", "", "some-bucket-name", "")
				verifier = blob_store.NewVerifier(config)
			})

			It("returns a new S3BlobStore Verifier", func() {
				_, ok := verifier.(*s3_blob_store.Verifier)
				Expect(ok).To(BeTrue())
			})
		})
	})
})
