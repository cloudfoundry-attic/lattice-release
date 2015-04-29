package models_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("Actions", func() {
	itSerializes := func(actionPayload string, action models.Action) {
		It("Action -> JSON for "+string(action.ActionType()), func() {
			By("marshalling to JSON", func() {
				marshalledAction := action

				json, err := json.Marshal(&marshalledAction)
				Expect(err).NotTo(HaveOccurred())
				Expect(json).To(MatchJSON(actionPayload))
			})

			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)
			By("wrapping", func() {
				marshalledAction := action

				json, err := models.MarshalAction(marshalledAction)
				Expect(err).NotTo(HaveOccurred())
				Expect(json).To(MatchJSON(wrappedJSON))
			})
		})
	}

	itDeserializes := func(actionPayload string, action models.Action) {
		It("JSON -> Action for "+string(action.ActionType()), func() {
			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)

			By("unwrapping", func() {
				var unmarshalledAction models.Action
				unmarshalledAction, err := models.UnmarshalAction([]byte(wrappedJSON))
				Expect(err).NotTo(HaveOccurred())
				Expect(unmarshalledAction).To(Equal(action))
			})
		})
	}

	itSerializesAndDeserializes := func(actionPayload string, action models.Action) {
		itSerializes(actionPayload, action)
		itDeserializes(actionPayload, action)
	}

	Describe("Download", func() {
		itSerializesAndDeserializes(
			`{
					"artifact": "mouse",
					"from": "web_location",
					"to": "local_location",
					"cache_key": "elephant"
			}`,
			&models.DownloadAction{
				Artifact: "mouse",
				From:     "web_location",
				To:       "local_location",
				CacheKey: "elephant",
			},
		)

		Describe("Validate", func() {
			var downloadAction models.DownloadAction

			Context("when the action has 'from' and 'to' specified", func() {
				It("is valid", func() {
					downloadAction = models.DownloadAction{
						From: "web_location",
						To:   "local_location",
					}

					err := downloadAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"from",
					models.DownloadAction{
						To: "local_location",
					},
				},
				{
					"to",
					models.DownloadAction{
						From: "web_location",
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Upload", func() {
		itSerializesAndDeserializes(
			`{
					"artifact": "mouse",
					"from": "local_location",
					"to": "web_location"
			}`,
			&models.UploadAction{
				Artifact: "mouse",
				From:     "local_location",
				To:       "web_location",
			},
		)

		Describe("Validate", func() {
			var uploadAction models.UploadAction

			Context("when the action has 'from' and 'to' specified", func() {
				It("is valid", func() {
					uploadAction = models.UploadAction{
						To:   "web_location",
						From: "local_location",
					}

					err := uploadAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"from",
					models.UploadAction{
						To: "web_location",
					},
				},
				{
					"to",
					models.UploadAction{
						From: "local_location",
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Run", func() {
		itSerializesAndDeserializes(
			`{
					"path": "rm",
					"args": ["-rf", "/"],
					"dir": "./some-dir",
					"env": [
						{"name":"FOO", "value":"1"},
						{"name":"BAR", "value":"2"}
					],
					"resource_limits":{},
					"privileged": true
			}`,
			&models.RunAction{
				Path: "rm",
				Dir:  "./some-dir",
				Args: []string{"-rf", "/"},
				Env: []models.EnvironmentVariable{
					{"FOO", "1"},
					{"BAR", "2"},
				},
				Privileged: true,
			},
		)

		Describe("Validate", func() {
			var runAction models.RunAction

			Context("when the action has 'path' specified", func() {
				It("is valid", func() {
					runAction = models.RunAction{
						Path: "ls",
					}

					err := runAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"path",
					models.RunAction{},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Timeout", func() {
		itSerializesAndDeserializes(
			`{
				"action": {
					"run": {
						"path": "echo",
						"args": null,
						"env": null,
						"resource_limits":{}
					}
				},
				"timeout": 10000000
			}`,
			models.Timeout(
				&models.RunAction{
					Path: "echo",
				},
				10*time.Millisecond,
			),
		)

		itSerializesAndDeserializes(
			`{
				"action": null,
				"timeout": 10000000
			}`,
			models.Timeout(
				nil,
				10*time.Millisecond,
			),
		)

		itDeserializes(
			`{
				"timeout": 10000000
			}`,
			models.Timeout(
				nil,
				10*time.Millisecond,
			),
		)

		Describe("Validate", func() {
			var timeoutAction models.TimeoutAction

			Context("when the action has 'action' specified and a positive timeout", func() {
				It("is valid", func() {
					timeoutAction = models.TimeoutAction{
						Action: &models.UploadAction{
							From: "local_location",
							To:   "web_location",
						},
						Timeout: time.Second,
					}

					err := timeoutAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					models.TimeoutAction{
						Timeout: time.Second,
					},
				},
				{
					"from",
					models.TimeoutAction{
						Action: &models.UploadAction{
							To: "web_location",
						},
						Timeout: time.Second,
					},
				},
				{
					"timeout",
					models.TimeoutAction{
						Action: &models.UploadAction{
							From: "local_location",
							To:   "web_location",
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Try", func() {
		itSerializesAndDeserializes(
			`{
					"action": {
						"run": {
							"path": "echo",
							"args": null,
							"env": null,
							"resource_limits":{}
						}
					}
			}`,
			models.Try(&models.RunAction{Path: "echo"}),
		)

		itSerializesAndDeserializes(
			`{
					"action": null
			}`,
			models.Try(nil),
		)

		itDeserializes(
			`{}`,
			models.Try(nil),
		)

		Describe("Validate", func() {
			var tryAction models.TryAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					tryAction = models.TryAction{
						Action: &models.UploadAction{
							From: "local_location",
							To:   "web_location",
						},
					}

					err := tryAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					models.TryAction{},
				},
				{
					"from",
					models.TryAction{
						Action: &models.UploadAction{
							To: "web_location",
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Parallel", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"env": null,
								"path": "echo",
								"args": null
							}
						}
					]
			}`,
			models.Parallel(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
				},
				&models.RunAction{Path: "echo"},
			),
		)

		itDeserializes(
			`{}`,
			&models.ParallelAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": null
			}`,
			&models.ParallelAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": []
			}`,
			&models.ParallelAction{
				Actions: []models.Action{},
			},
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			&models.ParallelAction{
				Actions: []models.Action{
					nil,
				},
			},
		)

		Describe("Validate", func() {
			var parallelAction models.ParallelAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					parallelAction = models.ParallelAction{
						Actions: []models.Action{
							&models.UploadAction{
								From: "local_location",
								To:   "web_location",
							},
						},
					}

					err := parallelAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					models.ParallelAction{},
				},
				{
					"action at index 0",
					models.ParallelAction{
						Actions: []models.Action{
							nil,
						},
					},
				},
				{
					"from",
					models.ParallelAction{
						Actions: []models.Action{
							&models.UploadAction{
								To: "web_location",
							},
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Serial", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"env": null,
								"path": "echo",
								"args": null
							}
						}
					]
			}`,
			models.Serial(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
				},
				&models.RunAction{Path: "echo"},
			),
		)

		itDeserializes(
			`{}`,
			&models.SerialAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": null
			}`,
			&models.SerialAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": []
			}`,
			&models.SerialAction{
				Actions: []models.Action{},
			},
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			&models.SerialAction{
				Actions: []models.Action{
					nil,
				},
			},
		)

		Describe("Validate", func() {
			var serialAction models.SerialAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					serialAction = models.SerialAction{
						Actions: []models.Action{
							&models.UploadAction{
								From: "local_location",
								To:   "web_location",
							},
						},
					}

					err := serialAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					models.SerialAction{},
				},
				{
					"action at index 0",
					models.SerialAction{
						Actions: []models.Action{
							nil,
						},
					},
				},
				{
					"from",
					models.SerialAction{
						Actions: []models.Action{
							&models.UploadAction{
								To: "web_location",
							},
							nil,
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("EmitProgressAction", func() {
		itSerializesAndDeserializes(
			`{
					"start_message": "reticulating splines",
					"success_message": "reticulated splines",
					"failure_message_prefix": "reticulation failed",
					"action": {
						"run": {
							"path": "echo",
							"args": null,
							"env": null,
							"resource_limits":{}
						}
					}
			}`,
			models.EmitProgressFor(
				&models.RunAction{
					Path: "echo",
				},
				"reticulating splines", "reticulated splines", "reticulation failed",
			),
		)

		itSerializesAndDeserializes(
			`{
					"start_message": "reticulating splines",
					"success_message": "reticulated splines",
					"failure_message_prefix": "reticulation failed",
					"action": null
			}`,
			models.EmitProgressFor(
				nil,
				"reticulating splines", "reticulated splines", "reticulation failed",
			),
		)

		itDeserializes(
			`{
					"start_message": "reticulating splines",
					"success_message": "reticulated splines",
					"failure_message_prefix": "reticulation failed"
			}`,
			models.EmitProgressFor(
				nil,
				"reticulating splines", "reticulated splines", "reticulation failed",
			),
		)

		Describe("Validate", func() {
			var emitProgressAction models.EmitProgressAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					emitProgressAction = models.EmitProgressAction{
						Action: &models.UploadAction{
							From: "local_location",
							To:   "web_location",
						},
					}

					err := emitProgressAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					models.EmitProgressAction{},
				},
				{
					"from",
					models.EmitProgressAction{
						Action: &models.UploadAction{
							To: "web_location",
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})
})
