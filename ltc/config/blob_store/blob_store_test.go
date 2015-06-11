package blob_store_test

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

var _ = Describe("BlobStore", func() {
	var (
		config     *config_package.Config
		httpClient *http.Client
		blobStore  blob_store.BlobStore
		fakeServer *ghttp.Server
		awsRegion  = aws.Region{Name: "faux-region-1", S3Endpoint: "http://s3.amazonaws.com"}
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		config = config_package.New(persister.NewMemPersister())

		fakeServerURL, err := url.Parse(fakeServer.URL())
		Expect(err).NotTo(HaveOccurred())
		proxyHostArr := strings.Split(fakeServerURL.Host, ":")
		Expect(proxyHostArr).To(HaveLen(2))
		proxyHostPort, err := strconv.Atoi(proxyHostArr[1])
		Expect(err).NotTo(HaveOccurred())
		config.SetBlobTarget(proxyHostArr[0], uint16(proxyHostPort), "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==", "buck")

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: config.BlobTarget().Proxy(),
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		}

		s3Auth := aws.Auth{
			AccessKey: config.BlobTarget().AccessKey,
			SecretKey: config.BlobTarget().SecretKey,
		}

		s3S3 := s3.New(s3Auth, awsRegion, httpClient)
		blobStore = blob_store.NewBlobStore(config, s3S3)
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	Describe("BlobStore", func() {
		Describe("Bucket", func() {
			It("returns a bucket", func() {
				fakeServer.AppendHandlers(ghttp.VerifyRequest("GET", "/v1/desired_lrps", "domain=diego"))

				bucket := blobStore.Bucket("whats up")
				Expect(bucket).ToNot(BeNil())

				bucket.PutReader("sdflj", gbytes.NewBuffer(), int64(22), "text/plain", s3.ACL("sdfljk"), s3.Options{})
			})
		})
	})
})
