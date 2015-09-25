package models_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Requests", func() {
	Describe("UpsertDomainRequest", func() {
		Describe("Validate", func() {
			var request models.UpsertDomainRequest

			BeforeEach(func() {
				request = models.UpsertDomainRequest{
					Domain: "something",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the Domain is blank", func() {
				BeforeEach(func() {
					request.Domain = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"domain"}))
				})
			})
		})
	})
})
