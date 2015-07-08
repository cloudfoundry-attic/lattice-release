package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskHandler", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.TaskHandler
		request          *http.Request
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewTaskHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		validCreateRequest := receptor.TaskCreateRequest{
			TaskGuid:   "task-guid-1",
			Domain:     "test-domain",
			RootFS:     "docker://docker",
			Action:     &models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}},
			MemoryMB:   24,
			DiskMB:     12,
			CPUWeight:  10,
			LogGuid:    "guid",
			LogSource:  "source-name",
			ResultFile: "result-file",
			Annotation: "some annotation",
			Privileged: true,
		}

		expectedTask := models.Task{
			TaskGuid:   "task-guid-1",
			Domain:     "test-domain",
			RootFS:     "docker://docker",
			Action:     &models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}},
			MemoryMB:   24,
			DiskMB:     12,
			CPUWeight:  10,
			LogGuid:    "guid",
			LogSource:  "source-name",
			ResultFile: "result-file",
			Annotation: "some annotation",
			Privileged: true,
		}

		Context("when everything succeeds", func() {
			JustBeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Expect(fakeBBS.DesireTaskCallCount()).To(Equal(1))
				_, task := fakeBBS.DesireTaskArgsForCall(0)
				Expect(task).To(Equal(expectedTask))
			})

			It("responds with 201 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})

			Context("when env vars are specified", func() {
				BeforeEach(func() {
					validCreateRequest.EnvironmentVariables = []receptor.EnvironmentVariable{
						{Name: "var1", Value: "val1"},
						{Name: "var2", Value: "val2"},
					}
				})

				AfterEach(func() {
					validCreateRequest.EnvironmentVariables = []receptor.EnvironmentVariable{}
				})

				It("passes them to the BBS", func() {
					Expect(fakeBBS.DesireTaskCallCount()).To(Equal(1))
					_, task := fakeBBS.DesireTaskArgsForCall(0)
					Expect(task.EnvironmentVariables).To(Equal([]models.EnvironmentVariable{
						{Name: "var1", Value: "val1"},
						{Name: "var2", Value: "val2"},
					}))

				})
			})

			Context("when no env vars are specified", func() {
				It("passes a nil slice to the BBS", func() {
					Expect(fakeBBS.DesireTaskCallCount()).To(Equal(1))
					_, task := fakeBBS.DesireTaskArgsForCall(0)
					Expect(task.EnvironmentVariables).To(BeNil())
				})
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func() {
				fakeBBS.DesireTaskReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Expect(fakeBBS.DesireTaskCallCount()).To(Equal(1))
				_, task := fakeBBS.DesireTaskArgsForCall(0)
				Expect(task.TaskGuid).To(Equal("task-guid-1"))
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
			var validationError = models.ValidationError{}

			BeforeEach(func() {
				fakeBBS.DesireTaskReturns(validationError)
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

		Context("when the request does not contain a TaskCreateRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireTask on the BBS", func() {
				Expect(fakeBBS.DesireTaskCallCount()).To(Equal(0))
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
				fakeBBS.TasksReturns([]models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				handler.GetAll(responseRecorder, newTestRequest(""))
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			var domain1Task, domain2Task models.Task

			BeforeEach(func() {
				domain1Task = models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "domain-1",
					Action: &models.RunAction{
						User: "me",
						Path: "the-path",
					},
					State: models.TaskStatePending,
				}

				domain2Task = models.Task{
					TaskGuid: "task-guid-2",
					Domain:   "domain-2",
					Action: &models.RunAction{
						User: "me",
						Path: "the-path",
					},
					State: models.TaskStatePending,
				}

				fakeBBS.TasksReturns([]models.Task{
					domain1Task,
					domain2Task,
				}, nil)

				fakeBBS.TasksByDomainReturns([]models.Task{
					domain1Task,
				}, nil)
			})

			Context("when a domain query param is provided", func() {
				It("gets all tasks", func() {
					var tasks []receptor.TaskResponse

					request, err := http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Expect(err).NotTo(HaveOccurred())

					_, actualDomain := fakeBBS.TasksByDomainArgsForCall(0)
					Expect(actualDomain).To(Equal("domain-1"))
					expectedTasks := []receptor.TaskResponse{
						{
							TaskGuid: domain1Task.TaskGuid,
							Domain:   domain1Task.Domain,
							Action:   domain1Task.Action,
							State:    receptor.TaskStatePending,
						},
					}
					Expect(tasks).To(Equal(expectedTasks))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("gets all tasks", func() {
					var tasks []receptor.TaskResponse

					handler.GetAll(responseRecorder, newTestRequest(""))
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Expect(err).NotTo(HaveOccurred())
					expectedTasks := []receptor.TaskResponse{
						{
							TaskGuid: domain1Task.TaskGuid,
							Domain:   domain1Task.Domain,
							Action:   domain1Task.Action,
							State:    receptor.TaskStatePending,
						},
						{
							TaskGuid: domain2Task.TaskGuid,
							Domain:   domain2Task.Domain,
							Action:   domain2Task.Action,
							State:    receptor.TaskStatePending,
						},
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
				Expect(fakeBBS.TaskByGuidCallCount()).To(Equal(0))
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
				fakeBBS.TaskByGuidReturns(models.Task{}, bbserrors.ErrStoreResourceNotFound)
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
				fakeBBS.TaskByGuidReturns(models.Task{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the task is successfully found in the BBS", func() {
			var expectedTask receptor.TaskResponse

			BeforeEach(func() {
				task := models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "domain-1",
					Action: &models.RunAction{
						User: "me",
						Path: "the-path",
					},
					State: models.TaskStateRunning,
				}

				fakeBBS.TaskByGuidReturns(task, nil)

				expectedTask = receptor.TaskResponse{
					TaskGuid: task.TaskGuid,
					Domain:   task.Domain,
					Action:   task.Action,
					State:    receptor.TaskStateRunning,
				}
			})

			It("retrieves the task by the given guid", func() {
				_, guid := fakeBBS.TaskByGuidArgsForCall(0)
				Expect(guid).To(Equal("the-task-guid"))
			})

			It("gets the task", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var actualTask receptor.TaskResponse
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualTask)
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedTask).To(Equal(actualTask))
			})
		})
	})

	Describe("Delete", func() {
		Context("when marking the task as resolving fails", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				fakeBBS.ResolvingTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("does not try to resolve the task", func() {
				handler.Delete(responseRecorder, request)
				Expect(fakeBBS.ResolveTaskCallCount()).To(BeZero())
			})
		})

		Context("when task cannot be resolved", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				fakeBBS.ResolveTaskReturns(errors.New("Failed to resolve task"))
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

		Context("when the task cannot be found in the BBS", func() {
			BeforeEach(func() {
				fakeBBS.CancelTaskReturns(bbserrors.ErrStoreResourceNotFound)
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
				fakeBBS.CancelTaskReturns(errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when cancelling the task is successful", func() {
			BeforeEach(func() {
				fakeBBS.CancelTaskReturns(nil)
			})

			It("responds with a 200", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
