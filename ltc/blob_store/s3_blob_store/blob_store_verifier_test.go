package s3_blob_store_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("S3BlobStore", func() {
	var (
		verifier   *s3_blob_store.Verifier
		fakeServer *ghttp.Server
		config     *config_package.Config
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		verifier = &s3_blob_store.Verifier{fakeServer.URL()}
		config = config_package.New(nil)
		config.SetS3BlobStore("some-access-key", "some-secret-key", "bucket", "some-region")
	})

	Describe("Verify", func() {
		Context("when the blob store credentials are valid", func() {
			It("returns authorized as true", func() {
				responseBody := `
					 <?xml version="1.0" encoding="UTF-8"?>
					 <ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
						 <Name>bucket</Name>
						 <Prefix/>
						 <Marker/>
						 <MaxKeys>1000</MaxKeys>
						 <IsTruncated>false</IsTruncated>
					 </ListBucketResult>
				 `

				fakeServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/bucket"),
					ghttp.RespondWith(http.StatusOK, responseBody, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				authorized, err := verifier.Verify(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(authorized).To(BeTrue())

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the blob store credentials are incorrect", func() {
			It("returns authorized as false", func() {
				fakeServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/bucket"),
					ghttp.RespondWith(http.StatusForbidden, nil),
				))

				authorized, err := verifier.Verify(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(authorized).To(BeFalse())

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the blob store is inaccessible", func() {
			It("returns an error", func() {
				fakeServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/bucket"),
					ghttp.RespondWith(http.StatusBadRequest, nil),
				))
				_, err := verifier.Verify(config)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
