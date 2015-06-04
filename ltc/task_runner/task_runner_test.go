package task_runner_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("TaskRunner", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		taskRunner         task_runner.TaskRunner
		taskExaminer       task_examiner.TaskExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		taskExaminer = task_examiner.New(fakeReceptorClient)
		taskRunner = task_runner.New(fakeReceptorClient, taskExaminer)
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
			taskJson, marshalErr := json.Marshal(task)
			Expect(marshalErr).ToNot(HaveOccurred())

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
				taskJson, marshalErr := json.Marshal(task)
				Expect(marshalErr).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)

				Expect(err).To(HaveOccurred())
				Expect(taskName).To(Equal("task-already-submitted"))

				Expect(err).To(MatchError("task-already-submitted has already been submitted"))
				Expect(fakeReceptorClient.TasksCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(0))
			})
		})

		Context("when 'lattice-debug' is passed as the taskGuid", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				task := receptor.TaskCreateRequest{
					TaskGuid: "lattice-debug",
				}
				taskJson, marshalErr := json.Marshal(task)
				Expect(marshalErr).ToNot(HaveOccurred())

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
				taskJson, marshalErr := json.Marshal(task)
				Expect(marshalErr).ToNot(HaveOccurred())

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
				taskJson, marshalErr := json.Marshal(task)
				Expect(marshalErr).ToNot(HaveOccurred())

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
				taskJson, marshalErr := json.Marshal(task)
				Expect(marshalErr).ToNot(HaveOccurred())

				taskName, err := taskRunner.SubmitTask(taskJson)

				Expect(err).To(MatchError(receptorError))
				Expect(taskName).To(Equal("whatever-task"))
				Expect(fakeReceptorClient.CreateTaskCallCount()).To(Equal(1))
			})

		})
	})
	Describe("Delete Task", func() {
		It("delete task when task in COMPLETED state", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStateCompleted,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)

			err := taskRunner.DeleteTask("task-guid-1")

			Expect(err).ToNot(HaveOccurred())
		})

		It("return error when task is not in COMPLETED state", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStatePending,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)

			err := taskRunner.DeleteTask("task-guid-1")
			Expect(err).To(MatchError("task-guid-1 has not completed"))
		})

		It("returns error when task not found", func() {
			fakeReceptorClient.GetTaskReturns(receptor.TaskResponse{}, errors.New("Task not found"))

			err := taskRunner.DeleteTask("task-guid-1")

			Expect(err).To(MatchError("Task not found"))
		})

		It("returns error when not able to delete the task", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStateCompleted,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)
			fakeReceptorClient.DeleteTaskReturns(errors.New("task in unknown state"))

			err := taskRunner.DeleteTask("task-guid-1")
			Expect(err).To(MatchError("task in unknown state"))
		})
	})
	Describe("Cancel Task", func() {
		It("cancel task when task not in COMPLETED state", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStatePending,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)

			err := taskRunner.CancelTask("task-guid-1")

			Expect(err).ToNot(HaveOccurred())
		})

		It("return success when task in COMPLETED state", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStateCompleted,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)

			err := taskRunner.CancelTask("task-guid-1")

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error when task not found", func() {
			fakeReceptorClient.GetTaskReturns(receptor.TaskResponse{}, errors.New("Task not found"))

			err := taskRunner.CancelTask("task-guid-1")

			Expect(err).To(MatchError("Task not found"))
		})

		It("returns error when not able to cancel the task", func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid: "task-guid-1",
				State:    receptor.TaskStatePending,
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)
			fakeReceptorClient.CancelTaskReturns(errors.New("task in unknown state"))

			err := taskRunner.CancelTask("task-guid-1")
			Expect(err).To(MatchError("task in unknown state"))
		})
	})

})
