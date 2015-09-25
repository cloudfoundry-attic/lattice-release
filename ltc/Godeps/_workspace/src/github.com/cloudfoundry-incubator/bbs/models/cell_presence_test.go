package models_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CellPresence", func() {
	var cellPresence models.CellPresence

	var payload string

	var capacity models.CellCapacity

	BeforeEach(func() {
		capacity = models.NewCellCapacity(128, 1024, 3)
		rootfsProviders := []string{"provider-1"}
		preloadedRootFSes := []string{"provider-2"}
		cellPresence = models.NewCellPresence("some-id", "some-address", "some-zone", capacity, rootfsProviders, preloadedRootFSes)

		payload = `{
    "cell_id":"some-id",
    "rep_address": "some-address",
    "zone": "some-zone",
    "capacity": {
       "memory_mb": 128,
       "disk_mb": 1024,
       "containers": 3
		 },
		 "rootfs_providers": {
			 "provider-1": [],
			 "preloaded": ["provider-2"]
		 }
   }`
	})

	Describe("Validate", func() {
		Context("when cell presence is valid", func() {
			It("does not return an error", func() {
				Expect(cellPresence.Validate()).NotTo(HaveOccurred())
			})
		})

		Context("when cell presence is invalid", func() {
			Context("when cell id is invalid", func() {
				BeforeEach(func() {
					cellPresence.CellID = ""
				})

				It("returns an error", func() {
					err := cellPresence.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("cell_id"))
				})
			})

			Context("when rep address is invalid", func() {
				BeforeEach(func() {
					cellPresence.RepAddress = ""
				})

				It("returns an error", func() {
					err := cellPresence.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("rep_address"))
				})
			})

			Context("when cell capacity is invalid", func() {
				Context("when memory is zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMB = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("memory_mb"))
					})
				})

				Context("when memory is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMB = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("memory_mb"))
					})
				})

				Context("when containers are zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("containers"))
					})
				})

				Context("when containers are negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("containers"))
					})
				})

				Context("when disk is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.DiskMB = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("disk_mb"))
					})
				})
			})
		})
	})
})
