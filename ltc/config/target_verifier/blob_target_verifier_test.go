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

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyBlobTarget", func() {
		var (
			fakeServer     *ghttp.Server
			targetVerifier target_verifier.TargetVerifier
			statusCode     int
			responseBody   string
			targetInfo     config.BlobTargetInfo
		)

		BeforeEach(func() {
			targetVerifier = target_verifier.New(nil)
			fakeServer = ghttp.NewServer()
			proxyURL, err := url.Parse(fakeServer.URL())
			Expect(err).NotTo(HaveOccurred())
			proxyHostArr := strings.Split(proxyURL.Host, ":")
			Expect(proxyHostArr).To(HaveLen(2))
			proxyHostPort, err := strconv.Atoi(proxyHostArr[1])
			Expect(err).NotTo(HaveOccurred())

			targetInfo = config.BlobTargetInfo{
				TargetHost: proxyHostArr[0],
				TargetPort: uint16(proxyHostPort),
				AccessKey:  "some-access-key",
				SecretKey:  "some-secret-key",
				BucketName: "bucket",
			}

			httpHeader := http.Header{
				http.CanonicalHeaderKey("Content-Type"): []string{"application/xml"},
			}
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket"),
				ghttp.RespondWithPtr(&statusCode, &responseBody, httpHeader),
			))
		})

		It("returns ok=true if able to connect and auth and it exists", func() {
			statusCode = http.StatusOK
			responseBody = `
				<?xml version="1.0" encoding="UTF-8"?>
				<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
					<Name>bucket</Name>
					<Prefix/>
					<Marker/>
					<MaxKeys>1000</MaxKeys>
					<IsTruncated>false</IsTruncated>
				</ListBucketResult>
			`

			err := targetVerifier.VerifyBlobTarget(targetInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns ok=false if able to connect but can't auth", func() {
			statusCode = http.StatusForbidden
			responseBody = `
				<?xml version="1.0" encoding="UTF-8"?>
				<Error>
				  <Code>NoSuchKey</Code>
				  <Message>The resource you requested does not exist</Message>
				  <Resource>/bucket</Resource>
				  <RequestId>4442587FB7D0A2F9</RequestId>
				</Error>
			`

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
