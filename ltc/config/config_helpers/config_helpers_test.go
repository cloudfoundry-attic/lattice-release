package config_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/config_helpers"
)

var _ = Describe("ConfigHelpers", func() {
	Describe("ConfigFileLocation", func() {
		It("returns the config location for the diego home path", func() {
			fileLocation := config_helpers.ConfigFileLocation("/home/chicago")
			Expect(fileLocation).To(Equal("/home/chicago/.lattice/config.json"))
		})
	})
})
