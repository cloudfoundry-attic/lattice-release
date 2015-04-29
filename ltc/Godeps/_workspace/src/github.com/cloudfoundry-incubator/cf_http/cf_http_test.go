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
			Expect(client.Timeout).To(Equal(timeout))
			transport := client.Transport.(*http.Transport)
			Expect(transport.Dial).NotTo(BeNil())
			Expect(transport.DisableKeepAlives).To(BeFalse())
		})
	})

	Describe("NewStreamingClient", func() {
		It("returns an http client", func() {
			client := cf_http.NewStreamingClient()
			Expect(client.Timeout).To(BeZero())
			transport := client.Transport.(*http.Transport)
			Expect(transport.Dial).NotTo(BeNil())
			Expect(transport.DisableKeepAlives).To(BeFalse())
		})
	})
})
