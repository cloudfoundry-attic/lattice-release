package setup_cli_helpers_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/setup_cli/setup_cli_helpers"
)

var _ = Describe("setup_cli_helpers", func() {
	Describe("Timeout", func() {
		It("returns the timeout in seconds", func() {
			Expect(setup_cli_helpers.Timeout("25")).To(Equal(25 * time.Second))
		})

		It("returns one minute for an empty string", func() {
			Expect(setup_cli_helpers.Timeout("")).To(Equal(time.Minute))
		})

		It("returns one minute for an invalid string", func() {
			Expect(setup_cli_helpers.Timeout("CANNOT PARSE")).To(Equal(time.Minute))
		})
	})

	Describe("LoggregatorUrl", func() {
		It("returns loggregator url with the websocket scheme added", func() {
			loggregatorUrl := setup_cli_helpers.LoggregatorUrl("doppler.diego.io")
			Expect(loggregatorUrl).To(Equal("ws://doppler.diego.io"))
		})
	})
})
