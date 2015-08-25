package s3_blob_store_test

// import (
// 	"io/ioutil"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/onsi/gomega/ghttp"

// 	"github.com/aws/aws-sdk-go/aws"
// 	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
// 	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
// )

// var _ = Describe("BlobStore", func() {
// 	var (
// 		blobStore  *s3_blob_store.BlobStore
// 		fakeServer *ghttp.Server
// 	)

// 	BeforeEach(func() {
// 		fakeServer = ghttp.NewServer()
// 		fakeServerURL, err := url.Parse(fakeServer.URL())
// 		Expect(err).NotTo(HaveOccurred())

// 		serverHost, serverPort, err := net.SplitHostPort(fakeServerURL.Host)
// 		Expect(err).NotTo(HaveOccurred())

// 		blobTargetInfo := config_package.S3BlobStoreConfig{
// 			Host:       serverHost,
// 			Port:       serverPort,
// 			AccessKey:  "V8GDQFR_VDOGM55IV8OH",
// 			SecretKey:  "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==",
// 			BucketName: "bucket",
// 		}

// 		blobStore = s3_blob_store.New(blobTargetInfo)
// 		blobStore.S3.ShouldRetry = func(_ *aws.Request) bool { return false }
// 	})

// 	AfterEach(func() {
// 		fakeServer.Close()
// 	})

// 	Describe(".New", func() {
// 		It("returns a BlobStore for the riak-region-1 region", func() {
// 			Expect(*blobStore.S3.Config.Region).To(Equal("riak-region-1"))
// 		})

// 		It("returns a BlobStore that signs requests and includes a content-length header", func() {
// 			requestTime, err := time.Parse(time.RFC1123, "Tue, 21 Jul 2015 23:09:05 UTC")
// 			Expect(err).NotTo(HaveOccurred())
// 			request := &http.Request{
// 				Header: http.Header{"some-header": []string{"some-value"}},
// 				URL:    &url.URL{Scheme: "http", Opaque: "//some-host/some-bucket", RawQuery: "some-param=some-value"},
// 			}
// 			blobStore.S3.Handlers.Sign.Run(&aws.Request{Time: requestTime, HTTPRequest: request})
// 			Expect(request.Header).To(Equal(http.Header{
// 				"Host":           {"some-host"},
// 				"Date":           {"Tue, 21 Jul 2015 23:09:05 UTC"},
// 				"Authorization":  {"AWS V8GDQFR_VDOGM55IV8OH:6WPghcgpPKDq70e4x3vPBZOwiqg="},
// 				"Content-Length": {"0"},
// 				"some-header":    {"some-value"},
// 			}))
// 			Expect(request.URL.String()).To(Equal("http://some-host/some-bucket?some-param=some-value"))
// 		})
// 	})

// 	Describe("#List", func() {
// 		It("lists objects in a bucket", func() {
// 			responseBody := `
// 				<?xml version="1.0" encoding="UTF-8"?>
// 				<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
// 					<Name>bucket</Name>
// 					<Prefix/>
// 					<Marker/>
// 					<MaxKeys>1000</MaxKeys>
// 					<IsTruncated>false</IsTruncated>
// 					<Contents>
// 						<Key>my-image.jpg</Key>
// 						<LastModified>2009-10-12T17:50:30.000Z</LastModified>
// 						<ETag>&quot;fba9dede5f27731c9771645a39863328&quot;</ETag>
// 						<Size>434234</Size>
// 						<StorageClass>STANDARD</StorageClass>
// 						<Owner>
// 							<ID>75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a</ID>
// 							<DisplayName>mtd@amazon.com</DisplayName>
// 						</Owner>
// 					</Contents>
// 					<Contents>
// 					   <Key>my-third-image.jpg</Key>
// 						 <LastModified>2009-10-12T17:50:30.000Z</LastModified>
// 						<ETag>&quot;1b2cf535f27731c974343645a3985328&quot;</ETag>
// 						<Size>64994</Size>
// 						<StorageClass>STANDARD</StorageClass>
// 						<Owner>
// 							<ID>75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a</ID>
// 							<DisplayName>mtd@amazon.com</DisplayName>
// 						</Owner>
// 					</Contents>
// 				</ListBucketResult>
// 			`

// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("GET", "/bucket"),
// 				ghttp.RespondWith(http.StatusOK, responseBody, http.Header{"Content-Type": []string{"application/xml"}}),
// 			))

// 			expectedTime, err := time.Parse(time.RFC3339Nano, "2009-10-12T17:50:30.000Z")
// 			Expect(err).NotTo(HaveOccurred())

// 			Expect(blobStore.List()).To(Equal([]s3_blob_store.Blob{
// 				{Path: "my-image.jpg", Size: 434234, Created: expectedTime},
// 				{Path: "my-third-image.jpg", Size: 64994, Created: expectedTime},
// 			}))

// 			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
// 		})

// 		It("returns an error when we fail to retrieve the objects from S3", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("GET", "/bucket"),
// 				ghttp.RespondWith(http.StatusInternalServerError, nil, http.Header{"Content-Type": []string{"application/xml"}}),
// 			))

// 			_, err := blobStore.List()
// 			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
// 		})
// 	})

// 	Describe("#Upload", func() {
// 		It("uploads the provided reader into the bucket", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("PUT", "/bucket/some-path/some-object"),
// 				ghttp.VerifyHeader(http.Header{"X-Amz-Acl": []string{"private"}}),
// 				func(_ http.ResponseWriter, request *http.Request) {
// 					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
// 				},
// 				ghttp.RespondWith(http.StatusOK, "", http.Header{}),
// 			))

// 			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

// 			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
// 		})

// 		It("returns an error when S3 fail to receive the object", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("PUT", "/bucket/some-path/some-object"),
// 				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
// 			))

// 			err := blobStore.Upload("some-path/some-object", strings.NewReader("some data"))
// 			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
// 		})
// 	})

// 	Describe("#Download", func() {
// 		It("dowloads the requested path", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("GET", "/bucket/some-path/some-object"),
// 				ghttp.RespondWith(http.StatusOK, "some data", http.Header{"Content-length": []string{"9"}}),
// 			))

// 			pathReader, err := blobStore.Download("some-path/some-object")
// 			Expect(err).NotTo(HaveOccurred())
// 			Expect(ioutil.ReadAll(pathReader)).To(Equal([]byte("some data")))
// 			Expect(pathReader.Close()).To(Succeed())

// 			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
// 		})

// 		It("returns an error when S3 fails to retrieve the object", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("GET", "/bucket/some-path/some-object"),
// 				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
// 			))

// 			_, err := blobStore.Download("some-path/some-object")
// 			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
// 		})
// 	})

// 	Describe("#Delete", func() {
// 		It("deletes the object at the provided path", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("DELETE", "/bucket/some-path/some-object"),
// 				ghttp.RespondWith(http.StatusNoContent, ""),
// 			))
// 			Expect(blobStore.Delete("some-path/some-object")).NotTo(HaveOccurred())
// 			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
// 		})

// 		It("returns an error when S3 fails to delete the object", func() {
// 			fakeServer.AppendHandlers(ghttp.CombineHandlers(
// 				ghttp.VerifyRequest("DELETE", "/bucket/some-path/some-object"),
// 				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
// 			))

// 			err := blobStore.Delete("some-path/some-object")
// 			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
// 		})
// 	})
// })
