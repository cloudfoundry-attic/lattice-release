package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("AuctioneerPresence", func() {
	var auctioneerPresence models.AuctioneerPresence

	var payload string

	BeforeEach(func() {
		auctioneerPresence = models.NewAuctioneerPresence("some-id", "some-address")

		payload = `{
    "auctioneer_id":      "some-id",
    "auctioneer_address": "some-address"
  }`
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json, err := models.ToJSON(&auctioneerPresence)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(json)).To(MatchJSON(payload))
		})
	})

	Describe("FromJSON", func() {
		It("returns an AuctioneerPresence with correct fields", func() {
			decodedAuctioneerPresence := &models.AuctioneerPresence{}
			err := models.FromJSON([]byte(payload), decodedAuctioneerPresence)
			Expect(err).NotTo(HaveOccurred())

			Expect(decodedAuctioneerPresence).To(Equal(&auctioneerPresence))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				payload = "aliens lol"
				decodedAuctioneerPresence := &models.AuctioneerPresence{}
				err := models.FromJSON([]byte(payload), decodedAuctioneerPresence)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Validate", func() {
		Context("when auctioneer presense is valid", func() {
			BeforeEach(func() {
				auctioneerPresence = models.NewAuctioneerPresence("some-id", "some-address")
			})

			It("returns no error", func() {
				Expect(auctioneerPresence.Validate()).NotTo(HaveOccurred())
			})
		})

		Context("when ID of auctioneer presense is invalid", func() {
			BeforeEach(func() {
				auctioneerPresence = models.NewAuctioneerPresence("", "some-address")
			})

			It("returns no error", func() {
				err := auctioneerPresence.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err).To(ContainElement(models.ErrInvalidField{"auctioneer_id"}))
			})
		})

		Context("when address of auctioneer presense is invalid", func() {
			BeforeEach(func() {
				auctioneerPresence = models.NewAuctioneerPresence("some-id", "")
			})

			It("returns no error", func() {
				err := auctioneerPresence.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err).To(ContainElement(models.ErrInvalidField{"auctioneer_address"}))
			})
		})
	})
})
