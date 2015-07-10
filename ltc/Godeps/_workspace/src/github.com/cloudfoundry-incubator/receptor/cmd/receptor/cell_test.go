package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cell API", func() {
	var cellPresence models.CellPresence

	BeforeEach(func() {
		capacity := models.NewCellCapacity(128, 1024, 6)
		cellPresence = models.NewCellPresence("cell-0", "1.2.3.4", "the-zone", capacity, []string{}, []string{})
		value, err := models.ToJSON(cellPresence)

		_, err = consulSession.SetPresence(shared.CellSchemaPath(cellPresence.CellID), value)
		Expect(err).NotTo(HaveOccurred())

		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("GET /v1/cells", func() {
		var cellResponses []receptor.CellResponse
		var getErr error

		BeforeEach(func() {
			Eventually(func() []models.CellPresence {
				cellPresences, err := legacyBBS.Cells()
				Expect(err).NotTo(HaveOccurred())
				return cellPresences
			}).Should(HaveLen(1))

			cellResponses, getErr = client.Cells()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			cellPresences, err := legacyBBS.Cells()
			Expect(err).NotTo(HaveOccurred())

			expectedResponses := make([]receptor.CellResponse, 0, 1)
			for _, cellPresence := range cellPresences {
				expectedResponses = append(expectedResponses, serialization.CellPresenceToCellResponse(cellPresence))
			}

			Expect(cellResponses).To(ConsistOf(expectedResponses))
		})
	})
})
