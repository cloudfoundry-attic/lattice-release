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
			cellPresence = models.NewCellPresence("cell-id-0", "1.2.3.4", "the-zone", capacity, []string{"provider-1", "provider-2"}, []string{"stack-1"})
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.CellResponse{
				CellID: "cell-id-0",
				Zone:   "the-zone",
				Capacity: receptor.CellCapacity{
					MemoryMB:   128,
					DiskMB:     1024,
					Containers: 6,
				},
				RootFSProviders: map[string][]string{
					"provider-1": []string{},
					"provider-2": []string{},
					"preloaded":  []string{"stack-1"},
				},
			}

			actualResponse := serialization.CellPresenceToCellResponse(cellPresence)
			Expect(actualResponse).To(Equal(expectedResponse))
		})
	})
})
