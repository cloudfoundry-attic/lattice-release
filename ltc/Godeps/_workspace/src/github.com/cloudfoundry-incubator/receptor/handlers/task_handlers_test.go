package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskHandler", func() {
	var (
		logger           lager.Logger
		fakeClient       *fake_bbs.FakeClient
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.TaskHandler
		request          *http.Request
	)

	BeforeEach(func() {
		fakeClient = &fake_bbs.FakeClient{}
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewTaskHandler(fakeClient, logger)
	})

	Describe("Create", func() {
		var validCreateRequest receptor.TaskCreateRequest
		var expectedTask *models.Task

		BeforeEach(func() {
			validCreateRequest = receptor.TaskCreateRequest{
				TaskGuid:   "task-guid-1",
				Domain:     "test-domain",
				RootFS:     "docker://docker",
				Action:     models.WrapAction(&models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}}),
				MemoryMB:   24,
				DiskMB:     12,
				CPUWeight:  10,
				LogGuid:    "guid",
				LogSource:  "source-name",
				ResultFile: "result-file",
				Annotation: "some annotation",
				Privileged: true,
			}

			expectedTask = model_helpers.NewValidTask("task-guid-1")
			expectedTask.Domain = "test-domain"
			expectedTask.TaskDefinition.RootFs = "docker://docker"
			expectedTask.TaskDefinition.Action = models.WrapAction(&models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}})
			expectedTask.TaskDefinition.MemoryMb = 24
			expectedTask.TaskDefinition.DiskMb = 12
			expectedTask.TaskDefinition.CpuWeight = 10
			expectedTask.TaskDefinition.LogGuid = "guid"
			expectedTask.TaskDefinition.LogSource = "source-name"
			expectedTask.TaskDefinition.ResultFile = "result-file"
			expectedTask.TaskDefinition.Annotation = "some annotation"
			expectedTask.TaskDefinition.Privileged = true
			expectedTask.TaskDefinition.EgressRules = nil
			expectedTask.TaskDefinition.MetricsGuid = ""
			expectedTask.TaskDefinition.EnvironmentVariables = nil

		})

		Context("when everything succeeds", func() {
			JustBeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Expect(fakeClient.DesireTaskCallCount()).To(Equal(1))
				_, _, def := fakeClient.DesireTaskArgsForCall(0)
				Expect(def).To(Equal(expectedTask.TaskDefinition))
			})

			It("responds with 201 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})

			Context("when omitempty fields", func() {
				Context("are specified", func() {
					BeforeEach(func() {
						validCreateRequest.EnvironmentVariables = []*models.EnvironmentVariable{
							{Name: "var1", Value: "val1"},
							{Name: "var2", Value: "val2"},
						}
						validCreateRequest.EgressRules = []*models.SecurityGroupRule{
							{Protocol: "tcp"},
						}
					})

					It("passes them to the BBS", func() {
						Expect(fakeClient.DesireTaskCallCount()).To(Equal(1))
						_, _, def := fakeClient.DesireTaskArgsForCall(0)
						Expect(def.EnvironmentVariables).To(Equal([]*models.EnvironmentVariable{
							{Name: "var1", Value: "val1"},
							{Name: "var2", Value: "val2"},
						}))
						Expect(def.EgressRules).To(Equal([]*models.SecurityGroupRule{{Protocol: "tcp"}}))
					})
				})

				Context("when are not specified", func() {
					It("passes a nil slice to the BBS", func() {
						Expect(fakeClient.DesireTaskCallCount()).To(Equal(1))
						_, _, def := fakeClient.DesireTaskArgsForCall(0)
						Expect(def.EnvironmentVariables).To(BeNil())
						Expect(def.EgressRules).To(BeNil())
					})
				})
			})

			Context("when completion_callback_url is invalid", func() {
				BeforeEach(func() {
					validCreateRequest.CompletionCallbackURL = "ಠ_ಠ"
				})

				It("errors", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func() {
				fakeClient.DesireTaskReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Expect(fakeClient.DesireTaskCallCount()).To(Equal(1))
				taskGuid, _, _ := fakeClient.DesireTaskArgsForCall(0)
				Expect(taskGuid).To(Equal("task-guid-1"))
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

		Context("when the requested task is invalid", func() {
			var validationError = models.ErrBadRequest

			BeforeEach(func() {
				fakeClient.DesireTaskReturns(validationError)
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidTask,
					Message: validationError.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the requested task exists", func() {
			var desireError = models.ErrResourceExists

			BeforeEach(func() {
				fakeClient.DesireTaskReturns(desireError)
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("responds with 409 CONFLICT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusConflict))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.TaskGuidAlreadyExists,
					Message: "task already exists",
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a TaskCreateRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireTask on the BBS", func() {
				Expect(fakeClient.DesireTaskCallCount()).To(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.TaskCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAll", func() {
		Context("when reading tasks from the BBS fails", func() {
			BeforeEach(func() {
				fakeClient.TasksReturns(nil, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			var domain1Task, domain2Task *models.Task

			BeforeEach(func() {
				domain1Task = model_helpers.NewValidTask("task-guid-1")
				domain1Task.Domain = "domain-1"
				domain2Task = model_helpers.NewValidTask("task-guid-2")
				domain2Task.Domain = "domain-2"

				fakeClient.TasksReturns([]*models.Task{
					domain1Task,
					domain2Task,
				}, nil)

				fakeClient.TasksByDomainReturns([]*models.Task{
					domain1Task,
				}, nil)
			})

			Context("when a domain query param is provided", func() {
				It("gets all tasks", func() {
					request, err := http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))

					var tasks []receptor.TaskResponse
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeClient.TasksByDomainCallCount()).To(Equal(1))
					actualDomain := fakeClient.TasksByDomainArgsForCall(0)
					Expect(actualDomain).To(Equal("domain-1"))

					expectedTasks := []receptor.TaskResponse{
						serialization.TaskToResponse(domain1Task),
					}
					Expect(tasks).To(ConsistOf(expectedTasks))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("gets all tasks", func() {
					handler.GetAll(responseRecorder, newTestRequest(""))
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))

					var tasks []receptor.TaskResponse
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Expect(err).NotTo(HaveOccurred())

					expectedTasks := []receptor.TaskResponse{
						serialization.TaskToResponse(domain1Task),
						serialization.TaskToResponse(domain2Task),
					}
					Expect(tasks).To(ConsistOf(expectedTasks))
				})
			})
		})
	})

	Describe("GetByGuid", func() {
		var taskGuid string
		BeforeEach(func() {
			taskGuid = "the-task-guid"
		})

		JustBeforeEach(func() {
			var err error
			request, err = http.NewRequest("", fmt.Sprintf("http://example.com/?:task_guid=%s", taskGuid), nil)
			Expect(err).NotTo(HaveOccurred())
			handler.GetByGuid(responseRecorder, request)
		})

		Context("when the guid is not provided", func() {
			BeforeEach(func() {
				taskGuid = ""
			})

			It("does not call TaskByGuid", func() {
				Expect(fakeClient.TaskByGuidCallCount()).To(Equal(0))
			})

			It("responds with a 400 Bad Request", func() {
				handler.GetByGuid(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				var taskError receptor.Error

				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Expect(err).NotTo(HaveOccurred())
				Expect(taskError).To(Equal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: "task_guid missing from request",
				}))
			})
		})

		Context("when the task is not found", func() {
			BeforeEach(func() {
				fakeClient.TaskByGuidReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with a 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("responds with a TaskNotFound error in the body", func() {
				var taskError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Expect(err).NotTo(HaveOccurred())

				Expect(taskError).To(Equal(receptor.Error{
					Type:    receptor.TaskNotFound,
					Message: "task with guid 'the-task-guid' not found",
				}))

			})
		})

		Context("when reading the task from the BBS fails", func() {
			BeforeEach(func() {
				fakeClient.TaskByGuidReturns(nil, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the task is successfully found in the BBS", func() {
			var task *models.Task

			BeforeEach(func() {
				task = model_helpers.NewValidTask("the-task-guid")
				task.State = models.Task_Running
				fakeClient.TaskByGuidReturns(task, nil)
			})

			It("retrieves the task by the given guid", func() {
				guid := fakeClient.TaskByGuidArgsForCall(0)
				Expect(guid).To(Equal(task.TaskGuid))
			})

			It("gets the task", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var actualTask receptor.TaskResponse
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualTask)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualTask).To(Equal(serialization.TaskToResponse(task)))
			})
		})
	})

	Describe("Delete", func() {
		var resolvingErr error

		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
			Expect(err).NotTo(HaveOccurred())

			resolvingErr = nil
		})

		JustBeforeEach(func() {
			fakeClient.ResolvingTaskReturns(resolvingErr)
			handler.Delete(responseRecorder, request)
		})

		It("succeeds", func() {
			Expect(fakeClient.ResolvingTaskCallCount()).To(Equal(1))
			Expect(fakeClient.ResolvingTaskArgsForCall(0)).To(Equal("the-task-guid"))
			Expect(fakeClient.DeleteTaskCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteTaskArgsForCall(0)).To(Equal("the-task-guid"))

			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

		Context("when marking the task as resolving fails", func() {
			Context("with invalid transition", func() {
				BeforeEach(func() {
					resolvingErr = models.NewTaskTransitionError(models.Task_Running, models.Task_Pending)
				})

				It("fails with a 409", func() {
					Expect(fakeClient.ResolvingTaskCallCount()).To(Equal(1))
					Expect(fakeClient.ResolvingTaskArgsForCall(0)).To(Equal("the-task-guid"))

					Expect(responseRecorder.Code).To(Equal(http.StatusConflict))
				})
			})

			Context("with resource not found", func() {
				BeforeEach(func() {
					resolvingErr = models.ErrResourceNotFound
				})

				It("fails with a 404", func() {
					Expect(fakeClient.ResolvingTaskCallCount()).To(Equal(1))
					Expect(fakeClient.ResolvingTaskArgsForCall(0)).To(Equal("the-task-guid"))

					Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
				})
			})

			Context("any other error", func() {
				BeforeEach(func() {
					resolvingErr = errors.New("Failed to resolve task")
				})

				It("responds with an error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("does not try to resolve the task", func() {
					handler.Delete(responseRecorder, request)
					Expect(fakeClient.DeleteTaskCallCount()).To(BeZero())
				})
			})
		})

		Context("when task cannot be resolved", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				fakeClient.DeleteTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("Cancel", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			handler.Cancel(responseRecorder, request)
		})

		Context("when cancelling the task is successful", func() {
			BeforeEach(func() {
				fakeClient.CancelTaskReturns(nil)
			})

			It("responds with a 200", func() {
				Expect(fakeClient.CancelTaskCallCount()).To(Equal(1))
				Expect(fakeClient.CancelTaskArgsForCall(0)).To(Equal("the-task-guid"))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when the task cannot be found in the BBS", func() {
			BeforeEach(func() {
				fakeClient.CancelTaskReturns(models.ErrResourceNotFound)
			})

			It("responds with a 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("responds with a TaskNotFound error in the body", func() {
				var taskError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Expect(err).NotTo(HaveOccurred())

				Expect(taskError).To(Equal(receptor.Error{
					Type:    receptor.TaskNotFound,
					Message: "task with guid 'the-task-guid' not found",
				}))

			})
		})

		Context("when cancelling fails", func() {
			BeforeEach(func() {
				fakeClient.CancelTaskReturns(errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})
