package cf_http_test

import (
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cf_http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfHttp", func() {
	var timeout time.Duration

	BeforeEach(func() {
		timeout = 1 * time.Second
	})

	JustBeforeEach(func() {
		cf_http.Initialize(timeout)
	})

	Describe("NewClient", func() {
		It("returns an http client", func() {
			client := cf_http.NewClient()
			Ω(client.Timeout).Should(Equal(timeout))
			transport := client.Transport.(*http.Transport)
			Ω(transport.Dial).ShouldNot(BeNil())
			Ω(transport.DisableKeepAlives).Should(BeFalse())
		})
	})

	Describe("NewStreamingClient", func() {
		It("returns an http client", func() {
			client := cf_http.NewStreamingClient()
			Ω(client.Timeout).Should(BeZero())
			transport := client.Transport.(*http.Transport)
			Ω(transport.Dial).ShouldNot(BeNil())
			Ω(transport.DisableKeepAlives).Should(BeFalse())
		})
	})
})
