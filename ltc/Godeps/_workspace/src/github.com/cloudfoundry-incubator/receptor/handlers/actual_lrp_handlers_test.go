package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Actual LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPHandler

		actualLRP1     models.ActualLRP
		actualLRP2     models.ActualLRP
		evacuatingLRP2 models.ActualLRP
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewActualLRPHandler(fakeBBS, logger)

		actualLRP1 = models.ActualLRP{
			ActualLRPKey: models.NewActualLRPKey(
				"process-guid-0",
				1,
				"domain-0",
			),
			ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			),
			State: models.ActualLRPStateRunning,
			Since: 1138,
		}

		actualLRP2 = models.ActualLRP{
			ActualLRPKey: models.NewActualLRPKey(
				"process-guid-1",
				2,
				"domain-1",
			),
			ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
				"instance-guid-1",
				"cell-id-1",
			),
			State: models.ActualLRPStateClaimed,
			Since: 4444,
		}

		evacuatingLRP2 = actualLRP2
		evacuatingLRP2.State = models.ActualLRPStateRunning
		evacuatingLRP2.Since = 3417
	})

	Describe("GetAll", func() {
		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsReturns([]models.ActualLRPGroup{
					{Instance: &actualLRP1},
					{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
				}, nil)

				fakeBBS.ActualLRPGroupsByDomainReturns([]models.ActualLRPGroup{
					{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
				}, nil)
			})

			It("calls the BBS to retrieve the actual LRP groups", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.ActualLRPGroupsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			Context("when a domain query param is provided", func() {
				It("returns a list of desired lrp responses for the domain", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Ω(err).ShouldNot(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					response := []receptor.ActualLRPResponse{}
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(1))
					Ω(response[0]).Should(Equal(serialization.ActualLRPToResponse(evacuatingLRP2, true)))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("returns a list of desired lrp responses", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					response := []receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(2))
					Ω(response[0].ProcessGuid).Should(Equal("process-guid-0"))
					Ω(response[1].ProcessGuid).Should(Equal("process-guid-1"))
					expectedResponses := []receptor.ActualLRPResponse{
						serialization.ActualLRPToResponse(actualLRP1, false),
						serialization.ActualLRPToResponse(evacuatingLRP2, true),
					}

					Ω(response).Should(ConsistOf(expectedResponses))
				})
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsReturns([]models.ActualLRPGroup{}, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.ActualLRPGroupsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsReturns([]models.ActualLRPGroup{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
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

	Describe("GetAllByProcessGuid", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-0"}}
		})

		JustBeforeEach(func() {
			handler.GetAllByProcessGuid(responseRecorder, req)
		})

		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsByProcessGuidReturns(models.ActualLRPGroupsByIndex{
					1: {Instance: &actualLRP1, Evacuating: nil},
				}, nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPGroupsByProcessGuidCallCount()).Should(Equal(1))
				Ω(fakeBBS.ActualLRPGroupsByProcessGuidArgsForCall(0)).Should(Equal("process-guid-0"))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp responses", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(1))
				Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(actualLRP1, false)))
			})

			Context("when the index is evacuating", func() {
				BeforeEach(func() {
					req.Form = url.Values{":process_guid": []string{"process-guid-1"}}

					fakeBBS.ActualLRPGroupsByProcessGuidReturns(
						models.ActualLRPGroupsByIndex{2: {Instance: &actualLRP2, Evacuating: &evacuatingLRP2}},
						nil,
					)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Ω(fakeBBS.ActualLRPGroupsByProcessGuidCallCount()).Should(Equal(1))
					Ω(fakeBBS.ActualLRPGroupsByProcessGuidArgsForCall(0)).Should(Equal("process-guid-1"))
				})

				It("responds with 200 Status OK", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
				})

				It("returns a list of actual lrp responses", func() {
					response := []receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(1))
					Ω(response).Should(ContainElement(serialization.ActualLRPToResponse(evacuatingLRP2, true)))
				})
			})
		})

		Context("when reading LRP groups from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsByProcessGuidReturns(nil, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRPs", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsByProcessGuidReturns(models.ActualLRPGroupsByIndex{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(0))
			})
		})

		Context("when the request does not contain a process_guid parameter", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

	})

	Describe("GetByProcessGuidAndIndex", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{
				":process_guid": []string{"process-guid-1"},
				":index":        []string{"2"},
			}
		})

		JustBeforeEach(func() {
			handler.GetByProcessGuidAndIndex(responseRecorder, req)
		})

		Context("when getting the LRP group from the BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
					models.ActualLRPGroup{Instance: &actualLRP2},
					nil,
				)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(1))
				processGuid, index := fakeBBS.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
				Ω(processGuid).Should(Equal("process-guid-1"))
				Ω(index).Should(Equal(2))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an actual lrp response", func() {
				response := receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(Equal(serialization.ActualLRPToResponse(actualLRP2, false)))
			})

			Context("when the LRP group contains an evacuating", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						models.ActualLRPGroup{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
						nil,
					)
				})

				It("responds with the reconciled LRP", func() {
					response := receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(Equal(serialization.ActualLRPToResponse(evacuatingLRP2, true)))
				})
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(models.ActualLRPGroup{}, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRP", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(models.ActualLRPGroup{}, bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with 404 Not Found", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})
		})

		Context("when request includes a bad index query parameter", func() {
			BeforeEach(func() {
				req.Form.Set(":index", "not-a-number")
			})

			It("does not call the BBS", func() {
				Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("KillByProcessGuidAndIndex", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-1"}}
		})

		JustBeforeEach(func() {
			handler.KillByProcessGuidAndIndex(responseRecorder, req)
		})

		Context("when request includes a valid index query parameter", func() {
			BeforeEach(func() {
				req.Form.Add(":index", "0")
			})

			Context("when getting the LRP group from BBS succeeds", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						models.ActualLRPGroup{Instance: &actualLRP2, Evacuating: nil},
						nil,
					)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(1))
					processGuid, index := fakeBBS.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
					Ω(processGuid).Should(Equal("process-guid-1"))
					Ω(index).Should(Equal(0))
				})

				It("calls the BBS to request stop LRP instances", func() {
					Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(1))
					_, actualLRPs := fakeBBS.RetireActualLRPsArgsForCall(0)
					Ω(actualLRPs).Should(ConsistOf(actualLRP2))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
				})

				Context("when the LRP group contains an evacuating", func() {
					BeforeEach(func() {
						fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
							models.ActualLRPGroup{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
							nil,
						)
					})

					It("calls the BBS to retire teh reconciled instance", func() {
						Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(1))
						_, actualLRPs := fakeBBS.RetireActualLRPsArgsForCall(0)
						Ω(actualLRPs).Should(ConsistOf(evacuatingLRP2))
					})
				})
			})

			Context("when the BBS returns no lrps", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						models.ActualLRPGroup{},
						bbserrors.ErrStoreResourceNotFound,
					)
				})

				It("call the BBS to retrieve the desired LRP", func() {
					Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(1))
				})

				It("responds with 404 Status NOT FOUND", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
				})
			})

			Context("when reading LRPs from BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						models.ActualLRPGroup{},
						errors.New("Something went wrong"))
				})

				It("does not call the BBS to request stopping instances", func() {
					Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(0))
				})

				It("responds with a 500 Internal Error", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "Something went wrong",
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})
		})

		Context("when the index is not specified", func() {
			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the index is not a number", func() {
			BeforeEach(func() {
				req.Form.Add(":index", "not-a-number")
			})

			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the process guid is not specified", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("does not call the BBS at all", func() {
				Ω(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).Should(Equal(0))
				Ω(fakeBBS.RetireActualLRPsCallCount()).Should(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})
})
