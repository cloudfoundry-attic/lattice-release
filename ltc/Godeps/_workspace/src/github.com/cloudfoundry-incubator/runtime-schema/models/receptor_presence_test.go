package models_test

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReceptorPresence", func() {
	var receptorPresence models.ReceptorPresence

	BeforeEach(func() {
		receptorPresence = models.NewReceptorPresence("receptor-id", "receptor-url")
	})

	Describe("Validate", func() {
		Context("when receptor presence is valid", func() {
			It("does not return an error", func() {
				Ω(receptorPresence.Validate()).ShouldNot(HaveOccurred())
			})
		})

		Context("when receptor presence is invalid", func() {
			Context("when receptor id is invalid", func() {
				BeforeEach(func() {
					receptorPresence.ReceptorID = ""
				})
				It("returns an error", func() {
					err := receptorPresence.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("id"))
				})
			})

			Context("when stack is invalid", func() {
				BeforeEach(func() {
					receptorPresence.ReceptorURL = ""
				})
				It("returns an error", func() {
					err := receptorPresence.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("address"))
				})
			})
		})
	})

})
