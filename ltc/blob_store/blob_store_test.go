package blob_store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/fake_blob_store_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

//go:generate counterfeiter -o fake_blob_store_verifier/fake_blob_store_verifier.go . Verifier
type Verifier interface {
	Verify(config *config_package.Config) (authorized bool, err error)
}

var _ = Describe("BlobStoreManager", func() {
	Describe(".New", func() {
		var blobStore blob_store.BlobStore

		Context("when a dav blob store is targeted", func() {
			BeforeEach(func() {
				config := config_package.New(nil)
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
				config := config_package.New(nil)
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

	Describe("#Verify", func() {
		var (
			verifier        blob_store.BlobStoreVerifier
			fakeDAVVerifier *fake_blob_store_verifier.FakeVerifier
			fakeS3Verifier  *fake_blob_store_verifier.FakeVerifier
		)

		BeforeEach(func() {
			fakeDAVVerifier = &fake_blob_store_verifier.FakeVerifier{}
			fakeS3Verifier = &fake_blob_store_verifier.FakeVerifier{}
			verifier = blob_store.BlobStoreVerifier{
				DAVBlobStoreVerifier: fakeDAVVerifier,
				S3BlobStoreVerifier:  fakeS3Verifier,
			}
		})

		Context("when a dav blob store is targeted", func() {
			var config *config_package.Config

			BeforeEach(func() {
				config = config_package.New(nil)
				config.SetBlobStore("some-host", "some-port", "some-user", "some-password")
			})

			It("returns a new DavBlobStore Verifier", func() {
				verifier.Verify(config)
				Expect(fakeDAVVerifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeS3Verifier.VerifyCallCount()).To(Equal(0))
			})
		})

		Context("when an s3 blob store is targeted", func() {
			var config *config_package.Config

			BeforeEach(func() {
				config = config_package.New(nil)
				config.SetS3BlobStore("", "", "some-bucket-name", "")
			})

			It("returns a new S3BlobStore Verifier", func() {
				verifier.Verify(config)
				Expect(fakeS3Verifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeDAVVerifier.VerifyCallCount()).To(Equal(0))
			})
		})
	})
})
