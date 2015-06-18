package target_verifier_test

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyBlobTarget", func() {
		var (
			config         *config_package.Config
			fakeServer     *ghttp.Server
			targetVerifier target_verifier.TargetVerifier
			statusCode     int
			responseBody   string
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
			config.SetBlobTarget(proxyHostArr[0], uint16(proxyHostPort), "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==", "bucket")

			httpHeader := http.Header{
				"Content-Type": []string{"application/xml"},
			}

			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket/"),
				ghttp.RespondWithPtr(&statusCode, &responseBody, httpHeader),
			))
		})

		It("returns ok=true if able to connect and auth and it exists", func() {
			statusCode = http.StatusOK
			responseBody = `<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Owner><ID>x</ID><DisplayName>x</DisplayName></Owner><Buckets><Bucket><Name>bucket</Name><CreationDate>2015-06-11T16:50:43.000Z</CreationDate></Bucket></Buckets></ListAllMyBucketsResult>`

			ok, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==", "bucket")

			Expect(ok).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if able to connect but can't auth", func() {
			statusCode = http.StatusForbidden

			ok, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==", "bucket")

			Expect(ok).To(BeFalse())
			Expect(err).To(MatchError("unauthorized"))
			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if the server is down", func() {
			listenerAddr := fakeServer.HTTPTestServer.Listener.Addr().String()
			fakeServer.Close()

			Eventually(func() error {
				_, err := net.Dial("tcp", listenerAddr)
				return err
			}).Should(HaveOccurred())

			ok, err := targetVerifier.VerifyBlobTarget(config.BlobTarget().TargetHost, config.BlobTarget().TargetPort, "V8GDQFR_VDOGM55IV8OH", "Wv_kltnl98hNWNdNwyQPYnFhK4gVPTxVS3NNMg==", "bucket")

			Expect(ok).To(BeFalse())
			Expect(err).To(MatchError(HavePrefix("blob target is down")))
		})
	})
})
