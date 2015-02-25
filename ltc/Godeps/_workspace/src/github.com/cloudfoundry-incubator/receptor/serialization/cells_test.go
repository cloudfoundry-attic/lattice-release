package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CellPresence Serialization", func() {
	Describe("CellPresenceToCellResponse", func() {
		var cellPresence models.CellPresence

		BeforeEach(func() {
			capacity := models.NewCellCapacity(128, 1024, 6)
			cellPresence = models.NewCellPresence("cell-id-0", "stack-0", "1.2.3.4", "the-zone", capacity)
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.CellResponse{
				CellID: "cell-id-0",
				Stack:  "stack-0",
				Zone:   "the-zone",
				Capacity: receptor.CellCapacity{
					MemoryMB:   128,
					DiskMB:     1024,
					Containers: 6,
				},
			}

			actualResponse := serialization.CellPresenceToCellResponse(cellPresence)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})
	})
})
