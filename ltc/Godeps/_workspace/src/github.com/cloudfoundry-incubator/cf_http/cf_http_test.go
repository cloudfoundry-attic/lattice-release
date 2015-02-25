package cf_http_test

import (
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cf_http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfHttp", func() {
	Describe("NewClient", func() {
		Context("when a default timeout has been initialized", func() {
			var timeout time.Duration

			BeforeEach(func() {
				timeout = 1 * time.Second
				cf_http.Initialize(timeout)
			})

			It("returns an http client with the default timeout set", func() {
				Ω(*cf_http.NewClient()).Should(Equal(http.Client{
					Timeout: timeout,
				}))
			})
		})

		Context("when nothing has been initialized", func() {
			It("returns a DefaultClient-equivalent http client", func() {
				Ω(*cf_http.NewClient()).Should(Equal(*http.DefaultClient))
			})
		})
	})
})
