package logs_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/diego-edge-cli/logs/logs_helpers"
)

var _ = Describe("logs_helpers", func() {
	Describe("LoggregatorUrl", func() {
		It("returns loggregator url with the websocket scheme added", func() {
			loggregatorUrl := logs_helpers.LoggregatorUrl("doppler.diego.io")
			Expect(loggregatorUrl).To(Equal("ws://doppler.diego.io"))
		})
	})
})
