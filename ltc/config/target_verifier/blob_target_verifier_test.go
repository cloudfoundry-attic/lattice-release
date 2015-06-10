package target_verifier_test

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyBlobTarget", func() {
		var (
			config         *config_package.Config
			fakeServer     *ghttp.Server
			httpClient     *http.Client
			targetVerifier target_verifier.TargetVerifier
			blobStore      blob_store.BlobStore
			awsRegion      = aws.Region{Name: "faux-region-1", S3Endpoint: "http://s3.amazonaws.com"}
			statusCode     int
		)

		BeforeEach(func() {
			targetVerifier = target_verifier.New(func(string) receptor.Client {
				return &fake_receptor.FakeClient{}
			})
			fakeServer = ghttp.NewServer()

			config = config_package.New(persister.NewMemPersister())
			proxyURL, err := url.Parse(fakeServer.URL())
			Expect(err).NotTo(HaveOccurred())
			proxyHostArr := strings.Split(proxyURL.Host, ":")
			Expect(proxyHostArr).To(HaveLen(2))
			proxyHostPort, _ := strconv.Atoi(proxyHostArr[1])
			config.SetBlobTarget(proxyHostArr[0], uint16(proxyHostPort), "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==")

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

			responseBody := ""
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("HEAD", "/condenser-bucket/verifier/invalid-path"),
				ghttp.RespondWithPtr(&statusCode, &responseBody),
			))
		})

		It("returns up=true, auth=true if able to connect and auth", func() {
			statusCode = http.StatusNotFound

			up, auth, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==")

			Expect(up).To(BeTrue())
			Expect(auth).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns up=true, auth=true if able to connect and auth and it exists", func() {
			statusCode = http.StatusOK

			up, auth, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==")

			Expect(up).To(BeTrue())
			Expect(auth).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns up=true, auth=false if able to connect but can't auth", func() {
			statusCode = http.StatusForbidden

			up, auth, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==")

			Expect(up).To(BeTrue())
			Expect(auth).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns up=false, auth=false, err=(the bubbled up error) if there is a non-receptor error", func() {
			fakeServer.Close()

			up, auth, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==")

			Expect(up).To(BeFalse())
			Expect(auth).To(BeFalse())
			Expect(err).To(HaveOccurred())
		})
	})
})
