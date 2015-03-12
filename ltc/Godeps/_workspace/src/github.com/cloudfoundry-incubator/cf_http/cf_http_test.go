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
		var timeout time.Duration

		BeforeEach(func() {
			timeout = 1 * time.Second
		})

		It("returns an http client", func() {
			By("Getting a client before initializaqtion", func() {
				Ω(*cf_http.NewClient()).Should(Equal(*http.DefaultClient))
			})

			cf_http.Initialize(timeout)

			Ω(*cf_http.NewClient()).Should(Equal(http.Client{
				Timeout: timeout,
			}))
		})
	})
})
