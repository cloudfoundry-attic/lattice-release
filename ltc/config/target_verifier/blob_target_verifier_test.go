package target_verifier_test

import (
	"net"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyBlobTarget", func() {
		var (
			fakeServer     *ghttp.Server
			targetVerifier target_verifier.TargetVerifier
			statusCode     int
			responseBody   string
			targetInfo     dav_blob_store.Config
		)

		BeforeEach(func() {
			targetVerifier = target_verifier.New(nil)
			fakeServer = ghttp.NewServer()
			proxyURL, err := url.Parse(fakeServer.URL())
			Expect(err).NotTo(HaveOccurred())

			serverHost, serverPort, err := net.SplitHostPort(proxyURL.Host)
			Expect(err).NotTo(HaveOccurred())

			targetInfo = dav_blob_store.Config{
				Host:     serverHost,
				Port:     serverPort,
				Username: "some-username",
				Password: "some-password",
			}

			httpHeader := http.Header{
				http.CanonicalHeaderKey("Content-Type"): []string{"application/xml"},
			}
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PROPFIND", "/blobs"),
				ghttp.RespondWithPtr(&statusCode, &responseBody, httpHeader),
			))
		})

		It("returns ok=true if able to connect and auth and it exists", func() {
			statusCode = 207
			responseBody = `<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/"/>`

			err := targetVerifier.VerifyBlobTarget(targetInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if able to connect but can't auth", func() {
			statusCode = http.StatusForbidden

			err := targetVerifier.VerifyBlobTarget(targetInfo)
			Expect(err).To(MatchError("unauthorized"))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if able to connect but unknown status code", func() {
			statusCode = http.StatusTeapot

			err := targetVerifier.VerifyBlobTarget(targetInfo)
			Expect(err).To(HaveOccurred())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if the server is down", func() {
			listenerAddr := fakeServer.HTTPTestServer.Listener.Addr().String()
			fakeServer.Close()
			Eventually(func() error {
				_, err := net.Dial("tcp", listenerAddr)
				return err
			}).Should(HaveOccurred())

			err := targetVerifier.VerifyBlobTarget(targetInfo)
			Expect(err).To(MatchError(HavePrefix("blob target is down")))
		})
	})
})
