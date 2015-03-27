package models_test

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rootfs", func() {
	Describe("PreloadedRootFS", func() {
		It("generates the correct preloaded rootfs URL for the stack", func() {
			Ω(models.PreloadedRootFS("bluth-cid64")).Should(Equal("preloaded:bluth-cid64"))
		})
	})
})
