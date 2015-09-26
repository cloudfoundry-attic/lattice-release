package droplet_runner_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("HTTPProxyConfReader", func() {
	var (
		proxyConfReader *droplet_runner.HTTPProxyConfReader
		fakeServer      *ghttp.Server
		badURL          string
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()

		badServer := ghttp.NewServer()
		badURL = badServer.URL()
		badServer.Close()

		proxyConfReader = &droplet_runner.HTTPProxyConfReader{
			URL: fakeServer.URL() + "/pc.json",
		}
	})

	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	Context("#ProxyConf", func() {
		It("should parse JSON on 200", func() {
			fakeServer.RouteToHandler("GET", "/pc.json", ghttp.CombineHandlers(
				ghttp.RespondWith(200,
					`{"http_proxy": "http://proxy", "https_proxy": "https://proxy", "no_proxy": "no-proxy"}`,
					http.Header{"Content-Type": []string{"application/json"}},
				),
			))

			proxyConf, err := proxyConfReader.ProxyConf()
			Expect(err).NotTo(HaveOccurred())

			Expect(proxyConf.HTTPProxy).To(Equal("http://proxy"))
			Expect(proxyConf.HTTPSProxy).To(Equal("https://proxy"))
			Expect(proxyConf.NoProxy).To(Equal("no-proxy"))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("should fail when it receives invalid JSON on 200", func() {
			fakeServer.RouteToHandler("GET", "/pc.json", ghttp.CombineHandlers(
				ghttp.RespondWith(200, `{`, http.Header{"Content-Type": []string{"application/json"}}),
			))

			_, err := proxyConfReader.ProxyConf()
			Expect(err).To(MatchError("unexpected end of JSON input"))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("should return an empty ProxyConf on 404", func() {
			fakeServer.RouteToHandler("GET", "/pc.json", ghttp.CombineHandlers(
				ghttp.RespondWith(404, "/pc.json not found"),
			))

			proxyConf, err := proxyConfReader.ProxyConf()
			Expect(err).NotTo(HaveOccurred())

			Expect(proxyConf.HTTPProxy).To(BeEmpty())
			Expect(proxyConf.HTTPSProxy).To(BeEmpty())
			Expect(proxyConf.NoProxy).To(BeEmpty())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("should return an error on any other HTTP status", func() {
			fakeServer.RouteToHandler("GET", "/pc.json", ghttp.CombineHandlers(
				ghttp.RespondWith(500, "fail"),
			))

			_, err := proxyConfReader.ProxyConf()
			Expect(err).To(MatchError("500 Internal Server Error"))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("should return an error on any other HTTP status", func() {
			proxyConfReader.URL = badURL

			_, err := proxyConfReader.ProxyConf()
			Expect(err.Error()).To(MatchRegexp(`dial tcp 127\.0\.0\.1:[0-9]+: getsockopt: connection refused`))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(0))
		})
	})
})
