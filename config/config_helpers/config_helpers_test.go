package config_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/diego-edge-cli/config/config_helpers"
)

var _ = Describe("config_helpers", func() {
	Describe("configFileLocation", func() {

		It("returns the config location for the diego home path", func() {
			fileLocation := config_helpers.ConfigFileLocation("/home/diego")
			Expect(fileLocation).To(Equal("/home/diego/.diego/config.json"))
		})
	})
})
