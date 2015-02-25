package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("CellPresence", func() {
	var cellPresence models.CellPresence

	var payload string

	var capacity models.CellCapacity

	BeforeEach(func() {
		capacity = models.NewCellCapacity(128, 1024, 3)
		cellPresence = models.NewCellPresence("some-id", "some-stack", "some-address", "some-zone", capacity)

		payload = `{
    "cell_id":"some-id",
    "stack": "some-stack",
    "rep_address": "some-address",
    "zone": "some-zone",
    "capacity": {
       "memory_mb": 128,
       "disk_mb": 1024,
       "containers": 3
    }
  }`
	})

	Describe("Validate", func() {
		Context("when cell presence is valid", func() {
			It("does not return an error", func() {
				Ω(cellPresence.Validate()).ShouldNot(HaveOccurred())
			})
		})
		Context("when cell presence is invalid", func() {
			Context("when cell id is invalid", func() {
				BeforeEach(func() {
					cellPresence.CellID = ""
				})
				It("returns an error", func() {
					err := cellPresence.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("cell_id"))
				})
			})
			Context("when stack is invalid", func() {
				BeforeEach(func() {
					cellPresence.Stack = ""
				})
				It("returns an error", func() {
					err := cellPresence.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("stack"))
				})
			})
			Context("when rep address is invalid", func() {
				BeforeEach(func() {
					cellPresence.RepAddress = ""
				})
				It("returns an error", func() {
					err := cellPresence.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("rep_address"))
				})
			})

			Context("when cell capacity is invalid", func() {
				Context("when memory is zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMB = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Ω(err).Should(HaveOccurred())
						Ω(err.Error()).Should(ContainSubstring("memory_mb"))
					})
				})

				Context("when memory is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMB = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Ω(err).Should(HaveOccurred())
						Ω(err.Error()).Should(ContainSubstring("memory_mb"))
					})
				})

				Context("when containers are zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Ω(err).Should(HaveOccurred())
						Ω(err.Error()).Should(ContainSubstring("containers"))
					})
				})

				Context("when containers are negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Ω(err).Should(HaveOccurred())
						Ω(err.Error()).Should(ContainSubstring("containers"))
					})
				})

				Context("when disk is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.DiskMB = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Ω(err).Should(HaveOccurred())
						Ω(err.Error()).Should(ContainSubstring("disk_mb"))
					})
				})
			})
		})
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json, err := models.ToJSON(&cellPresence)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(json)).Should(MatchJSON(payload))
		})
	})

	Describe("FromJSON", func() {
		It("returns a CellPresence with correct fields", func() {
			decodedCellPresence := &models.CellPresence{}
			err := models.FromJSON([]byte(payload), decodedCellPresence)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedCellPresence).Should(Equal(&cellPresence))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				payload = "aliens lol"
				decodedCellPresence := &models.CellPresence{}
				err := models.FromJSON([]byte(payload), decodedCellPresence)

				Ω(err).Should(HaveOccurred())
			})
		})
	})
})
