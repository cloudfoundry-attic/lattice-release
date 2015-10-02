package s3_blob_store_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/blob"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
)

type nullRetryer struct {
	service.DefaultRetryer
}

func (n nullRetryer) ShouldRetry(_ *request.Request) bool {
	return false
}

var _ = Describe("BlobStore", func() {
	var (
		blobStore  *s3_blob_store.BlobStore
		fakeServer *ghttp.Server
	)

	BeforeEach(func() {
		blobTargetInfo := config.S3BlobStoreConfig{
			AccessKey:  "some-access-key",
			SecretKey:  "some-secret-key",
			BucketName: "bucket",
			Region:     "some-s3-region",
		}

		blobStore = s3_blob_store.New(blobTargetInfo)
		blobStore.S3.Retryer = nullRetryer{}

		fakeServer = ghttp.NewServer()
		blobStore.S3.Endpoint = fakeServer.URL()
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	Describe(".New", func() {
		It("returns a new BlobStore with the provided credentials, region, and bucket", func() {
			Expect(*blobStore.S3.Config.Region).To(Equal("some-s3-region"))
			Expect(*blobStore.S3.Config.S3ForcePathStyle).To(BeTrue())
			Expect(blobStore.S3.Config.Credentials.Get()).To(Equal(credentials.Value{
				AccessKeyID:     "some-access-key",
				SecretAccessKey: "some-secret-key",
			}))
			Expect(blobStore.Bucket).To(Equal("bucket"))
		})
	})

	Describe("#List", func() {
		It("lists objects in a bucket", func() {
			responseBody := `
				 <?xml version="1.0" encoding="UTF-8"?>
				 <ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
					 <Name>bucket</Name>
					 <Prefix/>
					 <Marker/>
					 <MaxKeys>1000</MaxKeys>
					 <IsTruncated>false</IsTruncated>
					 <Contents>
						 <Key>my-image.jpg</Key>
						 <LastModified>2009-10-12T17:50:30.000Z</LastModified>
						 <ETag>&quot;fba9dede5f27731c9771645a39863328&quot;</ETag>
						 <Size>434234</Size>
						 <StorageClass>STANDARD</StorageClass>
						 <Owner>
							 <ID>75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a</ID>
							 <DisplayName>mtd@amazon.com</DisplayName>
						 </Owner>
					 </Contents>
					 <Contents>
						<Key>my-third-image.jpg</Key>
						  <LastModified>2009-10-12T17:50:30.000Z</LastModified>
						 <ETag>&quot;1b2cf535f27731c974343645a3985328&quot;</ETag>
						 <Size>64994</Size>
						 <StorageClass>STANDARD</StorageClass>
						 <Owner>
							 <ID>75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a</ID>
							 <DisplayName>mtd@amazon.com</DisplayName>
						 </Owner>
					 </Contents>
				 </ListBucketResult>
			 `

			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket"),
				ghttp.RespondWith(http.StatusOK, responseBody, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			expectedTime, err := time.Parse(time.RFC3339Nano, "2009-10-12T17:50:30.000Z")
			Expect(err).NotTo(HaveOccurred())

			Expect(blobStore.List()).To(Equal([]blob.Blob{
				{Path: "my-image.jpg", Size: 434234, Created: expectedTime},
				{Path: "my-third-image.jpg", Size: 64994, Created: expectedTime},
			}))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when we fail to retrieve the objects from S3", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket"),
				ghttp.RespondWith(http.StatusInternalServerError, nil, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			_, err := blobStore.List()
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
	})

	Describe("#Upload", func() {
		It("uploads the provided reader into the bucket", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/bucket/some-path/some-object"),
				ghttp.VerifyHeader(http.Header{"X-Amz-Acl": []string{"private"}}),
				func(_ http.ResponseWriter, request *http.Request) {
					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
				},
				ghttp.RespondWith(http.StatusOK, "", http.Header{}),
			))

			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when S3 fail to receive the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/bucket/some-path/some-object"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			err := blobStore.Upload("some-path/some-object", strings.NewReader("some data"))
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
	})

	Describe("#Download", func() {
		It("dowloads the requested path", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket/some-path/some-object"),
				ghttp.RespondWith(http.StatusOK, "some data", http.Header{"Content-Length": []string{"9"}}),
			))

			pathReader, err := blobStore.Download("some-path/some-object")
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.ReadAll(pathReader)).To(Equal([]byte("some data")))
			Expect(pathReader.Close()).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when S3 fails to retrieve the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket/some-path/some-object"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			_, err := blobStore.Download("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
	})

	Describe("#Delete", func() {
		It("deletes the object at the provided path", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/bucket/some-path/some-object"),
				ghttp.RespondWith(http.StatusNoContent, ""),
			))
			Expect(blobStore.Delete("some-path/some-object")).NotTo(HaveOccurred())
			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when S3 fails to delete the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/bucket/some-path/some-object"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			err := blobStore.Delete("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
	})

	Context("Droplet Actions", func() {
		Describe("#DownloadAppBitsAction", func() {
			It("constructs the correct Action to download app bits", func() {
				Expect(blobStore.DownloadAppBitsAction("droplet-name")).To(Equal(models.WrapAction(&models.SerialAction{
					Actions: []*models.Action{
						models.WrapAction(&models.RunAction{
							Path: "/tmp/s3tool",
							Dir:  "/",
							Args: []string{
								"get",
								"some-access-key",
								"some-secret-key",
								"bucket",
								"some-s3-region",
								"/droplet-name/bits.zip",
								"/tmp/bits.zip",
							},
							User: "vcap",
						}),
						models.WrapAction(&models.RunAction{
							Path: "/bin/mkdir",
							Args: []string{"/tmp/app"},
							User: "vcap",
						}),
						models.WrapAction(&models.RunAction{
							Path: "/usr/bin/unzip",
							Dir:  "/tmp/app",
							Args: []string{"-q", "/tmp/bits.zip"},
							User: "vcap",
						}),
					},
				})))
			})
		})

		Describe("#DeleteAppBitsAction", func() {
			It("constructs the correct Action to delete app bits", func() {
				Expect(blobStore.DeleteAppBitsAction("droplet-name")).To(Equal(models.WrapAction(&models.RunAction{
					Path: "/tmp/s3tool",
					Dir:  "/",
					Args: []string{
						"delete",
						"some-access-key",
						"some-secret-key",
						"bucket",
						"some-s3-region",
						"/droplet-name/bits.zip",
					},
					User: "vcap",
				})))
			})
		})

		Describe("#UploadDropletAction", func() {
			It("constructs the correct Action to upload the droplet", func() {
				Expect(blobStore.UploadDropletAction("droplet-name")).To(Equal(models.WrapAction(&models.RunAction{
					Path: "/tmp/s3tool",
					Dir:  "/",
					Args: []string{
						"put",
						"some-access-key",
						"some-secret-key",
						"bucket",
						"some-s3-region",
						"/droplet-name/droplet.tgz",
						"/tmp/droplet",
					},
					User: "vcap",
				})))
			})
		})

		Describe("#UploadDropletMetadataAction", func() {
			It("constructs the correct Action to upload the droplet metadata", func() {
				Expect(blobStore.UploadDropletMetadataAction("droplet-name")).To(Equal(models.WrapAction(&models.RunAction{
					Path: "/tmp/s3tool",
					Dir:  "/",
					Args: []string{
						"put",
						"some-access-key",
						"some-secret-key",
						"bucket",
						"some-s3-region",
						"/droplet-name/result.json",
						"/tmp/result.json",
					},
					User: "vcap",
				})))
			})
		})

		Describe("#DownloadDropletAction", func() {
			It("constructs the correct Action to download the droplet", func() {
				Expect(blobStore.DownloadDropletAction("droplet-name")).To(Equal(models.WrapAction(&models.SerialAction{
					Actions: []*models.Action{
						models.WrapAction(&models.RunAction{
							Path: "/tmp/s3tool",
							Dir:  "/",
							Args: []string{
								"get",
								"some-access-key",
								"some-secret-key",
								"bucket",
								"some-s3-region",
								"/droplet-name/droplet.tgz",
								"/tmp/droplet.tgz",
							},
							User: "vcap",
						}),
						models.WrapAction(&models.RunAction{
							Path: "/bin/tar",
							Args: []string{"zxf", "/tmp/droplet.tgz"},
							Dir:  "/home/vcap",
							User: "vcap",
						}),
					},
				})))
			})
		})
	})
})
