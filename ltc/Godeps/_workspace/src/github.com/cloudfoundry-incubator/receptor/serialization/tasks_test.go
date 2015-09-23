package serialization_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Serialization", func() {
	Describe("TaskToResponse", func() {
		var task *models.Task

		BeforeEach(func() {
			task = &models.Task{
				TaskGuid: "the-task-guid",
				Domain:   "the-domain",
				RootFs:   "the-rootfs",
				CellId:   "the-cell-id",
				Action: models.WrapAction(&models.UploadAction{
					From: "from",
					To:   "to",
				}),
				MemoryMb:    100,
				DiskMb:      100,
				CpuWeight:   50,
				Privileged:  true,
				LogGuid:     "the-log-guid",
				LogSource:   "the-source-name",
				MetricsGuid: "the-metrics-guid",
				Annotation:  "the-annotation",

				CreatedAt:     1234,
				FailureReason: "the-failure-reason",
				Failed:        true,
				Result:        "the-result",
				State:         models.Task_Invalid,
				EnvironmentVariables: []*models.EnvironmentVariable{
					{Name: "var1", Value: "val1"},
					{Name: "var2", Value: "val2"},
				},
				EgressRules: []*models.SecurityGroupRule{
					{
						Protocol:     "tcp",
						Destinations: []string{"0.0.0.0/0"},
						Ports:        []uint32{80, 443},
						Log:          true,
					},
				},
			}
		})

		It("serializes the state", func() {
			EXPECTED_STATE_MAP := map[models.Task_State]string{
				models.Task_Invalid:   "INVALID",
				models.Task_Pending:   "PENDING",
				models.Task_Running:   "RUNNING",
				models.Task_Completed: "COMPLETED",
				models.Task_Resolving: "RESOLVING",
			}

			for modelState, jsonState := range EXPECTED_STATE_MAP {
				task.State = modelState
				Expect(serialization.TaskToResponse(task).State).To(Equal(jsonState))
			}
		})

		It("serializes the task's fields", func() {
			actualResponse := serialization.TaskToResponse(task)

			expectedResponse := receptor.TaskResponse{
				TaskGuid: "the-task-guid",
				Domain:   "the-domain",
				RootFS:   "the-rootfs",
				CellID:   "the-cell-id",
				Action: models.WrapAction(&models.UploadAction{
					From: "from",
					To:   "to",
				}),
				MemoryMB:    100,
				DiskMB:      100,
				CPUWeight:   50,
				Privileged:  true,
				LogGuid:     "the-log-guid",
				LogSource:   "the-source-name",
				MetricsGuid: "the-metrics-guid",
				Annotation:  "the-annotation",

				CreatedAt:     1234,
				FailureReason: "the-failure-reason",
				Failed:        true,
				Result:        "the-result",
				State:         receptor.TaskStateInvalid,
				EnvironmentVariables: []*models.EnvironmentVariable{
					{Name: "var1", Value: "val1"},
					{Name: "var2", Value: "val2"},
				},
				EgressRules: []*models.SecurityGroupRule{
					{
						Protocol:     "tcp",
						Destinations: []string{"0.0.0.0/0"},
						Ports:        []uint32{80, 443},
						Log:          true,
					},
				},
			}

			Expect(actualResponse).To(Equal(expectedResponse))
		})

		Context("when the task has a CompletionCallbackURL", func() {
			BeforeEach(func() {
				task.CompletionCallbackUrl = "http://example.com/the-path"
			})

			It("serializes the completion callback URL", func() {
				Expect(serialization.TaskToResponse(task).CompletionCallbackURL).To(Equal("http://example.com/the-path"))
			})
		})

		Context("when the task doesn't have a CompletionCallbackURL", func() {
			It("leaves the completion callback URL blank", func() {
				Expect(serialization.TaskToResponse(task).CompletionCallbackURL).To(Equal(""))
			})
		})
	})

	Describe("TaskFromRequest", func() {
		var request receptor.TaskCreateRequest
		var expectedTask *models.Task

		BeforeEach(func() {
			request = receptor.TaskCreateRequest{
				TaskGuid: "the-task-guid",
				Domain:   "the-domain",
				RootFS:   "the-rootfs",
				Action: models.WrapAction(&models.RunAction{
					User: "me",
					Path: "the-path",
				}),
				MemoryMB:    100,
				DiskMB:      100,
				CPUWeight:   50,
				Privileged:  true,
				LogGuid:     "the-log-guid",
				LogSource:   "the-source-name",
				MetricsGuid: "the-metrics-guid",
				ResultFile:  "the/result/file",
				Annotation:  "the-annotation",
				EnvironmentVariables: []*models.EnvironmentVariable{
					{Name: "var1", Value: "val1"},
					{Name: "var2", Value: "val2"},
				},
				EgressRules: []*models.SecurityGroupRule{
					{
						Protocol:     "tcp",
						Destinations: []string{"0.0.0.0/0"},
						PortRange: &models.PortRange{
							Start: 1,
							End:   1024,
						},
					},
				},
				CompletionCallbackURL: "http://stager.service.discovery.thing/endpoint",
			}

			expectedTask = &models.Task{
				TaskGuid: "the-task-guid",
				Domain:   "the-domain",
				RootFs:   "the-rootfs",
				Action: models.WrapAction(&models.RunAction{
					User: "me",
					Path: "the-path",
				}),
				MemoryMb:    100,
				DiskMb:      100,
				CpuWeight:   50,
				Privileged:  true,
				LogGuid:     "the-log-guid",
				LogSource:   "the-source-name",
				MetricsGuid: "the-metrics-guid",
				ResultFile:  "the/result/file",
				Annotation:  "the-annotation",
				EnvironmentVariables: []*models.EnvironmentVariable{
					{Name: "var1", Value: "val1"},
					{Name: "var2", Value: "val2"},
				},
				EgressRules: []*models.SecurityGroupRule{
					{
						Protocol:     "tcp",
						Destinations: []string{"0.0.0.0/0"},
						PortRange: &models.PortRange{
							Start: 1,
							End:   1024,
						},
					},
				},
				CompletionCallbackUrl: "http://stager.service.discovery.thing/endpoint",
			}
		})

		It("translates the request into a task model, preserving attributes", func() {
			actualTask, err := serialization.TaskFromRequest(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualTask).To(Equal(expectedTask))
		})
	})
})
