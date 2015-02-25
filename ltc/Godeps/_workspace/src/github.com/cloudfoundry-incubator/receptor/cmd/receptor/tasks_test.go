package main_test

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Task API", func() {
	var (
		logger lager.Logger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("Headers", func() {
		It("includes the Content-Length and Content-Type headers", func() {
			httpClient := new(http.Client)
			res, err := httpClient.Get("http://" + receptorAddress + "/tasks")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res.Header.Get("Content-Length")).Should(MatchRegexp(`\d+`))
			Ω(res.Header.Get("Content-Type")).Should(Equal("application/json"))
		})
	})

	Describe("POST /v1/tasks", func() {
		var (
			taskToCreate receptor.TaskCreateRequest
			err          error
			testServer   *ghttp.Server
		)

		BeforeEach(func() {
			testServer = ghttp.NewServer()

			taskToCreate = receptor.TaskCreateRequest{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "some-stack",
				CompletionCallbackURL: testServer.URL() + "/the/callback/path",
				Action:                &models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}},
			}

			err = client.CreateTask(taskToCreate)
		})

		It("responds without an error", func() {
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("desires the task in the BBS", func() {
			Eventually(func() ([]models.Task, error) {
				return bbs.PendingTasks(logger)
			}).Should(HaveLen(1))
		})

		Context("when trying to create a task with a GUID that already exists", func() {
			BeforeEach(func() {
				err = client.CreateTask(taskToCreate)
			})

			It("returns an error indicating that the key already exists", func() {
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.TaskGuidAlreadyExists))
			})
		})

		Describe("when the task completes", func() {
			BeforeEach(func() {
				_, err = bbs.StartTask(logger, "task-guid-1", "the-cell-id")
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("sends a POST request to the specified callback URL", func() {
				testServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/the/callback/path"),
					func(res http.ResponseWriter, req *http.Request) {
						var taskResponse receptor.TaskResponse
						err := json.NewDecoder(req.Body).Decode(&taskResponse)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(taskResponse.TaskGuid).Should(Equal("task-guid-1"))
						Ω(taskResponse.Domain).Should(Equal("test-domain"))
						Ω(taskResponse.Stack).Should(Equal("some-stack"))
						Ω(taskResponse.State).Should(Equal(receptor.TaskStateCompleted))
						Ω(taskResponse.Result).Should(Equal("the-result"))
						Ω(taskResponse.Failed).Should(Equal(true))
						Ω(taskResponse.FailureReason).Should(Equal("the-failure-reason"))
						Ω(taskResponse.Action).Should(Equal(&models.RunAction{Path: "/bin/bash", Args: []string{"echo", "hi"}}))
					},
				))

				Ω(testServer.ReceivedRequests()).Should(HaveLen(0))

				err = bbs.CompleteTask(logger, "task-guid-1", "the-cell-id", true, "the-failure-reason", "the-result")
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(testServer.ReceivedRequests).Should(HaveLen(1))
			})
		})
	})

	Describe("GET /v1/tasks", func() {
		Context("when there are no tasks", func() {
			It("returns an empty array", func() {
				tasks, err := client.Tasks()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(tasks).Should(BeEmpty())
			})
		})

		Context("when there are tasks", func() {
			BeforeEach(func() {
				err := bbs.DesireTask(logger, models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "test-domain",
					Stack:    "some-stack",
					Action:   &models.RunAction{Path: "/bin/true"},
				})
				Ω(err).ShouldNot(HaveOccurred())

				err = bbs.DesireTask(logger, models.Task{
					TaskGuid: "task-guid-2",
					Domain:   "test-domain",
					Stack:    "some-stack",
					Action:   &models.RunAction{Path: "/bin/true"},
				})
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("returns an array of all the tasks", func() {
				tasks, err := client.Tasks()
				Ω(err).ShouldNot(HaveOccurred())

				taskGuids := []string{}
				for _, task := range tasks {
					taskGuids = append(taskGuids, task.TaskGuid)
				}
				Ω(taskGuids).Should(ConsistOf([]string{"task-guid-1", "task-guid-2"}))
			})

		})
	})

	Describe("GET /v1/domains/:domain/tasks", func() {
		BeforeEach(func() {
			err := bbs.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "stack-1",
				Action:   &models.RunAction{Path: "/bin/true"},
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = bbs.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-2",
				Domain:   "other-domain",
				Stack:    "stack-2",
				Action:   &models.RunAction{Path: "/bin/true"},
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = bbs.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-3",
				Domain:   "test-domain",
				Stack:    "stack-3",
				Action:   &models.RunAction{Path: "/bin/true"},
			})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns an array of all the tasks for the domain", func() {
			tasks, err := client.TasksByDomain("test-domain")
			Ω(err).ShouldNot(HaveOccurred())

			taskGuids := []string{}
			for _, task := range tasks {
				taskGuids = append(taskGuids, task.TaskGuid)
			}
			Ω(taskGuids).Should(ConsistOf([]string{"task-guid-1", "task-guid-3"}))
		})
	})

	Describe("GET /v1/tasks/:task_guid", func() {
		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "stack-1",
				Action:   &models.RunAction{Path: "/bin/true"},
			}
			err := bbs.DesireTask(logger, task)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns the task", func() {
			task, err := client.GetTask("task-guid-1")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(task.TaskGuid).Should(Equal("task-guid-1"))
			Ω(task.Domain).Should(Equal("test-domain"))
		})

		It("includes all of the task's publicly-visible fields", func() {
			_, err := bbs.StartTask(logger, "task-guid-1", "the-cell-id")
			Ω(err).ShouldNot(HaveOccurred())
			err = bbs.CompleteTask(logger, "task-guid-1", "the-cell-id", true, "the-failure-reason", "the-task-result")
			Ω(err).ShouldNot(HaveOccurred())

			task, err := client.GetTask("task-guid-1")
			Ω(err).ShouldNot(HaveOccurred())

			Ω(task.FailureReason).Should(Equal("the-failure-reason"))
			Ω(task.Failed).Should(Equal(true))
			Ω(task.Result).Should(Equal("the-task-result"))
			Ω(task.State).Should(Equal(receptor.TaskStateCompleted))
		})

		Context("when the task doesn't exist", func() {
			It("responds with a TaskNotFound error", func() {
				_, err := client.GetTask("some-other-task-guid")
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.TaskNotFound))
			})
		})
	})

	Describe("DELETE /v1/tasks/:task_guid", func() {
		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "stack-1",
				Action:   &models.RunAction{Path: "/bin/true"},
			}

			err := bbs.DesireTask(logger, task)
			Ω(err).ShouldNot(HaveOccurred())

			_, err = bbs.StartTask(logger, "task-guid-1", "the-cell-id")
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when the task is in the COMPLETED state", func() {
			BeforeEach(func() {
				err := bbs.CompleteTask(logger, "task-guid-1", "the-cell-id", false, "", "the-task-result")
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("deletes the task", func() {
				err := client.DeleteTask("task-guid-1")
				Ω(err).ShouldNot(HaveOccurred())

				_, err = bbs.TaskByGuid("task-guid-1")
				Ω(err).Should(Equal(bbserrors.ErrStoreResourceNotFound))
			})
		})

		Context("when the task is *not* in the COMPLETED state", func() {
			It("returns an error", func() {
				err := client.DeleteTask("task-guid-1")
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.TaskNotDeletable))
			})

			It("does not delete the task", func() {
				client.DeleteTask("task-guid-1")
				_, err := bbs.TaskByGuid("task-guid-1")
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the task does not exist", func() {
			It("returns a TaskNotFound error", func() {
				err := client.DeleteTask("some-other-task-guid")
				Ω(err).Should(HaveOccurred())
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.TaskNotFound))
			})
		})
	})

	Describe("POST /v1/tasks/:task_guid/cancel", func() {
		var cancelErr error

		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				Stack:    "stack-1",
				Action:   &models.RunAction{Path: "/bin/true"},
			}

			err := bbs.DesireTask(logger, task)
			Ω(err).ShouldNot(HaveOccurred())

			_, err = bbs.StartTask(logger, "task-guid-1", "the-cell-id")
			Ω(err).ShouldNot(HaveOccurred())

			cancelErr = client.CancelTask("task-guid-1")
		})

		It("cancels the task", func() {
			task, err := bbs.TaskByGuid("task-guid-1")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(task.State).Should(Equal(models.TaskStateCompleted))
		})

		It("does not error", func() {
			Ω(cancelErr).ShouldNot(HaveOccurred())
		})
	})

})
