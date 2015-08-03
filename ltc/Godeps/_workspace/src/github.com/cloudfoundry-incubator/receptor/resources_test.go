package receptor_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {
	Describe("TaskCreateRequest", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedRequest receptor.TaskCreateRequest
				var payload string

				BeforeEach(func() {
					expectedRequest = receptor.TaskCreateRequest{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "action": null
            }`
					})

					It("unmarshals", func() {
						var actualRequest receptor.TaskCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualRequest).To(Equal(expectedRequest))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualRequest receptor.TaskCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualRequest).To(Equal(expectedRequest))
					})
				})
			})
			Context("with security group rules", func() {
				var expectedRequest receptor.TaskCreateRequest
				var payload string
				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedRequest = receptor.TaskCreateRequest{
						EgressRules: []oldmodels.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								PortRange: &oldmodels.PortRange{
									Start: 1,
									End:   1024,
								},
							},
						},
					}
				})

				It("unmarshals", func() {
					var actualRequest receptor.TaskCreateRequest
					err := json.Unmarshal([]byte(payload), &actualRequest)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualRequest).To(Equal(expectedRequest))
				})
			})
		})
	})

	Describe("TaskResponse", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedResponse receptor.TaskResponse
				var payload string

				BeforeEach(func() {
					expectedResponse = receptor.TaskResponse{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "action": null
            }`
					})

					It("unmarshals", func() {
						var actualResponse receptor.TaskResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualResponse).To(Equal(expectedResponse))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualResponse receptor.TaskResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualResponse).To(Equal(expectedResponse))
					})
				})
			})
			Context("with security group rules", func() {
				var expectedResponse receptor.TaskResponse
				var payload string
				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedResponse = receptor.TaskResponse{
						EgressRules: []oldmodels.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								PortRange: &oldmodels.PortRange{
									Start: 1,
									End:   1024,
								},
							},
						},
					}
				})

				It("unmarshals", func() {
					var actualRequest receptor.TaskResponse
					err := json.Unmarshal([]byte(payload), &actualRequest)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualRequest).To(Equal(expectedResponse))
				})
			})
		})
	})

	Describe("DesiredLRPCreateRequest", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedRequest receptor.DesiredLRPCreateRequest
				var payload string

				BeforeEach(func() {
					expectedRequest = receptor.DesiredLRPCreateRequest{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "setup": null,
              "action": null,
              "monitor": null
            }`
					})

					It("unmarshals", func() {
						var actualRequest receptor.DesiredLRPCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualRequest).To(Equal(expectedRequest))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualRequest receptor.DesiredLRPCreateRequest
						err := json.Unmarshal([]byte(payload), &actualRequest)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualRequest).To(Equal(expectedRequest))
					})
				})
			})
			Context("with security group rules", func() {
				var expectedRequest receptor.DesiredLRPCreateRequest
				var payload string

				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "ports": [80, 443],
				        "log": true
			        }
		        ]
					}`
					expectedRequest = receptor.DesiredLRPCreateRequest{
						EgressRules: []oldmodels.SecurityGroupRule{
							{
								Protocol:     "tcp",
								Destinations: []string{"0.0.0.0/0"},
								Ports:        []uint16{80, 443},
								Log:          true,
							},
						},
					}
				})

				It("unmarshals", func() {
					var actualRequest receptor.DesiredLRPCreateRequest
					err := json.Unmarshal([]byte(payload), &actualRequest)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualRequest).To(Equal(expectedRequest))
				})
			})
		})
	})

	Describe("DesiredLRPResponse", func() {
		Describe("Unmarshalling", func() {
			Context("with invalid actions", func() {
				var expectedResponse receptor.DesiredLRPResponse
				var payload string

				BeforeEach(func() {
					expectedResponse = receptor.DesiredLRPResponse{}
				})

				Context("with null action", func() {
					BeforeEach(func() {
						payload = `{
              "setup": null,
              "action": null,
              "monitor": null
            }`
					})

					It("unmarshals", func() {
						var actualResponse receptor.DesiredLRPResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualResponse).To(Equal(expectedResponse))
					})
				})

				Context("with missing action", func() {
					BeforeEach(func() {
						payload = `{}`
					})

					It("unmarshals", func() {
						var actualResponse receptor.DesiredLRPResponse
						err := json.Unmarshal([]byte(payload), &actualResponse)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualResponse).To(Equal(expectedResponse))
					})
				})
			})
			Context("with security group rules", func() {
				var expectedResponse receptor.DesiredLRPResponse
				var payload string

				BeforeEach(func() {
					payload = `{
						"egress_rules":[
		          {
				        "protocol": "tcp",
								"destinations": ["0.0.0.0/0"],
				        "port_range": {
					        "start": 1,
					        "end": 1024
				        }
			        }
		        ]
					}`
					expectedResponse = receptor.DesiredLRPResponse{
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
					}
				})

				It("unmarshals", func() {
					var actualResponse receptor.DesiredLRPResponse
					err := json.Unmarshal([]byte(payload), &actualResponse)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualResponse).To(Equal(expectedResponse))
				})
			})
		})
	})

	Describe("RoutingInfo", func() {
		const jsonRoutes = `{
			"cf-router": [{ "port": 1, "hostnames": ["a", "b"]}],
			"foo" : "bar"
		}`

		var routingInfo receptor.RoutingInfo

		BeforeEach(func() {
			routingInfo = receptor.RoutingInfo{}
		})

		Describe("MarshalJson", func() {
			It("marshals routes when present", func() {
				routingInfo := receptor.RoutingInfo{}

				bytes, err := json.Marshal(cfroutes.CFRoutes{
					{Hostnames: []string{"a", "b"}, Port: 1},
				})
				Expect(err).NotTo(HaveOccurred())

				cfRawMessage := json.RawMessage(bytes)
				routingInfo[cfroutes.CF_ROUTER] = &cfRawMessage

				fooRawMessage := json.RawMessage([]byte(`"bar"`))
				routingInfo["foo"] = &fooRawMessage

				bytes, err = json.Marshal(routingInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(bytes).To(MatchJSON(jsonRoutes))
			})

			Context("when the routing info is empty", func() {
				It("serialization can omit it", func() {
					type wrapper struct {
						Name   string               `json:"name"`
						Routes receptor.RoutingInfo `json:"routes,omitempty"`
					}

					w := wrapper{
						Name:   "wrapper",
						Routes: receptor.RoutingInfo{},
					}

					Expect(w.Routes).NotTo(BeNil())
					Expect(w.Routes).To(BeEmpty())

					bytes, err := json.Marshal(w)
					Expect(err).NotTo(HaveOccurred())

					Expect(bytes).To(MatchJSON(`{"name":"wrapper"}`))
				})
			})
		})

		Describe("Unmarshal", func() {
			It("returns both cf-routes and other", func() {
				err := json.Unmarshal([]byte(jsonRoutes), &routingInfo)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(routingInfo)).To(Equal(2))
				Expect(string(*routingInfo["cf-router"])).To(MatchJSON(`[{"port": 1,"hostnames":["a", "b"]}]`))
				Expect(string(*routingInfo["foo"])).To(MatchJSON(`"bar"`))
			})
		})
	})

	Describe("ModificationTag", func() {
		currentTag := receptor.ModificationTag{Epoch: "abc", Index: 1}
		differentEpochCurrentIndexTag := receptor.ModificationTag{Epoch: "def", Index: 1}
		differentEpochNewerIndexTag := receptor.ModificationTag{Epoch: "def", Index: 2}
		differentEpochOlderIndexTag := receptor.ModificationTag{Epoch: "def", Index: 0}
		missingEpochTag := receptor.ModificationTag{Epoch: "", Index: 0}
		sameEpochNewerIndexTag := receptor.ModificationTag{Epoch: "abc", Index: 2}
		sameEpochOlderIndexTag := receptor.ModificationTag{Epoch: "abc", Index: 0}

		Describe("Equal", func() {
			It("returns true when equivalent", func() {
				Expect(currentTag.Equal(currentTag)).To(BeTrue())
			})

			It("returns false when the indexes do not match", func() {
				Expect(currentTag.Equal(sameEpochNewerIndexTag)).To(BeFalse())
				Expect(currentTag.Equal(sameEpochOlderIndexTag)).To(BeFalse())
			})

			It("returns false when the epochs do not match", func() {
				Expect(currentTag.Equal(differentEpochCurrentIndexTag)).To(BeFalse())
			})

			It("returns false when epoch is empty", func() {
				Expect(currentTag.Equal(missingEpochTag)).To(BeFalse())
			})
		})

		Describe("SucceededBy", func() {
			It("returns true when the modification tag has an empty epoch", func() {
				Expect(missingEpochTag.SucceededBy(currentTag)).To(BeTrue())
			})

			It("returns true when the other modification tag has an empty epoch", func() {
				Expect(currentTag.SucceededBy(missingEpochTag)).To(BeTrue())
			})

			It("returns true when the epoch is different and the index is older", func() {
				Expect(currentTag.SucceededBy(differentEpochOlderIndexTag)).To(BeTrue())
			})

			It("returns true when the epoch is different and the index is the same", func() {
				Expect(currentTag.SucceededBy(differentEpochCurrentIndexTag)).To(BeTrue())
			})

			It("returns true when the epoch is different and the index is newer", func() {
				Expect(currentTag.SucceededBy(differentEpochNewerIndexTag)).To(BeTrue())
			})

			It("returns false when the epoch the same and the index is older", func() {
				Expect(currentTag.SucceededBy(sameEpochOlderIndexTag)).To(BeFalse())
			})

			It("returns false when the epoch the same and the index is the same", func() {
				Expect(currentTag.SucceededBy(currentTag)).To(BeFalse())
			})

			It("returns true when the epoch the same and the index is newer", func() {
				Expect(currentTag.SucceededBy(sameEpochNewerIndexTag)).To(BeTrue())
			})
		})
	})
})
