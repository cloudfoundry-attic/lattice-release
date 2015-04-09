package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cell API", func() {
	var cellPresence models.CellPresence

	BeforeEach(func() {
		capacity := models.NewCellCapacity(128, 1024, 6)
		cellPresence = models.NewCellPresence("cell-0", "1.2.3.4", "the-zone", capacity)
		value, err := models.ToJSON(cellPresence)

		_, err = consulAdapter.AcquireAndMaintainLock(
			shared.CellSchemaPath(cellPresence.CellID),
			value,
			structs.SessionTTLMin,
			nil)
		Ω(err).ShouldNot(HaveOccurred())

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
				cellPresences, err := bbs.Cells()
				Ω(err).ShouldNot(HaveOccurred())
				return cellPresences
			}).Should(HaveLen(1))

			cellResponses, getErr = client.Cells()
		})

		It("responds without error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			cellPresences, err := bbs.Cells()
			Ω(err).ShouldNot(HaveOccurred())

			expectedResponses := make([]receptor.CellResponse, 0, 1)
			for _, cellPresence := range cellPresences {
				expectedResponses = append(expectedResponses, serialization.CellPresenceToCellResponse(cellPresence))
			}

			Ω(cellResponses).Should(ConsistOf(expectedResponses))
		})

	})

})
