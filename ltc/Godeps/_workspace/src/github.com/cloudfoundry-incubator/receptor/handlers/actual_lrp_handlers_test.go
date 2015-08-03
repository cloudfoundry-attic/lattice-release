package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	fake_legacy_bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Actual LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeLegacyBBS    *fake_legacy_bbs.FakeReceptorBBS
		fakeBBS          *fake_bbs.FakeClient
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPHandler

		oldActualLRP2     oldmodels.ActualLRP
		oldEvacuatingLRP2 oldmodels.ActualLRP

		actualLRP1     *models.ActualLRP
		actualLRP2     *models.ActualLRP
		evacuatingLRP2 *models.ActualLRP
	)

	BeforeEach(func() {
		fakeLegacyBBS = new(fake_legacy_bbs.FakeReceptorBBS)
		fakeBBS = new(fake_bbs.FakeClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewActualLRPHandler(fakeBBS, fakeLegacyBBS, logger)

		actualLRP1 = models.NewRunningActualLRP(
			models.NewActualLRPKey(
				"process-guid-0",
				1,
				"domain-0",
			),
			models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			),
			models.NewActualLRPNetInfo("1.1.1.1", models.NewPortMapping(80, 5050)),
			1138,
		)

		actualLRP2 = models.NewClaimedActualLRP(
			models.NewActualLRPKey(
				"process-guid-1",
				2,
				"domain-1",
			),
			models.NewActualLRPInstanceKey(
				"instance-guid-1",
				"cell-id-1",
			),
			4444,
		)

		evacuatingLRP2 = models.NewRunningActualLRP(
			models.NewActualLRPKey(
				"process-guid-1",
				2,
				"domain-1",
			),
			models.NewActualLRPInstanceKey(
				"instance-guid-1",
				"cell-id-1",
			),
			models.NewActualLRPNetInfo(""),
			3417,
		)
	})

	// old before each
	BeforeEach(func() {
		oldActualLRP2 = oldmodels.ActualLRP{
			ActualLRPKey: oldmodels.NewActualLRPKey(
				"process-guid-1",
				2,
				"domain-1",
			),
			ActualLRPInstanceKey: oldmodels.NewActualLRPInstanceKey(
				"instance-guid-1",
				"cell-id-1",
			),
			State: oldmodels.ActualLRPStateClaimed,
			Since: 4444,
		}

		oldEvacuatingLRP2 = oldActualLRP2
		oldEvacuatingLRP2.State = oldmodels.ActualLRPStateRunning
		oldEvacuatingLRP2.Since = 3417
	})

	Describe("GetAll", func() {
		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsStub = func(filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
					groups := []*models.ActualLRPGroup{}
					if filter.Domain == "" {
						groups = append(groups, &models.ActualLRPGroup{Instance: actualLRP1})
					}
					groups = append(groups, &models.ActualLRPGroup{Instance: actualLRP2, Evacuating: evacuatingLRP2})
					return groups, nil
				}
			})

			It("calls the BBS to retrieve the actual LRP groups", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(fakeBBS.ActualLRPGroupsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			Context("when a domain query param is provided", func() {
				It("returns a list of desired lrp responses for the domain", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					response := []receptor.ActualLRPResponse{}
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(HaveLen(1))
					Expect(response[0]).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingLRP2, true)))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("returns a list of desired lrp responses", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					response := []receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(HaveLen(2))
					Expect(response[0].ProcessGuid).To(Equal("process-guid-0"))
					Expect(response[1].ProcessGuid).To(Equal("process-guid-1"))
					expectedResponses := []receptor.ActualLRPResponse{
						serialization.ActualLRPProtoToResponse(actualLRP1, false),
						serialization.ActualLRPProtoToResponse(evacuatingLRP2, true),
					}

					Expect(response).To(ConsistOf(expectedResponses))
				})
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsReturns([]*models.ActualLRPGroup{}, nil)
			})

			It("call the BBS to retrieve the actual LRPs", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(fakeBBS.ActualLRPGroupsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsReturns([]*models.ActualLRPGroup{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Expect(err).NotTo(HaveOccurred())

				Expect(receptorError).To(Equal(receptor.Error{
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
				fakeBBS.ActualLRPGroupsByProcessGuidReturns([]*models.ActualLRPGroup{&models.ActualLRPGroup{Instance: actualLRP1}}, nil)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Expect(fakeBBS.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))
				actualProcessGuid := fakeBBS.ActualLRPGroupsByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal("process-guid-0"))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp responses", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(1))
				Expect(response).To(ContainElement(serialization.ActualLRPProtoToResponse(actualLRP1, false)))
			})

			Context("when the index is evacuating", func() {
				BeforeEach(func() {
					req.Form = url.Values{":process_guid": []string{"process-guid-1"}}

					fakeBBS.ActualLRPGroupsByProcessGuidReturns([]*models.ActualLRPGroup{&models.ActualLRPGroup{Instance: actualLRP2, Evacuating: evacuatingLRP2}}, nil)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Expect(fakeBBS.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))
					actualProcessGuid := fakeBBS.ActualLRPGroupsByProcessGuidArgsForCall(0)
					Expect(actualProcessGuid).To(Equal("process-guid-1"))
				})

				It("responds with 200 Status OK", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				})

				It("returns a list of actual lrp responses", func() {
					response := []receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(HaveLen(1))
					Expect(response).To(ContainElement(serialization.ActualLRPProtoToResponse(evacuatingLRP2, true)))
				})
			})
		})

		Context("when reading LRP groups from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsByProcessGuidReturns(nil, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRPs", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupsByProcessGuidReturns([]*models.ActualLRPGroup{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := []receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(0))
			})
		})

		Context("when the request does not contain a process_guid parameter", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
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
					&models.ActualLRPGroup{Instance: actualLRP2},
					nil,
				)
			})

			It("calls the BBS to retrieve the actual LRPs", func() {
				Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
				processGuid, index := fakeBBS.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
				Expect(processGuid).To(Equal("process-guid-1"))
				Expect(index).To(Equal(2))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an actual lrp response", func() {
				response := receptor.ActualLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(actualLRP2, false)))
			})

			Context("when the LRP group contains an evacuating", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						&models.ActualLRPGroup{Instance: actualLRP2, Evacuating: evacuatingLRP2},
						nil,
					)
				})

				It("responds with the reconciled LRP", func() {
					response := receptor.ActualLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingLRP2, true)))
				})
			})
		})

		Context("when reading LRPs from BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(&models.ActualLRPGroup{}, errors.New("Something went wrong"))
			})

			It("responds with a 500 Internal Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS does not return any actual LRP", func() {
			BeforeEach(func() {
				fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(&models.ActualLRPGroup{}, models.ErrResourceNotFound)
			})

			It("responds with 404 Not Found", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when request includes a bad index query parameter", func() {
			BeforeEach(func() {
				req.Form.Set(":index", "not-a-number")
			})

			It("does not call the BBS", func() {
				Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
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
						&models.ActualLRPGroup{Instance: actualLRP2, Evacuating: nil},
						nil,
					)
				})

				It("calls the BBS to retrieve the actual LRPs", func() {
					Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
					processGuid, index := fakeBBS.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
					Expect(processGuid).To(Equal("process-guid-1"))
					Expect(index).To(Equal(0))
				})

				It("calls the BBS to request stop LRP instances", func() {
					Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(1))
					_, oldActualLRPKeys := fakeLegacyBBS.RetireActualLRPsArgsForCall(0)
					Expect(oldActualLRPKeys).To(ConsistOf(oldActualLRP2.ActualLRPKey))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
				})

				Context("when the LRP group contains an evacuating", func() {
					BeforeEach(func() {
						fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
							&models.ActualLRPGroup{Instance: actualLRP2, Evacuating: evacuatingLRP2},
							nil,
						)
					})

					It("calls the BBS to retire teh reconciled instance", func() {
						Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(1))
						_, oldActualLRPKeys := fakeLegacyBBS.RetireActualLRPsArgsForCall(0)
						Expect(oldActualLRPKeys).To(ConsistOf(oldEvacuatingLRP2.ActualLRPKey))
					})
				})
			})

			Context("when the BBS returns no lrps", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						&models.ActualLRPGroup{},
						models.ErrResourceNotFound,
					)
				})

				It("call the BBS to retrieve the desired LRP", func() {
					Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
				})

				It("responds with 404 Status NOT FOUND", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
				})
			})

			Context("when reading LRPs from BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.ActualLRPGroupByProcessGuidAndIndexReturns(
						&models.ActualLRPGroup{},
						errors.New("Something went wrong"))
				})

				It("does not call the BBS to request stopping instances", func() {
					Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(0))
				})

				It("responds with a 500 Internal Error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "Something went wrong",
					})

					Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
				})
			})
		})

		Context("when the index is not specified", func() {
			It("does not call the BBS at all", func() {
				Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(0))
				Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index missing from request",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the index is not a number", func() {
			BeforeEach(func() {
				req.Form.Add(":index", "not-a-number")
			})

			It("does not call the BBS at all", func() {
				Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(0))
				Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "index not a number",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the process guid is not specified", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("does not call the BBS at all", func() {
				Expect(fakeBBS.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(0))
				Expect(fakeLegacyBBS.RetireActualLRPsCallCount()).To(Equal(0))
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})
	})
})
