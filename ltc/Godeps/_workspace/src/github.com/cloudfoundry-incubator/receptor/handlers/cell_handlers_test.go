package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Cell Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.CellHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewCellHandler(fakeBBS, logger)
	})

	Describe("GetAll", func() {
		var cellPresences []models.CellPresence

		BeforeEach(func() {
			capacity := models.NewCellCapacity(128, 1024, 6)
			cellPresences = []models.CellPresence{
				models.NewCellPresence("cell-id-0", "stack-0", "1.2.3.4", "the-zone", capacity),
				models.NewCellPresence("cell-id-1", "stack-1", "4.5.6.7", "the-zone", capacity),
			}
		})

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading Cells from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.CellsReturns(cellPresences, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.CellsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of cell presence responses", func() {
				response := []receptor.CellResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(2))
				for _, cellPresence := range cellPresences {
					Ω(response).Should(ContainElement(serialization.CellPresenceToCellResponse(cellPresence)))
				}
			})
		})

		Context("when the BBS returns no cells", func() {
			BeforeEach(func() {
				fakeBBS.CellsReturns([]models.CellPresence{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.CellsReturns([]models.CellPresence{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(receptorError).Should(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))
			})
		})
	})
})
