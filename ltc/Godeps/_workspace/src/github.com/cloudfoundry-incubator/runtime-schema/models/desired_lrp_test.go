package models_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP", func() {
	var lrp DesiredLRP

	lrpPayload := `{
	  "process_guid": "some-guid",
	  "domain": "some-domain",
	  "rootfs": "docker:///docker.com/docker",
	  "instances": 1,
		"annotation": "some-annotation",
		"start_timeout": 0,
	  "env":[
	    {
	      "name": "ENV_VAR_NAME",
	      "value": "some environment variable value"
	    }
	  ],
		"setup": {
			"download": {
				"from": "http://example.com",
				"to": "/tmp/internet",
				"cache_key": ""
			}
		},
		"action": {
			"run": {
				"path": "ls",
				"args": null,
				"env": null,
				"resource_limits":{}
			}
		},
		"monitor": {
			"run": {
				"path": "reboot",
				"args": null,
				"env": null,
				"resource_limits":{}
			}
		},
	  "disk_mb": 512,
	  "memory_mb": 1024,
	  "cpu_weight": 42,
		"privileged": true,
	  "ports": [
	    5678
	  ],
	  "routes": {
	  	"router":	{"port": 8080,"hosts":["route-1","route-2"]}
	  },
	  "log_guid": "log-guid",
	  "log_source": "the cloud",
		"metrics_guid": "metrics-guid",
		"modification_tag": {
			"epoch": "some-epoch",
			"index": 50
		},
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

	BeforeEach(func() {
		rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["route-1","route-2"]}`))
		lrp = DesiredLRP{
			Domain:      "some-domain",
			ProcessGuid: "some-guid",

			Instances:  1,
			RootFS:     "docker:///docker.com/docker",
			MemoryMB:   1024,
			DiskMB:     512,
			CPUWeight:  42,
			Privileged: true,
			Routes: map[string]*json.RawMessage{
				"router": &rawMessage,
			},
			Annotation: "some-annotation",
			Ports: []uint16{
				5678,
			},
			LogGuid:     "log-guid",
			LogSource:   "the cloud",
			MetricsGuid: "metrics-guid",
			ModificationTag: ModificationTag{
				Epoch: "some-epoch",
				Index: 50,
			},
			EnvironmentVariables: []EnvironmentVariable{
				{
					Name:  "ENV_VAR_NAME",
					Value: "some environment variable value",
				},
			},
			Setup: &DownloadAction{
				From: "http://example.com",
				To:   "/tmp/internet",
			},
			Action: &RunAction{
				Path: "ls",
			},
			Monitor: &RunAction{
				Path: "reboot",
			},
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
		}
	})

	Describe("To JSON", func() {
		It("should JSONify", func() {
			json, err := ToJSON(&lrp)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(json)).To(MatchJSON(lrpPayload))
		})
	})

	Describe("ApplyUpdate", func() {
		It("updates instances", func() {
			instances := 100
			update := DesiredLRPUpdate{Instances: &instances}

			expectedLRP := lrp
			expectedLRP.Instances = instances

			updatedLRP := lrp.ApplyUpdate(update)
			Expect(updatedLRP).To(Equal(expectedLRP))
		})

		It("allows empty routes to be set", func() {
			update := DesiredLRPUpdate{
				Routes: map[string]*json.RawMessage{},
			}

			expectedLRP := lrp
			expectedLRP.Routes = map[string]*json.RawMessage{}

			updatedLRP := lrp.ApplyUpdate(update)
			Expect(updatedLRP).To(Equal(expectedLRP))
		})

		It("allows annotation to be set", func() {
			annotation := "new-annotation"
			update := DesiredLRPUpdate{
				Annotation: &annotation,
			}

			expectedLRP := lrp
			expectedLRP.Annotation = annotation

			updatedLRP := lrp.ApplyUpdate(update)
			Expect(updatedLRP).To(Equal(expectedLRP))
		})

		It("allows empty annotation to be set", func() {
			emptyAnnotation := ""
			update := DesiredLRPUpdate{
				Annotation: &emptyAnnotation,
			}

			expectedLRP := lrp
			expectedLRP.Annotation = emptyAnnotation

			updatedLRP := lrp.ApplyUpdate(update)
			Expect(updatedLRP).To(Equal(expectedLRP))
		})

		It("updates routes", func() {
			rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
			update := DesiredLRPUpdate{
				Routes: map[string]*json.RawMessage{
					"router": &rawMessage,
				},
			}

			expectedLRP := lrp
			expectedLRP.Routes = map[string]*json.RawMessage{
				"router": &rawMessage,
			}

			updatedLRP := lrp.ApplyUpdate(update)
			Expect(updatedLRP).To(Equal(expectedLRP))
		})
	})

	Describe("Validate", func() {
		var assertDesiredLRPValidationFailsWithMessage = func(lrp DesiredLRP, substring string) {
			validationErr := lrp.Validate()
			Expect(validationErr).To(HaveOccurred())
			Expect(validationErr.Error()).To(ContainSubstring(substring))
		}

		Context("process_guid only contains `A-Z`, `a-z`, `0-9`, `-`, and `_`", func() {
			validGuids := []string{"a", "A", "0", "-", "_", "-aaaa", "_-aaa", "09a87aaa-_aASKDn"}
			for _, validGuid := range validGuids {
				func(validGuid string) {
					It(fmt.Sprintf("'%s' is a valid process_guid", validGuid), func() {
						lrp.ProcessGuid = validGuid
						err := lrp.Validate()
						Expect(err).NotTo(HaveOccurred())
					})
				}(validGuid)
			}

			invalidGuids := []string{"", "bang!", "!!!", "\\slash", "star*", "params()", "invalid/key", "with.dots"}
			for _, invalidGuid := range invalidGuids {
				func(invalidGuid string) {
					It(fmt.Sprintf("'%s' is an invalid process_guid", invalidGuid), func() {
						lrp.ProcessGuid = invalidGuid
						assertDesiredLRPValidationFailsWithMessage(lrp, "process_guid")
					})
				}(invalidGuid)
			}
		})

		It("requires a positive nonzero number of instances", func() {
			lrp.Instances = -1
			assertDesiredLRPValidationFailsWithMessage(lrp, "instances")

			lrp.Instances = 0
			validationErr := lrp.Validate()
			Expect(validationErr).NotTo(HaveOccurred())

			lrp.Instances = 1
			validationErr = lrp.Validate()
			Expect(validationErr).NotTo(HaveOccurred())
		})

		It("requires a domain", func() {
			lrp.Domain = ""
			assertDesiredLRPValidationFailsWithMessage(lrp, "domain")
		})

		It("requires a rootfs", func() {
			lrp.RootFS = ""
			assertDesiredLRPValidationFailsWithMessage(lrp, "rootfs")
		})

		It("requires a valid URL with a non-empty scheme for the rootfs", func() {
			lrp.RootFS = ":not-a-url"
			assertDesiredLRPValidationFailsWithMessage(lrp, "rootfs")
		})

		It("requires a valid absolute URL for the rootfs", func() {
			lrp.RootFS = "not-an-absolute-url"
			assertDesiredLRPValidationFailsWithMessage(lrp, "rootfs")
		})

		It("requires an action", func() {
			lrp.Action = nil
			assertDesiredLRPValidationFailsWithMessage(lrp, "action")
		})

		It("requires a valid action", func() {
			lrp.Action = &UploadAction{
				From: "web_location",
			}
			assertDesiredLRPValidationFailsWithMessage(lrp, "to")
		})

		It("requires a valid setup action if specified", func() {
			lrp.Setup = &UploadAction{
				From: "web_location",
			}
			assertDesiredLRPValidationFailsWithMessage(lrp, "to")
		})

		It("requires a valid monitor action if specified", func() {
			lrp.Monitor = &UploadAction{
				From: "web_location",
			}
			assertDesiredLRPValidationFailsWithMessage(lrp, "to")
		})

		It("requires a valid CPU weight", func() {
			lrp.CPUWeight = 101
			assertDesiredLRPValidationFailsWithMessage(lrp, "cpu_weight")
		})

		Context("when security group is present", func() {
			It("must be valid", func() {
				lrp.EgressRules = []SecurityGroupRule{{
					Protocol: "foo",
				}}
				assertDesiredLRPValidationFailsWithMessage(lrp, "egress_rules")
			})
		})

		Context("when security group is not present", func() {
			It("does not error", func() {
				lrp.EgressRules = []SecurityGroupRule{}

				validationErr := lrp.Validate()
				Expect(validationErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Unmarshal", func() {
		It("returns a LRP with correct fields", func() {
			decodedLRP := DesiredLRP{}
			err := FromJSON([]byte(lrpPayload), &decodedLRP)
			Expect(err).NotTo(HaveOccurred())

			Expect(decodedLRP).To(Equal(lrp))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedLRP := DesiredLRP{}
				err := FromJSON([]byte("aliens lol"), &decodedLRP)
				Expect(err).To(HaveOccurred())

				Expect(decodedLRP).To(BeZero())
			})
		})

		Context("with a missing action", func() {
			It("returns the error", func() {
				decodedLRP := DesiredLRP{}
				err := FromJSON([]byte(`{
				"domain": "some-domain",
				"process_guid": "process_guid",
				"rootfs": "some-rootfs"
			}`,
				), &decodedLRP)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid actions", func() {
			var expectedLRP DesiredLRP
			var payload string

			BeforeEach(func() {
				expectedLRP = DesiredLRP{}
			})

			Context("with null actions", func() {
				BeforeEach(func() {
					payload = `{
					"setup": null,
					"action": null,
					"monitor": null
				}`
				})

				It("unmarshals", func() {
					var actualLRP DesiredLRP
					err := json.Unmarshal([]byte(payload), &actualLRP)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRP).To(Equal(expectedLRP))
				})
			})

			Context("with missing action", func() {
				BeforeEach(func() {
					payload = `{}`
				})

				It("unmarshals", func() {
					var actualLRP DesiredLRP
					err := json.Unmarshal([]byte(payload), &actualLRP)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRP).To(Equal(expectedLRP))
				})
			})
		})

		for field, payload := range map[string]string{
			"process_guid": `{
				"domain": "some-domain",
				"rootfs": "some-rootfs",
				"action":
					{"download":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
			}`,
			"rootfs": `{
				"domain": "some-domain",
				"process_guid": "process_guid",
				"action":
					{"download":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
			}`,
			"domain": `{
				"rootfs": "some-rootfs",
				"process_guid": "process_guid",
				"action":
					{"download":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
			}`,
		} {
			missingField := field
			jsonBytes := payload

			Context("when the json is missing a "+missingField, func() {
				It("returns an error indicating so", func() {
					decodedLRP := &DesiredLRP{}

					err := FromJSON([]byte(jsonBytes), decodedLRP)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(missingField))
				})
			})
		}

		for field, payload := range map[string]string{
			"annotation": `{
				"rootfs": "some-rootfs",
				"domain": "some-domain",
				"process_guid": "process_guid",
				"instances": 1,
				"action": {
					"download":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}
				},
				"annotation":"` + strings.Repeat("a", 10*1024+1) + `"
			}`,
			"routes": `{
				"rootfs": "some-rootfs",
				"domain": "some-domain",
				"process_guid": "process_guid",
				"instances": 1,
				"action": {
					"download":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}
				},
				"routes": {
					"cf-route": "` + strings.Repeat("r", 4*1024) + `"
				}
			}`,
		} {
			tooLongField := field
			jsonBytes := payload

			Context("when the json field "+tooLongField+" is too long", func() {
				It("returns an error indicating so", func() {
					decodedLRP := &DesiredLRP{}

					err := FromJSON([]byte(jsonBytes), decodedLRP)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tooLongField))
				})
			})
		}
	})
})
