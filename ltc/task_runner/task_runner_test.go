package task_runner_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/fake_task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("TaskRunner", func() {
	var (
		fakeReceptorClient *fake_receptor.FakeClient
		taskRunner         task_runner.TaskRunner
		fakeTaskExaminer   *fake_task_examiner.FakeTaskExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		fakeTaskExaminer = new(fake_task_examiner.FakeTaskExaminer)
		taskRunner = task_runner.New(fakeReceptorClient, fakeTaskExaminer)
	})

	Describe("CreateTask", func() {
		var (
			action             models.Action
			securityGroupRules []models.SecurityGroupRule
			createTaskParams   task_runner.CreateTaskParams
		)

		BeforeEach(func() {
			action = &models.RunAction{
				Path: "/my/path",
				Args: []string{"happy", "sad"},
				Dir:  "/my",
				Env: []models.EnvironmentVariable{
					models.EnvironmentVariable{"env-name", "env-value"},
				},
				ResourceLimits: models.ResourceLimits{},
				LogSource:      "log-source",
			}
			securityGroupRules = []models.SecurityGroupRule{
				models.SecurityGroupRule{
					Protocol:     models.TCPProtocol,
					Destinations: []string{"bermuda", "bahamas"},
					Ports:        []uint16{4242, 5353},
					PortRange:    &models.PortRange{6666, 7777},
					Log:          true,
				},
			}
			createTaskParams = task_runner.NewCreateTaskParams(
				action,
				"task-name",
				"preloaded:my-rootfs",
				"task-domain",
				"log-source",
				map[string]string{
					"MaRTY": "BiSHoP",
					"CoSMo": "CRaMeR",
				},
				securityGroupRules,
				128,
				100,
				0,
			)
		})

		It("creates a task", func() {
			err := taskRunner.CreateTask(createTaskParams)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))

			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(BeZero())

			Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(1))
			createTaskRequest := fakeReceptorClient.CreateTaskArgsForCall(0)
			Expect(createTaskRequest).ToNot(BeNil())
			Expect(createTaskRequest.Action).To(Equal(action))
			Expect(createTaskRequest.TaskGuid).To(Equal("task-name"))
			Expect(createTaskRequest.LogGuid).To(Equal("task-name"))
			Expect(createTaskRequest.MetricsGuid).To(Equal("task-name"))
			Expect(createTaskRequest.RootFS).To(Equal("preloaded:my-rootfs"))
			Expect(createTaskRequest.Domain).To(Equal("task-domain"))
			Expect(createTaskRequest.LogSource).To(Equal("log-source"))
			Expect(createTaskRequest.EnvironmentVariables).To(ConsistOf(
				receptor.EnvironmentVariable{"MaRTY", "BiSHoP"},
				receptor.EnvironmentVariable{"CoSMo", "CRaMeR"},
			))
			Expect(createTaskRequest.EgressRules).To(Equal(securityGroupRules))
		})

		Context("when the task already exists", func() {
			It("doesn't allow you", func() {
				tasksResponse := []receptor.TaskResponse{
					receptor.TaskResponse{TaskGuid: "task-name"},
				}
				fakeReceptorClient.TasksReturns(tasksResponse, nil)

				err := taskRunner.CreateTask(createTaskParams)
				Expect(err).To(MatchError("task-name has already been submitted"))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(0))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})
		})

		Context("when the guid is the reserved name lattice-debug", func() {
			It("doesn't allow you", func() {
				createTaskParams = task_runner.NewCreateTaskParams(
					action,
					"lattice-debug",
					"preloaded:my-rootfs",
					"task-domain",
					"log-source",
					nil,
					securityGroupRules,
					128,
					100,
					0,
				)

				err := taskRunner.CreateTask(createTaskParams)
				Expect(err).To(MatchError(task_runner.AttemptedToCreateLatticeDebugErrorMessage))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(0))
				Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(0))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})
		})

		Context("when the receptor returns errors", func() {
			It("returns error when getting existing tasks", func() {
				fakeReceptorClient.TasksReturns(nil, errors.New("unable to fetch tasks"))

				err := taskRunner.CreateTask(createTaskParams)
				Expect(err).To(MatchError("unable to fetch tasks"))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(0))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})

			It("returns error when upserting the domain", func() {
				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)

				err := taskRunner.CreateTask(createTaskParams)
				Expect(err).To(MatchError(upsertError))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})

			It("returns error when creating the task fails", func() {
				fakeReceptorClient.CreateTaskReturns(errors.New("not making your task"))

				err := taskRunner.CreateTask(createTaskParams)
				Expect(err).To(MatchError("not making your task"))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(1))
			})
		})
	})

	Describe("SubmitTask", func() {
		It("Submits a task from JSON", func() {
			environmentVariables := []receptor.EnvironmentVariable{
				receptor.EnvironmentVariable{
					Name:  "beans",
					Value: "cool",
				}}
			egressRules := []models.SecurityGroupRule{
				models.SecurityGroupRule{
					Protocol:     models.UDPProtocol,
					Destinations: []string{"dest1", "dest2"},
					Ports:        []uint16{1717, 2828},
					PortRange: &models.PortRange{
						Start: 1000,
						End:   2000,
					},
					IcmpInfo: &models.ICMPInfo{
						Code: 17,
						Type: 44,
					},
					Log: true,
				},
			}

			task := receptor.TaskCreateRequest{
				Action: &models.DownloadAction{
					From:      "/tmp/here",
					To:        "/tmp/there",
					LogSource: "MOVING",
					User:      "downloader-man",
				},
				Annotation:            "blah blah",
				CompletionCallbackURL: "http://sup.com",
				CPUWeight:             1000,
				DiskMB:                2,
				Domain:                "lattice",
				LogGuid:               "loggy-logs",
				LogSource:             "loggy-logs-be-here",
				MemoryMB:              128,
				ResultFile:            "i/come/back/here",
				TaskGuid:              "lattice-task",
				RootFS:                "docker://sup",
				Privileged:            true,
				EnvironmentVariables:  environmentVariables,
				EgressRules:           egressRules,
			}
			taskJson, err := json.Marshal(task)
			Expect(err).ToNot(HaveOccurred())

			taskName, err := taskRunner.SubmitTask(taskJson)
			Expect(err).NotTo(HaveOccurred())
			Expect(taskName).To(Equal("lattice-task"))

			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(BeZero())

			Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(1))
			taskRequest := fakeReceptorClient.CreateTaskArgsForCall(0)
			Expect(taskRequest).ToNot(BeZero())

			Expect(taskRequest.Action).To(Equal(&models.DownloadAction{
				From:      "/tmp/here",
				To:        "/tmp/there",
				LogSource: "MOVING",
				User:      "downloader-man",
			}))
			Expect(taskRequest.TaskGuid).To(Equal("lattice-task"))
			Expect(taskRequest.Annotation).To(Equal("blah blah"))
			Expect(taskRequest.CompletionCallbackURL).To(Equal("http://sup.com"))
			Expect(taskRequest.CPUWeight).To(Equal(uint(1000)))
			Expect(taskRequest.DiskMB).To(Equal(2))
			Expect(taskRequest.Domain).To(Equal("lattice"))
			Expect(taskRequest.LogGuid).To(Equal("loggy-logs"))
			Expect(taskRequest.LogSource).To(Equal("loggy-logs-be-here"))
			Expect(taskRequest.MemoryMB).To(Equal(128))
			Expect(taskRequest.ResultFile).To(Equal("i/come/back/here"))
			Expect(taskRequest.RootFS).To(Equal("docker://sup"))
			Expect(taskRequest.Privileged).To(BeTrue())
			Expect(taskRequest.EnvironmentVariables).To(HaveLen(1))
			Expect(taskRequest.EnvironmentVariables[0]).To(BeEquivalentTo(environmentVariables[0]))
			Expect(taskRequest.EgressRules).To(HaveLen(1))
			Expect(taskRequest.EgressRules[0]).To(Equal(egressRules[0]))
		})

		Context("when the task already exists", func() {
			It("aborts", func() {
				tasksResponse := []receptor.TaskResponse{
					receptor.TaskResponse{TaskGuid: "task-already-submitted"},
				}
				fakeReceptorClient.TasksReturns(tasksResponse, nil)

				task := receptor.TaskCreateRequest{
					TaskGuid: "task-already-submitted",
				}
				taskJson, err := json.Marshal(task)
				Expect(err).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)
				Expect(err).To(MatchError("task-already-submitted has already been submitted"))
				Expect(taskName).To(Equal("task-already-submitted"))

				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})
		})

		Context("when 'lattice-debug' is passed as the taskGuid", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				task := receptor.TaskCreateRequest{
					TaskGuid: "lattice-debug",
				}
				taskJson, err := json.Marshal(task)
				Expect(err).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)
				Expect(err).To(MatchError(task_runner.AttemptedToCreateLatticeDebugErrorMessage))
				Expect(taskName).To(Equal("lattice-debug"))

				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})
		})

		It("returns an error for invalid JSON", func() {
			taskName, err := taskRunner.SubmitTask([]byte(`{"Value":"test value`))
			Expect(err).To(MatchError("unexpected end of JSON input"))
			Expect(taskName).To(BeEmpty())

			Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
		})

		Context("when the receptor returns errors", func() {

			It("returns upsert domain errors", func() {
				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)
				task := receptor.TaskCreateRequest{
					TaskGuid: "whatever-task",
				}
				taskJson, err := json.Marshal(task)
				Expect(err).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)
				Expect(err).To(MatchError(upsertError))
				Expect(taskName).To(Equal("whatever-task"))

				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})

			It("returns tasks errors", func() {
				tasksResponseError := errors.New("wut")
				fakeReceptorClient.TasksReturns(nil, tasksResponseError)
				task := receptor.TaskCreateRequest{
					TaskGuid: "whatever-task",
				}
				taskJson, err := json.Marshal(task)
				Expect(err).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)
				Expect(err).To(MatchError(tasksResponseError))
				Expect(taskName).To(Equal("whatever-task"))

				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})

			It("returns create task errors", func() {
				receptorError := errors.New("you got tasked")
				fakeReceptorClient.CreateTaskReturns(receptorError)
				task := receptor.TaskCreateRequest{
					TaskGuid: "whatever-task",
				}
				taskJson, err := json.Marshal(task)
				Expect(err).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)
				Expect(err).To(MatchError(receptorError))
				Expect(taskName).To(Equal("whatever-task"))

				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(1))
			})

		})
	})

	Describe("Delete Task", func() {
		getTaskStatus := func(state string) task_examiner.TaskInfo {
			return task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    state,
			}
		}

		It("delete task when task in COMPLETED state", func() {
			fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStateCompleted), nil)

			err := taskRunner.DeleteTask("task-guid-1")
			Expect(err).ToNot(HaveOccurred())
		})

		It("return error when task is not in COMPLETED state", func() {
			fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStatePending), nil)

			err := taskRunner.DeleteTask("task-guid-1")
			Expect(err).To(MatchError("task-guid-1 is not in COMPLETED state"))
		})

		Context("when the receptor returns errors", func() {
			It("bubbles up the error from task_examiner.TaskStatus", func() {
				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New("Task not found"))

				err := taskRunner.DeleteTask("task-guid-1")
				Expect(err).To(MatchError("Task not found"))
			})

			It("returns error when not able to delete the task", func() {
				fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStateCompleted), nil)
				fakeReceptorClient.DeleteTaskReturns(errors.New("task in unknown state"))

				err := taskRunner.DeleteTask("task-guid-1")
				Expect(err).To(MatchError("task in unknown state"))
			})
		})
	})

	Describe("Cancel Task", func() {
		getTaskStatus := func(state string) task_examiner.TaskInfo {
			return task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    state,
			}
		}

		It("cancel task when task not in COMPLETED state", func() {
			fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStatePending), nil)

			err := taskRunner.CancelTask("task-guid-1")
			Expect(err).ToNot(HaveOccurred())
		})

		It("return error when task in COMPLETED state", func() {
			fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStateCompleted), nil)

			err := taskRunner.CancelTask("task-guid-1")
			Expect(err).To(MatchError("Unable to cancel COMPLETED task"))
		})

		It("returns error when task not found", func() {
			fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New("Task not found"))

			err := taskRunner.CancelTask("task-guid-1")
			Expect(err).To(MatchError("Task not found"))
		})

		Context("when the receptor returns errors", func() {
			It("bubbles up the error", func() {
				fakeTaskExaminer.TaskStatusReturns(getTaskStatus(receptor.TaskStatePending), nil)
				fakeReceptorClient.CancelTaskReturns(errors.New("task in unknown state"))

				err := taskRunner.CancelTask("task-guid-1")
				Expect(err).To(MatchError("task in unknown state"))
			})
		})
	})
})
