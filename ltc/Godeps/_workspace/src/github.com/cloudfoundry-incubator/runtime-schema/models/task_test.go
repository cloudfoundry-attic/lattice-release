package models_test

import (
	"encoding/json"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("Task", func() {
	var taskPayload string
	var task Task

	BeforeEach(func() {
		taskPayload = `{
		"task_guid":"some-guid",
		"domain":"some-domain",
		"rootfs": "docker:///docker.com/docker",
		"env":[
			{
				"name":"ENV_VAR_NAME",
				"value":"an environmment value"
			}
		],
		"cell_id":"cell",
		"action": {
			"download":{
				"from":"old_location",
				"to":"new_location",
				"cache_key":"the-cache-key"
			}
		},
		"result_file":"some-file.txt",
		"result": "turboencabulated",
		"failed":true,
		"failure_reason":"because i said so",
		"memory_mb":256,
		"disk_mb":1024,
		"cpu_weight": 42,
		"privileged": true,
		"log_guid": "123",
		"log_source": "APP",
		"metrics_guid": "456",
		"created_at": 1393371971000000000,
		"updated_at": 1393371971000000010,
		"first_completed_at": 1393371971000000030,
		"state": 1,
		"annotation": "[{\"anything\": \"you want!\"}]... dude",
		"egress_rules": [
		  {
				"protocol": "tcp",
				"destinations": ["0.0.0.0/0"],
				"port_range": {
					"start": 1,
					"end": 1024
				},
				"log": true
			},
		  {
				"protocol": "udp",
				"destinations": ["8.8.0.0/16"],
				"ports": [53],
				"log": false
			}
		]
	}`

		task = Task{
			TaskGuid: "some-guid",
			Domain:   "some-domain",
			RootFS:   "docker:///docker.com/docker",
			EnvironmentVariables: []EnvironmentVariable{
				{
					Name:  "ENV_VAR_NAME",
					Value: "an environmment value",
				},
			},
			Action: &DownloadAction{
				From:     "old_location",
				To:       "new_location",
				CacheKey: "the-cache-key",
			},
			MemoryMB:         256,
			DiskMB:           1024,
			CPUWeight:        42,
			Privileged:       true,
			LogGuid:          "123",
			LogSource:        "APP",
			MetricsGuid:      "456",
			CreatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 00, time.UTC).UnixNano(),
			UpdatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 10, time.UTC).UnixNano(),
			FirstCompletedAt: time.Date(2014, time.February, 25, 23, 46, 11, 30, time.UTC).UnixNano(),
			ResultFile:       "some-file.txt",
			State:            TaskStatePending,
			CellID:           "cell",

			Result:        "turboencabulated",
			Failed:        true,
			FailureReason: "because i said so",

			EgressRules: []SecurityGroupRule{
				{
					Protocol:     "tcp",
					Destinations: []string{"0.0.0.0/0"},
					PortRange: &PortRange{
						Start: 1,
						End:   1024,
					},
					Log: true,
				},
				{
					Protocol:     "udp",
					Destinations: []string{"8.8.0.0/16"},
					Ports:        []uint16{53},
				},
			},

			Annotation: `[{"anything": "you want!"}]... dude`,
		}
	})

	Describe("Validate", func() {
		Context("when the task has a domain, valid guid, stack, and valid action", func() {
			It("is valid", func() {
				task = Task{
					Domain:   "some-domain",
					TaskGuid: "some-task-guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
				}

				err := task.Validate()
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the task GUID is present but invalid", func() {
			It("returns an error indicating so", func() {
				task = Task{
					Domain:   "some-domain",
					TaskGuid: "invalid/guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
				}

				err := task.Validate()
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("task_guid"))
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{
				"task_guid",
				Task{
					Domain: "some-domain",
					RootFS: "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
				},
			},
			{
				"rootfs",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					Action: &RunAction{
						Path: "ls",
					},
				},
			},
			{
				"rootfs",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   ":invalid-url",
					Action: &RunAction{
						Path: "ls",
					},
				},
			},
			{
				"rootfs",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "invalid-absolute-url",
					Action: &RunAction{
						Path: "ls",
					},
				},
			},
			{
				"domain",
				Task{
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
				},
			},
			{
				"action",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
				}},
			{
				"path",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
					Action:   &RunAction{},
				},
			},
			{
				"annotation",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
					Annotation: strings.Repeat("a", 10*1024+1),
				},
			},
			{
				"cpu_weight",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
					CPUWeight: 101,
				},
			},
			{
				"egress_rules",
				Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					RootFS:   "some:rootfs",
					Action: &RunAction{
						Path: "ls",
					},
					EgressRules: []SecurityGroupRule{
						{Protocol: "invalid"},
					},
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})

	Describe("Marshal", func() {
		It("should JSONify", func() {
			json, err := ToJSON(&task)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(json)).Should(MatchJSON(taskPayload))
		})
	})

	Describe("Unmarshal", func() {
		It("returns a Task with correct fields", func() {
			decodedTask := &Task{}
			err := FromJSON([]byte(taskPayload), decodedTask)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedTask).Should(Equal(&task))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedTask := &Task{}
				err := FromJSON([]byte("aliens lol"), decodedTask)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("with invalid action", func() {
			var expectedTask Task
			var taskJSON string

			BeforeEach(func() {
				expectedTask = Task{
					TaskGuid: "some-guid",
					Domain:   "some-domain",
					RootFS:   "some:rootfs",
				}
			})

			Context("with null action", func() {
				BeforeEach(func() {
					taskJSON = `{
					"task_guid":"some-guid",
					"domain":"some-domain",
					"action": null,
					"rootfs":"some:rootfs"
				}`
				})

				It("unmarshals", func() {
					var actualTask Task
					err := json.Unmarshal([]byte(taskJSON), &actualTask)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(actualTask).Should(Equal(expectedTask))
				})
			})

			Context("with missing action", func() {
				BeforeEach(func() {
					taskJSON = `{
					"task_guid":"some-guid",
					"domain":"some-domain",
					"rootfs":"some:rootfs"
				}`
				})

				It("unmarshals", func() {
					var actualTask Task
					err := json.Unmarshal([]byte(taskJSON), &actualTask)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(actualTask).Should(Equal(expectedTask))
				})
			})
		})
	})
})
