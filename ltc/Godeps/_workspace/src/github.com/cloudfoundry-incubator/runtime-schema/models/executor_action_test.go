package models_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("Actions", func() {
	itSerializes := func(actionPayload string, action Action) {
		It("Action -> JSON for "+string(action.ActionType()), func() {
			By("marshalling to JSON", func() {
				marshalledAction := action

				json, err := json.Marshal(&marshalledAction)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(json).Should(MatchJSON(actionPayload))
			})

			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)
			By("wrapping", func() {
				marshalledAction := action

				json, err := MarshalAction(marshalledAction)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(json).Should(MatchJSON(wrappedJSON))
			})
		})
	}

	itDeserializes := func(actionPayload string, action Action) {
		It("JSON -> Action for "+string(action.ActionType()), func() {
			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)

			By("unwrapping", func() {
				var unmarshalledAction Action
				unmarshalledAction, err := UnmarshalAction([]byte(wrappedJSON))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(unmarshalledAction).Should(Equal(action))
			})
		})
	}

	itSerializesAndDeserializes := func(actionPayload string, action Action) {
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
			&DownloadAction{
				Artifact: "mouse",
				From:     "web_location",
				To:       "local_location",
				CacheKey: "elephant",
			},
		)

		Describe("Validate", func() {
			var downloadAction DownloadAction

			Context("when the action has 'from' and 'to' specified", func() {
				It("is valid", func() {
					downloadAction = DownloadAction{
						From: "web_location",
						To:   "local_location",
					}

					err := downloadAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"from",
					DownloadAction{
						To: "local_location",
					},
				},
				{
					"to",
					DownloadAction{
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
			&UploadAction{
				Artifact: "mouse",
				From:     "local_location",
				To:       "web_location",
			},
		)

		Describe("Validate", func() {
			var uploadAction UploadAction

			Context("when the action has 'from' and 'to' specified", func() {
				It("is valid", func() {
					uploadAction = UploadAction{
						To:   "web_location",
						From: "local_location",
					}

					err := uploadAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"from",
					UploadAction{
						To: "web_location",
					},
				},
				{
					"to",
					UploadAction{
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
			&RunAction{
				Path: "rm",
				Dir:  "./some-dir",
				Args: []string{"-rf", "/"},
				Env: []EnvironmentVariable{
					{"FOO", "1"},
					{"BAR", "2"},
				},
				Privileged: true,
			},
		)

		Describe("Validate", func() {
			var runAction RunAction

			Context("when the action has 'path' specified", func() {
				It("is valid", func() {
					runAction = RunAction{
						Path: "ls",
					}

					err := runAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"path",
					RunAction{},
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
			Timeout(
				&RunAction{
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
			Timeout(
				nil,
				10*time.Millisecond,
			),
		)

		itDeserializes(
			`{
				"timeout": 10000000
			}`,
			Timeout(
				nil,
				10*time.Millisecond,
			),
		)

		Describe("Validate", func() {
			var timeoutAction TimeoutAction

			Context("when the action has 'action' specified and a positive timeout", func() {
				It("is valid", func() {
					timeoutAction = TimeoutAction{
						Action: &UploadAction{
							From: "local_location",
							To:   "web_location",
						},
						Timeout: time.Second,
					}

					err := timeoutAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					TimeoutAction{
						Timeout: time.Second,
					},
				},
				{
					"from",
					TimeoutAction{
						Action: &UploadAction{
							To: "web_location",
						},
						Timeout: time.Second,
					},
				},
				{
					"timeout",
					TimeoutAction{
						Action: &UploadAction{
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
			Try(&RunAction{Path: "echo"}),
		)

		itSerializesAndDeserializes(
			`{
					"action": null
			}`,
			Try(nil),
		)

		itDeserializes(
			`{}`,
			Try(nil),
		)

		Describe("Validate", func() {
			var tryAction TryAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					tryAction = TryAction{
						Action: &UploadAction{
							From: "local_location",
							To:   "web_location",
						},
					}

					err := tryAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					TryAction{},
				},
				{
					"from",
					TryAction{
						Action: &UploadAction{
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
			Parallel(
				&DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
				},
				&RunAction{Path: "echo"},
			),
		)

		itDeserializes(
			`{}`,
			&ParallelAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": null
			}`,
			&ParallelAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": []
			}`,
			&ParallelAction{
				Actions: []Action{},
			},
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			&ParallelAction{
				Actions: []Action{
					nil,
				},
			},
		)

		Describe("Validate", func() {
			var parallelAction ParallelAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					parallelAction = ParallelAction{
						Actions: []Action{
							&UploadAction{
								From: "local_location",
								To:   "web_location",
							},
						},
					}

					err := parallelAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					ParallelAction{},
				},
				{
					"action at index 0",
					ParallelAction{
						Actions: []Action{
							nil,
						},
					},
				},
				{
					"from",
					ParallelAction{
						Actions: []Action{
							&UploadAction{
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
			Serial(
				&DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
				},
				&RunAction{Path: "echo"},
			),
		)

		itDeserializes(
			`{}`,
			&SerialAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": null
			}`,
			&SerialAction{},
		)

		itSerializesAndDeserializes(
			`{
				"actions": []
			}`,
			&SerialAction{
				Actions: []Action{},
			},
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			&SerialAction{
				Actions: []Action{
					nil,
				},
			},
		)

		Describe("Validate", func() {
			var serialAction SerialAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					serialAction = SerialAction{
						Actions: []Action{
							&UploadAction{
								From: "local_location",
								To:   "web_location",
							},
						},
					}

					err := serialAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					SerialAction{},
				},
				{
					"action at index 0",
					SerialAction{
						Actions: []Action{
							nil,
						},
					},
				},
				{
					"from",
					SerialAction{
						Actions: []Action{
							&UploadAction{
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
			EmitProgressFor(
				&RunAction{
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
			EmitProgressFor(
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
			EmitProgressFor(
				nil,
				"reticulating splines", "reticulated splines", "reticulation failed",
			),
		)

		Describe("Validate", func() {
			var emitProgressAction EmitProgressAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					emitProgressAction = EmitProgressAction{
						Action: &UploadAction{
							From: "local_location",
							To:   "web_location",
						},
					}

					err := emitProgressAction.Validate()
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					EmitProgressAction{},
				},
				{
					"from",
					EmitProgressAction{
						Action: &UploadAction{
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
