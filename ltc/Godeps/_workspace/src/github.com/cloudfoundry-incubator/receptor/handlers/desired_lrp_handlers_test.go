package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Desired LRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DesiredLRPHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDesiredLRPHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		validCreateLRPRequest := receptor.DesiredLRPCreateRequest{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			Stack:       "the-stack",
			RootFSPath:  "the-rootfs-path",
			Privileged:  true,
			Instances:   1,
			Action: &models.RunAction{
				Path: "the-path",
			},
		}

		expectedDesiredLRP := models.DesiredLRP{
			ProcessGuid: "the-process-guid",
			Domain:      "the-domain",
			Stack:       "the-stack",
			RootFSPath:  "the-rootfs-path",
			Privileged:  true,
			Instances:   1,
			Action: &models.RunAction{
				Path: "the-path",
			},
		}

		Context("when everything succeeds", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(1))
				_, desired := fakeBBS.DesireLRPArgsForCall(0)
				Ω(desired).To(Equal(expectedDesiredLRP))
			})

			It("responds with 201 CREATED", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Ω(responseRecorder.Body.String()).Should(Equal(""))
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.DesireLRPReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("calls DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(1))
				_, desired := fakeBBS.DesireLRPArgsForCall(0)
				Ω(desired).To(Equal(expectedDesiredLRP))
			})

			It("responds with 500 INTERNAL ERROR", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ka-boom",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the desired LRP is invalid", func() {
			var validationError = models.ValidationError{}

			BeforeEach(func(done Done) {
				fakeBBS.DesireLRPReturns(validationError)

				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidLRP,
					Message: validationError.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the desired LRP already exists", func() {
			BeforeEach(func(done Done) {
				fakeBBS.DesireLRPReturns(bbserrors.ErrStoreResourceExists)

				defer close(done)
				handler.Create(responseRecorder, newTestRequest(validCreateLRPRequest))
			})

			It("responds with 409 CONFLICT", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusConflict))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.DesiredLRPAlreadyExists,
					Message: "Desired LRP with guid 'the-process-guid' already exists",
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a DesiredLRPCreateRequest", func() {
			var garbageRequest = []byte(`farewell`)

			BeforeEach(func(done Done) {
				defer close(done)
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireLRP on the BBS", func() {
				Ω(fakeBBS.DesireLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
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
				fakeBBS.DesiredLRPByProcessGuidReturns(models.DesiredLRP{
					ProcessGuid: "process-guid-0",
					Domain:      "domain-1",
					Action: &models.RunAction{
						Path: "the-path",
					},
				}, nil)
			})

			It("calls DesiredLRPByProcessGuid on the BBS", func() {
				Ω(fakeBBS.DesiredLRPByProcessGuidCallCount()).Should(Equal(1))
				Ω(fakeBBS.DesiredLRPByProcessGuidArgsForCall(0)).Should(Equal("process-guid-0"))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a desired lrp response", func() {
				response := receptor.DesiredLRPResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.ProcessGuid).Should(Equal("process-guid-0"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPByProcessGuidReturns(models.DesiredLRP{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the BBS reports no lrp found", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPByProcessGuidReturns(models.DesiredLRP{}, bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with 404 Status NOT FOUND", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})

			It("returns an LRPNotFound error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(responseError).Should(Equal(receptor.Error{
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
				Ω(fakeBBS.DesiredLRPByProcessGuidCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("returns an unknown error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(responseError).Should(Equal(receptor.Error{
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

		expectedUpdate := models.DesiredLRPUpdate{
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
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(1))
				_, processGuid, update := fakeBBS.UpdateDesiredLRPArgsForCall(0)
				Ω(processGuid).Should(Equal(expectedProcessGuid))
				Ω(update).Should(Equal(expectedUpdate))
			})

			It("responds with 204 NO CONTENT", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
			})

			It("responds with an empty body", func() {
				Ω(responseRecorder.Body.String()).Should(Equal(""))
			})
		})

		Context("when the :process_guid is blank", func() {
			BeforeEach(func() {
				req = newTestRequest(validUpdateRequest)
				handler.Update(responseRecorder, req)
			})

			It("does not call UpdateDesiredLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
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

		Context("when the BBS responds with an error", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.UpdateDesiredLRPReturns(errors.New("ka-boom"))
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(1))
				_, processGuid, update := fakeBBS.UpdateDesiredLRPArgsForCall(0)
				Ω(processGuid).Should(Equal(expectedProcessGuid))
				Ω(update).Should(Equal(expectedUpdate))
			})

			It("responds with 500 INTERNAL ERROR", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ka-boom",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the BBS indicates the LRP was not found", func() {
			BeforeEach(func(done Done) {
				defer close(done)
				fakeBBS.UpdateDesiredLRPReturns(bbserrors.ErrStoreResourceNotFound)
				handler.Update(responseRecorder, req)
			})

			It("calls UpdateDesiredLRP on the BBS", func() {
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(1))
				_, processGuid, update := fakeBBS.UpdateDesiredLRPArgsForCall(0)
				Ω(processGuid).Should(Equal(expectedProcessGuid))
				Ω(update).Should(Equal(expectedUpdate))
			})

			It("responds with 404 NOT FOUND", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.DesiredLRPNotFound,
					Message: "Desired LRP with guid 'some-guid' not found",
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
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
				Ω(fakeBBS.UpdateDesiredLRPCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.DesiredLRPUpdateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
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
				fakeBBS.RemoveDesiredLRPByProcessGuidReturns(nil)
			})

			It("calls the BBS to remove the desired LRP", func() {
				Ω(fakeBBS.RemoveDesiredLRPByProcessGuidCallCount()).Should(Equal(1))
				_, actualProcessGuid := fakeBBS.RemoveDesiredLRPByProcessGuidArgsForCall(0)
				Ω(actualProcessGuid).Should(Equal("process-guid-0"))
			})

			It("responds with 204 NO CONTENT", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
			})

			It("returns no body", func() {
				Ω(responseRecorder.Body.Bytes()).Should(BeEmpty())
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.RemoveDesiredLRPByProcessGuidReturns(errors.New("Something went wrong"))
			})

			It("responds with 500 INTERNAL SERVER ERROR", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var deleteError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &deleteError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(deleteError).Should(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))
			})
		})

		Context("when the BBS returns no lrp", func() {
			BeforeEach(func() {
				fakeBBS.RemoveDesiredLRPByProcessGuidReturns(bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with 404 Status NOT FOUND", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})

			It("returns an LRPNotFound error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(responseError).Should(Equal(receptor.Error{
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
				Ω(fakeBBS.RemoveDesiredLRPByProcessGuidCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("returns an unknown error", func() {
				var responseError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(responseError).Should(Equal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "process_guid missing from request",
				}))
			})
		})
	})

	Describe("GetAll", func() {
		Context("when reading LRPs from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPsReturns([]models.DesiredLRP{
					{
						ProcessGuid: "process-guid-1",
						Domain:      "domain-1",
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
					{
						ProcessGuid: "process-guid-2",
						Domain:      "domain-2",
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				}, nil)

				fakeBBS.DesiredLRPsByDomainReturns([]models.DesiredLRP{
					{
						ProcessGuid: "process-guid-2",
						Domain:      "domain-2",
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.DesiredLRPsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			Context("when a domain query param is provided", func() {
				It("returns all desired lrp responses for the domain", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-2", nil)
					Ω(err).ShouldNot(HaveOccurred())

					handler.GetAll(responseRecorder, request)

					response := []receptor.DesiredLRPResponse{}
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(response).Should(HaveLen(1))
					Ω(response[0].ProcessGuid).Should(Equal("process-guid-2"))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("returns all desired lrp responses", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					response := []receptor.DesiredLRPResponse{}
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(response).Should(HaveLen(2))
					Ω(response[0].ProcessGuid).Should(Equal("process-guid-1"))
					Ω(response[1].ProcessGuid).Should(Equal("process-guid-2"))
				})
			})
		})

		Context("when the BBS returns no lrps", func() {
			BeforeEach(func() {
				fakeBBS.DesiredLRPsReturns([]models.DesiredLRP{}, nil)
			})

			It("call the BBS to retrieve the desired LRP", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Ω(fakeBBS.DesiredLRPsCallCount()).Should(Equal(1))
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
				fakeBBS.DesiredLRPsReturns([]models.DesiredLRP{}, errors.New("Something went wrong"))
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
})
