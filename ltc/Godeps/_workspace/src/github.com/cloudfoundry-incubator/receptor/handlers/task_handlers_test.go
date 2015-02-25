package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/cloudfoundry-incubator/receptor/handlers"
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
		handler          *TaskHandler
		request          *http.Request
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = NewTaskHandler(fakeBBS, logger)
	})

	Describe("Create", func() {
		validCreateRequest := receptor.TaskCreateRequest{
			TaskGuid:   "task-guid-1",
			Domain:     "test-domain",
			RootFSPath: "docker://docker",
			Stack:      "some-stack",
			Action:     &models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}},
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
			RootFSPath: "docker://docker",
			Stack:      "some-stack",
			Action:     &models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}},
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
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
				_, task := fakeBBS.DesireTaskArgsForCall(0)
				Ω(task).Should(Equal(expectedTask))
			})

			It("responds with 201 CREATED", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Ω(responseRecorder.Body.String()).Should(Equal(""))
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
					Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
					_, task := fakeBBS.DesireTaskArgsForCall(0)
					Ω(task.EnvironmentVariables).Should(Equal([]models.EnvironmentVariable{
						{Name: "var1", Value: "val1"},
						{Name: "var2", Value: "val2"},
					}))
				})
			})

			Context("when no env vars are specified", func() {
				It("passes a nil slice to the BBS", func() {
					Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
					_, task := fakeBBS.DesireTaskArgsForCall(0)
					Ω(task.EnvironmentVariables).Should(BeNil())
				})
			})
		})

		Context("when the BBS responds with an error", func() {
			BeforeEach(func() {
				fakeBBS.DesireTaskReturns(errors.New("ka-boom"))
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("calls DesireTask on the BBS with the correct task", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(1))
				_, task := fakeBBS.DesireTaskArgsForCall(0)
				Ω(task.TaskGuid).Should(Equal("task-guid-1"))
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

		Context("when the requested task is invalid", func() {
			var validationError = models.ValidationError{}

			BeforeEach(func() {
				fakeBBS.DesireTaskReturns(validationError)
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidTask,
					Message: validationError.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a TaskCreateRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("does not call DesireTask on the BBS", func() {
				Ω(fakeBBS.DesireTaskCallCount()).Should(Equal(0))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &receptor.TaskCreateRequest{})
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidJSON,
					Message: err.Error(),
				})
				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
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
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when reading tasks from BBS succeeds", func() {
			var domain1Task, domain2Task models.Task

			BeforeEach(func() {
				domain1Task = models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "domain-1",
					Action: &models.RunAction{
						Path: "the-path",
					},
					State: models.TaskStatePending,
				}

				domain2Task = models.Task{
					TaskGuid: "task-guid-2",
					Domain:   "domain-2",
					Action: &models.RunAction{
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
					Ω(err).ShouldNot(HaveOccurred())

					handler.GetAll(responseRecorder, request)
					Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
					err = json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Ω(err).ShouldNot(HaveOccurred())

					_, actualDomain := fakeBBS.TasksByDomainArgsForCall(0)
					Ω(actualDomain).Should(Equal("domain-1"))
					expectedTasks := []receptor.TaskResponse{
						{
							TaskGuid: domain1Task.TaskGuid,
							Domain:   domain1Task.Domain,
							Action:   domain1Task.Action,
							State:    receptor.TaskStatePending,
						},
					}
					Ω(tasks).Should(Equal(expectedTasks))
				})
			})

			Context("when a domain query param is not provided", func() {
				It("gets all tasks", func() {
					var tasks []receptor.TaskResponse

					handler.GetAll(responseRecorder, newTestRequest(""))
					Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
					err := json.Unmarshal(responseRecorder.Body.Bytes(), &tasks)
					Ω(err).ShouldNot(HaveOccurred())
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
					Ω(tasks).Should(ConsistOf(expectedTasks))
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
			Ω(err).ShouldNot(HaveOccurred())
			handler.GetByGuid(responseRecorder, request)
		})

		Context("when the guid is not provided", func() {
			BeforeEach(func() {
				taskGuid = ""
			})

			It("does not call TaskByGuid", func() {
				Ω(fakeBBS.TaskByGuidCallCount()).Should(Equal(0))
			})

			It("responds with a 400 Bad Request", func() {
				handler.GetByGuid(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				var taskError receptor.Error

				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(taskError).Should(Equal(receptor.Error{
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
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})

			It("responds with a TaskNotFound error in the body", func() {
				var taskError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(taskError).Should(Equal(receptor.Error{
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
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the task is successfully found in the BBS", func() {
			var expectedTask receptor.TaskResponse

			BeforeEach(func() {
				task := models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "domain-1",
					Action: &models.RunAction{
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
				Ω(fakeBBS.TaskByGuidArgsForCall(0)).Should(Equal("the-task-guid"))
			})

			It("gets the task", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))

				var actualTask receptor.TaskResponse
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualTask)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(expectedTask).Should(Equal(actualTask))
			})
		})
	})

	Describe("Delete", func() {
		Context("when marking the task as resolving fails", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Ω(err).ShouldNot(HaveOccurred())
				fakeBBS.ResolvingTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("does not try to resolve the task", func() {
				handler.Delete(responseRecorder, request)
				Ω(fakeBBS.ResolveTaskCallCount()).Should(BeZero())
			})
		})

		Context("when task cannot be resolved", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
				Ω(err).ShouldNot(HaveOccurred())
				fakeBBS.ResolveTaskReturns(errors.New("Failed to resolve task"))
			})

			It("responds with an error", func() {
				handler.Delete(responseRecorder, request)
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("Cancel", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "http://example.com?:task_guid=the-task-guid", nil)
			Ω(err).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			handler.Cancel(responseRecorder, request)
		})

		Context("when the task cannot be found in the BBS", func() {
			BeforeEach(func() {
				fakeBBS.CancelTaskReturns(bbserrors.ErrStoreResourceNotFound)
			})

			It("responds with a 404 NOT FOUND", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusNotFound))
			})

			It("responds with a TaskNotFound error in the body", func() {
				var taskError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &taskError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(taskError).Should(Equal(receptor.Error{
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
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when cancelling the task is successful", func() {
			BeforeEach(func() {
				fakeBBS.CancelTaskReturns(nil)
			})

			It("responds with a 200", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})
		})
	})
})
