package models_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/bbs/models"
)

var _ = Describe("Actions", func() {
	itSerializes := func(actionPayload string, a interface{}) {
		action := models.UnwrapAction(a)
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

	itDeserializes := func(actionPayload string, a *models.Action) {
		action := models.UnwrapAction(a)
		It("JSON -> Action for "+string(action.ActionType()), func() {
			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)

			By("unwrapping", func() {
				var unmarshalledAction models.ActionInterface
				unmarshalledAction, err := models.UnmarshalAction([]byte(wrappedJSON))
				Expect(err).NotTo(HaveOccurred())
				Expect(unmarshalledAction).To(Equal(action))
			})
		})
	}

	itSerializesAndDeserializes := func(actionPayload string, action *models.Action) {
		itSerializes(actionPayload, action)
		itDeserializes(actionPayload, action)
	}

	Describe("UnmarshalAction", func() {
		It("returns an error when the action is not registered", func() {
			_, err := models.UnmarshalAction([]byte(`{"bogusAction": {}}`))
			Expect(err).To(MatchError("Unknown action: bogusAction"))
		})
	})

	Describe("Download", func() {
		itSerializesAndDeserializes(
			`{
					"artifact": "mouse",
					"from": "web_location",
					"to": "local_location",
					"cache_key": "elephant",
					"user": "someone"
			}`,
			models.WrapAction(&models.DownloadAction{
				Artifact: "mouse",
				From:     "web_location",
				To:       "local_location",
				CacheKey: "elephant",
				User:     "someone",
			}),
		)

		Describe("Validate", func() {
			var downloadAction models.DownloadAction

			Context("when the action has 'from', 'to', and 'user' specified", func() {
				It("is valid", func() {
					downloadAction = models.DownloadAction{
						From: "web_location",
						To:   "local_location",
						User: "someone",
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
				{
					"user",
					models.DownloadAction{
						From: "web_location",
						To:   "local_location",
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
					"to": "web_location",
					"user": "someone"
			}`,
			models.WrapAction(&models.UploadAction{
				Artifact: "mouse",
				From:     "local_location",
				To:       "web_location",
				User:     "someone",
			}),
		)

		Describe("Validate", func() {
			var uploadAction models.UploadAction

			Context("when the action has 'from', 'to', and 'user' specified", func() {
				It("is valid", func() {
					uploadAction = models.UploadAction{
						To:   "web_location",
						From: "local_location",
						User: "someone",
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
				{
					"user",
					models.UploadAction{
						To:   "web_location",
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
					"user": "me",
					"path": "rm",
					"args": ["-rf", "/"],
					"dir": "./some-dir",
					"env": [
						{"name":"FOO", "value":"1"},
						{"name":"BAR", "value":"2"}
					],
					"resource_limits":{"nofile": 10}
			}`,
			models.WrapAction(&models.RunAction{
				User: "me",
				Path: "rm",
				Dir:  "./some-dir",
				Args: []string{"-rf", "/"},
				Env: []*models.EnvironmentVariable{
					{"FOO", "1"},
					{"BAR", "2"},
				},
				ResourceLimits: &models.ResourceLimits{Nofile: 10},
			}),
		)

		Describe("Validate", func() {
			var runAction models.RunAction

			Context("when the action has the required fields", func() {
				It("is valid", func() {
					runAction = models.RunAction{
						Path: "ls",
						User: "foo",
					}

					err := runAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"path",
					models.RunAction{
						User: "me",
					},
				},
				{
					"user",
					models.RunAction{
						Path: "ls",
					},
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
						"user": "someone",
						"resource_limits":{
							"nofile": 10	
						}
					}
				},
				"timeout": 10000000
			}`,
			models.Timeout(
				&models.RunAction{
					Path:           "echo",
					User:           "someone",
					ResourceLimits: &models.ResourceLimits{Nofile: 10},
				},
				10*time.Millisecond,
			),
		)

		itSerializesAndDeserializes(
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
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
						},
						Timeout: int64(time.Second),
					}

					err := timeoutAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					models.TimeoutAction{
						Timeout: int64(time.Second),
					},
				},
				{
					"from",
					models.TimeoutAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								To:   "web_location",
								User: "someone",
							},
						},
						Timeout: int64(time.Second),
					},
				},
				{
					"timeout",
					models.TimeoutAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
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
							"resource_limits":{},
							"user": "me"
						}
					}
			}`,
			models.Try(&models.Action{RunAction: &models.RunAction{
				Path:           "echo",
				User:           "me",
				ResourceLimits: &models.ResourceLimits{},
			}}),
		)

		itSerializesAndDeserializes(
			`{}`,
			models.Try(nil),
		)

		Describe("Validate", func() {
			var tryAction models.TryAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					tryAction = models.TryAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
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
						Action: &models.Action{
							UploadAction: &models.UploadAction{
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

	Describe("Parallel", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			models.Parallel(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				},
				&models.RunAction{
					Path:           "echo",
					User:           "me",
					ResourceLimits: &models.ResourceLimits{Nofile: 0},
				},
			),
		)

		itSerializesAndDeserializes(
			`{}`,
			models.WrapAction(&models.ParallelAction{}),
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			models.Parallel(nil),
		)

		Describe("Validate", func() {
			var parallelAction models.ParallelAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					parallelAction = models.ParallelAction{
						Actions: []*models.Action{
							&models.Action{
								UploadAction: &models.UploadAction{
									From: "local_location",
									To:   "web_location",
									User: "someone",
								},
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
						Actions: []*models.Action{nil},
					},
				},
				{
					"from",
					models.ParallelAction{
						Actions: []*models.Action{
							&models.Action{
								UploadAction: &models.UploadAction{
									To: "web_location",
								},
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
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			models.Serial(
				&models.Action{
					DownloadAction: &models.DownloadAction{
						From:     "web_location",
						To:       "local_location",
						CacheKey: "elephant",
						User:     "someone",
					},
				},
				&models.Action{
					RunAction: &models.RunAction{
						Path:           "echo",
						User:           "me",
						ResourceLimits: &models.ResourceLimits{},
					},
				},
			),
		)

		itSerializesAndDeserializes(
			`{}`,
			models.WrapAction(&models.SerialAction{}),
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			models.Serial(nil),
		)

		Describe("Validate", func() {
			var serialAction models.SerialAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					serialAction = *models.Serial(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
					).SerialAction
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
						Actions: []*models.Action{nil},
					},
				},
				{
					"from",
					*models.Serial(
						&models.UploadAction{
							To: "web_location",
						},
					).SerialAction,
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
							"resource_limits":{},
							"user": "me"
						}
					}
			}`,
			models.EmitProgressFor(
				&models.RunAction{
					Path:           "echo",
					User:           "me",
					ResourceLimits: &models.ResourceLimits{},
				},
				"reticulating splines", "reticulated splines", "reticulation failed",
			),
		)

		itSerializesAndDeserializes(
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
					emitProgressAction = *models.EmitProgressFor(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
						"", "", "",
					).EmitProgressAction

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
					*models.EmitProgressFor(
						&models.UploadAction{
							To: "web_location",
						},
						"", "", "",
					).EmitProgressAction,
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Codependent", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			models.Codependent(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				},
				&models.RunAction{Path: "echo", User: "me", ResourceLimits: &models.ResourceLimits{}},
			),
		)

		itSerializesAndDeserializes(
			`{}`,
			models.WrapAction(&models.CodependentAction{}),
		)

		itSerializesAndDeserializes(
			`{
				"actions": [null]
			}`,
			models.Codependent(nil),
		)

		Describe("Validate", func() {
			var codependentAction models.CodependentAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					codependentAction = *models.Codependent(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
					).CodependentAction

					err := codependentAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					models.CodependentAction{},
				},
				{
					"action at index 0",
					*models.Codependent(nil).CodependentAction,
				},
				{
					"from",
					*models.Codependent(&models.UploadAction{
						To: "web_location",
					}).CodependentAction,
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})
})
