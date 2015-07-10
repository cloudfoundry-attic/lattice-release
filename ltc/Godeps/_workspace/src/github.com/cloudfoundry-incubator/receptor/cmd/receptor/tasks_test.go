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
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Header.Get("Content-Length")).To(MatchRegexp(`\d+`))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))
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
				TaskGuid:              "task-guid-1",
				Domain:                "test-domain",
				RootFS:                "some:rootfs",
				CompletionCallbackURL: testServer.URL() + "/the/callback/path",
				Action:                &models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}},
			}

			err = client.CreateTask(taskToCreate)
		})

		It("responds without an error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("desires the task in the BBS", func() {
			Eventually(func() ([]models.Task, error) {
				return legacyBBS.PendingTasks(logger)
			}).Should(HaveLen(1))
		})

		Context("when trying to create a task with a GUID that already exists", func() {
			BeforeEach(func() {
				err = client.CreateTask(taskToCreate)
			})

			It("returns an error indicating that the key already exists", func() {
				Expect(err.(receptor.Error).Type).To(Equal(receptor.TaskGuidAlreadyExists))
			})
		})

		Describe("when the task completes", func() {
			BeforeEach(func() {
				_, err = legacyBBS.StartTask(logger, "task-guid-1", "the-cell-id")
				Expect(err).NotTo(HaveOccurred())
			})

			It("sends a POST request to the specified callback URL", func() {
				testServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/the/callback/path"),
					func(res http.ResponseWriter, req *http.Request) {
						var taskResponse receptor.TaskResponse
						err := json.NewDecoder(req.Body).Decode(&taskResponse)
						Expect(err).NotTo(HaveOccurred())

						Expect(taskResponse.TaskGuid).To(Equal("task-guid-1"))
						Expect(taskResponse.Domain).To(Equal("test-domain"))
						Expect(taskResponse.RootFS).To(Equal("some:rootfs"))
						Expect(taskResponse.State).To(Equal(receptor.TaskStateCompleted))
						Expect(taskResponse.Result).To(Equal("the-result"))
						Expect(taskResponse.Failed).To(Equal(true))
						Expect(taskResponse.FailureReason).To(Equal("the-failure-reason"))
						Expect(taskResponse.Action).To(Equal(&models.RunAction{User: "me", Path: "/bin/bash", Args: []string{"echo", "hi"}}))
					},
				))

				Expect(testServer.ReceivedRequests()).To(HaveLen(0))

				err = legacyBBS.CompleteTask(logger, "task-guid-1", "the-cell-id", true, "the-failure-reason", "the-result")
				Expect(err).NotTo(HaveOccurred())

				Eventually(testServer.ReceivedRequests).Should(HaveLen(1))
			})
		})
	})

	Describe("GET /v1/tasks", func() {
		Context("when there are no tasks", func() {
			It("returns an empty array", func() {
				tasks, err := client.Tasks()
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(BeEmpty())
			})
		})

		Context("when there are tasks", func() {
			BeforeEach(func() {
				err := legacyBBS.DesireTask(logger, models.Task{
					TaskGuid: "task-guid-1",
					Domain:   "test-domain",
					RootFS:   "some:rootfs",
					Action:   &models.RunAction{User: "me", Path: "/bin/true"},
				})
				Expect(err).NotTo(HaveOccurred())

				err = legacyBBS.DesireTask(logger, models.Task{
					TaskGuid: "task-guid-2",
					Domain:   "test-domain",
					RootFS:   "some:rootfs",
					Action:   &models.RunAction{User: "me", Path: "/bin/true"},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an array of all the tasks", func() {
				tasks, err := client.Tasks()
				Expect(err).NotTo(HaveOccurred())

				taskGuids := []string{}
				for _, task := range tasks {
					taskGuids = append(taskGuids, task.TaskGuid)
				}
				Expect(taskGuids).To(ConsistOf([]string{"task-guid-1", "task-guid-2"}))
			})

		})
	})

	Describe("GET /v1/domains/:domain/tasks", func() {
		BeforeEach(func() {
			err := legacyBBS.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			})
			Expect(err).NotTo(HaveOccurred())

			err = legacyBBS.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-2",
				Domain:   "other-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			})
			Expect(err).NotTo(HaveOccurred())

			err = legacyBBS.DesireTask(logger, models.Task{
				TaskGuid: "task-guid-3",
				Domain:   "test-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an array of all the tasks for the domain", func() {
			tasks, err := client.TasksByDomain("test-domain")
			Expect(err).NotTo(HaveOccurred())

			taskGuids := []string{}
			for _, task := range tasks {
				taskGuids = append(taskGuids, task.TaskGuid)
			}
			Expect(taskGuids).To(ConsistOf([]string{"task-guid-1", "task-guid-3"}))
		})
	})

	Describe("GET /v1/tasks/:task_guid", func() {
		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			}
			err := legacyBBS.DesireTask(logger, task)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the task", func() {
			task, err := client.GetTask("task-guid-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(task.TaskGuid).To(Equal("task-guid-1"))
			Expect(task.Domain).To(Equal("test-domain"))
		})

		It("includes all of the task's publicly-visible fields", func() {
			_, err := legacyBBS.StartTask(logger, "task-guid-1", "the-cell-id")
			Expect(err).NotTo(HaveOccurred())
			err = legacyBBS.CompleteTask(logger, "task-guid-1", "the-cell-id", true, "the-failure-reason", "the-task-result")
			Expect(err).NotTo(HaveOccurred())

			task, err := client.GetTask("task-guid-1")
			Expect(err).NotTo(HaveOccurred())

			Expect(task.FailureReason).To(Equal("the-failure-reason"))
			Expect(task.Failed).To(Equal(true))
			Expect(task.Result).To(Equal("the-task-result"))
			Expect(task.State).To(Equal(receptor.TaskStateCompleted))
		})

		Context("when the task doesn't exist", func() {
			It("responds with a TaskNotFound error", func() {
				_, err := client.GetTask("some-other-task-guid")
				Expect(err.(receptor.Error).Type).To(Equal(receptor.TaskNotFound))
			})
		})
	})

	Describe("DELETE /v1/tasks/:task_guid", func() {
		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			}

			err := legacyBBS.DesireTask(logger, task)
			Expect(err).NotTo(HaveOccurred())

			_, err = legacyBBS.StartTask(logger, "task-guid-1", "the-cell-id")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the task is in the COMPLETED state", func() {
			BeforeEach(func() {
				err := legacyBBS.CompleteTask(logger, "task-guid-1", "the-cell-id", false, "", "the-task-result")
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes the task", func() {
				err := client.DeleteTask("task-guid-1")
				Expect(err).NotTo(HaveOccurred())

				_, err = legacyBBS.TaskByGuid(logger, "task-guid-1")
				Expect(err).To(Equal(bbserrors.ErrStoreResourceNotFound))
			})
		})

		Context("when the task is *not* in the COMPLETED state", func() {
			It("returns an error", func() {
				err := client.DeleteTask("task-guid-1")
				Expect(err.(receptor.Error).Type).To(Equal(receptor.TaskNotDeletable))
			})

			It("does not delete the task", func() {
				client.DeleteTask("task-guid-1")
				_, err := legacyBBS.TaskByGuid(logger, "task-guid-1")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the task does not exist", func() {
			It("returns a TaskNotFound error", func() {
				err := client.DeleteTask("some-other-task-guid")
				Expect(err).To(HaveOccurred())
				Expect(err.(receptor.Error).Type).To(Equal(receptor.TaskNotFound))
			})
		})
	})

	Describe("POST /v1/tasks/:task_guid/cancel", func() {
		var cancelErr error

		BeforeEach(func() {
			task := models.Task{
				TaskGuid: "task-guid-1",
				Domain:   "test-domain",
				RootFS:   "some:rootfs",
				Action:   &models.RunAction{User: "me", Path: "/bin/true"},
			}

			err := legacyBBS.DesireTask(logger, task)
			Expect(err).NotTo(HaveOccurred())

			_, err = legacyBBS.StartTask(logger, "task-guid-1", "the-cell-id")
			Expect(err).NotTo(HaveOccurred())

			cancelErr = client.CancelTask("task-guid-1")
		})

		It("cancels the task", func() {
			task, err := legacyBBS.TaskByGuid(logger, "task-guid-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(task.State).To(Equal(models.TaskStateCompleted))
		})

		It("does not error", func() {
			Expect(cancelErr).NotTo(HaveOccurred())
		})
	})

})
