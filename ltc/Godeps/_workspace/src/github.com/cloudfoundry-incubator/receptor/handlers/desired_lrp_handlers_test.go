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
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	fake_legacy_bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Desired LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeLegacyBBS    *fake_legacy_bbs.FakeReceptorBBS
		fakeBBS          *fake_bbs.FakeClient
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DesiredLRPHandler
	)

	BeforeEach(func() {
		fakeLegacyBBS = new(fake_legacy_bbs.FakeReceptorBBS)
		fakeBBS = new(fake_bbs.FakeClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDesiredLRPHandler(fakeBBS, fakeLegacyBBS, logger)
	})

	Describe("Create", func() {
		validCreateLRPRequest := receptor.DesiredLRPCreateRequest{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			RootFS:      "the-rootfs",
			Privileged:  true,
			Instances:   1,
			Action: &oldmodels.RunAction{
				User: "me",
				Path: "the-path",
			},
		}

		expectedDesiredLRP := oldmodels.DesiredLRP{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			RootFS:      "the-rootfs",
			Privileged:  true,
			Instances:   1,
			Action: &oldmodels.RunAction{
				User: "me",
				Path: "the-path",
			},
		}

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Expect(fakeLegacyBBS.DesireLRPCallCount()).To(Equal(1))
				_, desired := fakeLegacyBBS.DesireLRPArgsForCall(0)
				Expect(desired).To(Equal(expectedDesiredLRP))
			})

			It("responds with 201 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeLegacyBBS.DesireLRPReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Expect(fakeLegacyBBS.DesireLRPCallCount()).To(Equal(1))
				_, desired := fakeLegacyBBS.DesireLRPArgsForCall(0)
				Expect(desired).To(Equal(expectedDesiredLRP))
			})

			It("responds with 500 INTERNAL ERROR", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ka-boom",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the desired LRP is invalid", func() {
			var validationError = oldmodels.ValidationError{}

			BeforeEach(func(done Done) {
				fakeLegacyBBS.DesireLRPReturns(validationError)

				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidLRP,
					Message: validationError.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the desired LRP already exists", func() {
			BeforeEach(func(done Done) {
				fakeLegacyBBS.DesireLRPReturns(bbserrors.ErrStoreResourceExists)

				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("responds with 409 CONFLICT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusConflict))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.DesiredLRPAlreadyExists,
					Message: "Desired LRP with guid 'the-process-guid' already exists",
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a DesiredLRPCreateRequest", func() {
			var garbageRequest = []byte(`farewell`)

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireLRP on the BBS", func() {
				Expect(fakeLegacyBBS.DesireLRPCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})
	})

	Describe("Get", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-0"}}
		})

		JustBeforeEach(func() {
			handler.Get(responseRecorder, req)
		})

		Context("when reading tasks from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPByProcessGuidReturns(&models.DesiredLRP{
					ProcessGuid: "process-guid-0",
					Domain:      "domain-1",
					Action: models.WrapAction(&models.RunAction{
						User: "me",
						Path: "the-path",
					}),
				}, nil)
			})

			It("calls DesiredLRPByProcessGuid on the BBS", func() {
				Expect(fakeBBS.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
				actualProcessGuid := fakeBBS.DesiredLRPByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal("process-guid-0"))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a desired lrp response", func() {
				response := receptor.DesiredLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.ProcessGuid).To(Equal("process-guid-0"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPByProcessGuidReturns(&models.DesiredLRP{}, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the BBS reports no lrp found", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPByProcessGuidReturns(&models.DesiredLRP{}, models.ErrResourceNotFound)
			})

			It("responds with 404 Status NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("returns an LRPNotFound error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Expect(err).NotTo(HaveOccurred())

				Expect(responseError).To(Equal(receptor.Error{
					Type:    receptor.DesiredLRPNotFound,
					Message: "Desired LRP with guid 'process-guid-0' not found",
				}))

			})
		})

		Context("when the process guid is not provided", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("does not call DesiredLRPByProcessGuid on the BBS", func() {
				Expect(fakeBBS.DesiredLRPByProcessGuidCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an unknown error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Expect(err).NotTo(HaveOccurred())

				Expect(responseError).To(Equal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				}))

			})
		})
	})

	Describe("Update", func() {
		expectedProcessGuid := "some-guid"
		instances := 15
		annotation := "new-annotation"

		routeMessage := json.RawMessage(`[{"port":8080,"hostnames":["new-route-1","new-route-2"]}]`)
		routes := map[string]*json.RawMessage{
			"cf-router": &routeMessage,
		}
		routingInfo := receptor.RoutingInfo{
			"cf-router": &routeMessage,
		}

		validUpdateRequest := receptor.DesiredLRPUpdateRequest{
			Instances:  &instances,
			Annotation: &annotation,
			Routes:     routingInfo,
		}

		expectedUpdate := oldmodels.DesiredLRPUpdate{
			Instances:  &instances,
			Annotation: &annotation,
			Routes:     routes,
		}

		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest(validUpdateRequest)
			req.Form = url.Values{":process_guid": []string{expectedProcessGuid}}
		})

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Expect(fakeLegacyBBS.UpdateDesiredLRPCallCount()).To(Equal(1))
				_, processGuid, update := fakeLegacyBBS.UpdateDesiredLRPArgsForCall(0)
				Expect(processGuid).To(Equal(expectedProcessGuid))
				Expect(update).To(Equal(expectedUpdate))
			})

			It("responds with 204 NO CONTENT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})
		})

		Context("when the :process_guid is blank", func() {
			BeforeEach(func() {
				req = newTestRequest(validUpdateRequest)
				handler.Update(responseRecorder, req)
			})

			It("does not call UpdateDesiredLRP on the BBS", func() {
				Expect(fakeLegacyBBS.UpdateDesiredLRPCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
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

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeLegacyBBS.UpdateDesiredLRPReturns(errors.New("ka-boom"))
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Expect(fakeLegacyBBS.UpdateDesiredLRPCallCount()).To(Equal(1))
				_, processGuid, update := fakeLegacyBBS.UpdateDesiredLRPArgsForCall(0)
				Expect(processGuid).To(Equal(expectedProcessGuid))
				Expect(update).To(Equal(expectedUpdate))
			})

			It("responds with 500 INTERNAL ERROR", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ka-boom",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS returns a Compare and Swap error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeLegacyBBS.UpdateDesiredLRPReturns(bbserrors.ErrStoreComparisonFailed)
			})

			JustBeforeEach(func() {
				handler.Update(responseRecorder, req)
			})

			It("retries up to one time", func() {
				Eventually(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
				Consistently(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
			})

			Context("when the second attempt succeeds", func() {
				BeforeEach(func() {
					fakeLegacyBBS.UpdateDesiredLRPStub = func(logger lager.Logger, processGuid string, update oldmodels.DesiredLRPUpdate) error {
						if fakeLegacyBBS.UpdateDesiredLRPCallCount() == 1 {
							return bbserrors.ErrStoreComparisonFailed
						} else if fakeLegacyBBS.UpdateDesiredLRPCallCount() == 2 {
							return nil
						} else {
							return errors.New("We shouldn't call this function more than twice")
						}
					}
				})

				It("returns a 204 No Content", func() {
					Eventually(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
					Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
					Consistently(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
				})
			})

			Context("when the second attempt fails", func() {
				It("returns a 500 Internal Server Error", func() {
					Eventually(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
					Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
					Consistently(fakeLegacyBBS.UpdateDesiredLRPCallCount).Should(Equal(2))
				})
			})
		})

		Context("when the BBS indicates the LRP was not found", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeLegacyBBS.UpdateDesiredLRPReturns(bbserrors.ErrStoreResourceNotFound)
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Expect(fakeLegacyBBS.UpdateDesiredLRPCallCount()).To(Equal(1))
				_, processGuid, update := fakeLegacyBBS.UpdateDesiredLRPArgsForCall(0)
				Expect(processGuid).To(Equal(expectedProcessGuid))
				Expect(update).To(Equal(expectedUpdate))
			})

			It("responds with 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.DesiredLRPNotFound,
					Message: "Desired LRP with guid 'some-guid' not found",
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain an DesiredLRPUpdateRequest", func() {
			var garbageRequest = []byte(`farewell`)

			BeforeEach(func(done Done) {
				defer close(done)
				req = newTestRequest(garbageRequest)
				req.Form = url.Values{":process_guid": []string{expectedProcessGuid}}
				handler.Update(responseRecorder, req)
			})

			It("does not call DesireLRP on the BBS", func() {
				Expect(fakeLegacyBBS.UpdateDesiredLRPCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPUpdateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})
	})

	Describe("Delete", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"process-guid-0"}}
		})

		JustBeforeEach(func() {
			handler.Delete(responseRecorder, req)
		})

		Context("when deleting lrp from BBS succeeds", func() {
			BeforeEach(func() {
				fakeLegacyBBS.RemoveDesiredLRPByProcessGuidReturns(nil)
			})

			It("calls the BBS to remove the desired LRP", func() {
				Expect(fakeLegacyBBS.RemoveDesiredLRPByProcessGuidCallCount()).To(Equal(1))
				_, actualProcessGuid := fakeLegacyBBS.RemoveDesiredLRPByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal("process-guid-0"))
			})

			It("responds with 204 NO CONTENT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("returns no body", func() {
				Expect(responseRecorder.Body.Bytes()).To(BeEmpty())
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeLegacyBBS.RemoveDesiredLRPByProcessGuidReturns(errors.New("Something went wrong"))
			})

			It("responds with 500 INTERNAL SERVER ERROR", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var deleteError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &deleteError)
				Expect(err).NotTo(HaveOccurred())

				Expect(deleteError).To(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))

			})
		})

		Context("when the BBS returns no lrp", func() {
			BeforeEach(func() {
				fakeLegacyBBS.RemoveDesiredLRPByProcessGuidReturns(bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with 404 Status NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("returns an LRPNotFound error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Expect(err).NotTo(HaveOccurred())

				Expect(responseError).To(Equal(receptor.Error{
					Type:    receptor.DesiredLRPNotFound,
					Message: "Desired LRP with guid 'process-guid-0' not found",
				}))

			})
		})

		Context("when the process guid is not provided", func() {
			BeforeEach(func() {
				req.Form = url.Values{}
			})

			It("does not call the BBS to remove the desired LRP", func() {
				Expect(fakeLegacyBBS.RemoveDesiredLRPByProcessGuidCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an unknown error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Expect(err).NotTo(HaveOccurred())

				Expect(responseError).To(Equal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				}))

			})
		})
	})

	Describe("GetAll", func() {
		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPsStub = func(filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
					if filter.Domain != "" {
						return []*models.DesiredLRP{
							{
								ProcessGuid: "process-guid-2",
								Domain:      "domain-2",
								Action: models.WrapAction(&models.RunAction{
									User: "me",
									Path: "the-path",
								}),
							},
						}, nil
					}
					return []*models.DesiredLRP{
						{
							ProcessGuid: "process-guid-1",
							Domain:      "domain-1",
							Action: models.WrapAction(&models.RunAction{
								User: "me",
								Path: "the-path",
							}),
						},
						{
							ProcessGuid: "process-guid-2",
							Domain:      "domain-2",
							Action: models.WrapAction(&models.RunAction{
								User: "me",
								Path: "the-path",
							}),
						},
					}, nil
				}
			})

			It("call the BBS to retrieve the desired LRP", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(fakeBBS.DesiredLRPsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			Context("when a domain query param is provided", func() {
				It("returns all desired lrp responses for the domain", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-2", nil)
					Expect(err).NotTo(HaveOccurred())

					handler.GetAll(responseRecorder, request)

					response := []receptor.DesiredLRPResponse{}
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(HaveLen(1))
					Expect(response[0].ProcessGuid).To(Equal("process-guid-2"))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("returns all desired lrp responses", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					response := []receptor.DesiredLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(HaveLen(2))
					Expect(response[0].ProcessGuid).To(Equal("process-guid-1"))
					Expect(response[1].ProcessGuid).To(Equal("process-guid-2"))
				})
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPsReturns([]*models.DesiredLRP{}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(fakeBBS.DesiredLRPsCallCount()).To(Equal(1))
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
				fakeBBS.DesiredLRPsReturns([]*models.DesiredLRP{}, errors.New("Something went wrong"))
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
})
