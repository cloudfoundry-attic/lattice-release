package serialization_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP Serialization", func() {
	var routes map[string]*json.RawMessage
	var routingInfo receptor.RoutingInfo

	BeforeEach(func() {
		raw := json.RawMessage([]byte(`[{"port":1,"hostnames":["route-1","route-2"]}]`))
		routes = map[string]*json.RawMessage{
			"cf-router": &raw,
		}
		routingInfo = receptor.RoutingInfo(routes)
	})

	Describe("DesiredLRPFromRequest", func() {
		var request receptor.DesiredLRPCreateRequest
		var desiredLRP models.DesiredLRP
		var securityRule models.SecurityGroupRule

		BeforeEach(func() {
			securityRule = models.SecurityGroupRule{
				Protocol:     "tcp",
				Destinations: []string{"0.0.0.0/0"},
				PortRange: &models.PortRange{
					Start: 1,
					End:   1024,
				},
			}
			request = receptor.DesiredLRPCreateRequest{
				ProcessGuid: "the-process-guid",
				Domain:      "the-domain",
				RootFS:      "the-rootfs",
				Annotation:  "foo",
				Instances:   1,
				Ports:       []uint16{2345, 6789},
				Action: &models.RunAction{
					Path: "the-path",
				},
				StartTimeout: 4,
				Privileged:   true,
				LogGuid:      "log-guid-0",
				LogSource:    "log-source-name-0",
				MetricsGuid:  "metrics-guid-0",
				EgressRules: []models.SecurityGroupRule{
					securityRule,
				},
				Routes: routingInfo,
			}
		})
		JustBeforeEach(func() {
			desiredLRP = serialization.DesiredLRPFromRequest(request)
		})

		It("translates the request into a DesiredLRP model, preserving attributes", func() {
			Ω(desiredLRP.ProcessGuid).Should(Equal("the-process-guid"))
			Ω(desiredLRP.Domain).Should(Equal("the-domain"))
			Ω(desiredLRP.RootFS).Should(Equal("the-rootfs"))
			Ω(desiredLRP.Annotation).Should(Equal("foo"))
			Ω(desiredLRP.StartTimeout).Should(Equal(uint(4)))
			Ω(desiredLRP.Ports).Should(HaveLen(2))
			Ω(desiredLRP.Ports[0]).Should(Equal(uint16(2345)))
			Ω(desiredLRP.Ports[1]).Should(Equal(uint16(6789)))
			Ω(desiredLRP.Privileged).Should(BeTrue())
			Ω(desiredLRP.EgressRules).Should(HaveLen(1))
			Ω(desiredLRP.EgressRules[0].Protocol).Should(Equal(securityRule.Protocol))
			Ω(desiredLRP.EgressRules[0].PortRange).Should(Equal(securityRule.PortRange))
			Ω(desiredLRP.EgressRules[0].Destinations).Should(Equal(securityRule.Destinations))
			Ω(desiredLRP.Routes).Should(HaveLen(1))
			Ω(desiredLRP.LogGuid).Should(Equal("log-guid-0"))
			Ω(desiredLRP.LogSource).Should(Equal("log-source-name-0"))
			Ω(desiredLRP.MetricsGuid).Should(Equal("metrics-guid-0"))
			Ω([]byte(*desiredLRP.Routes["cf-router"])).Should(MatchJSON(`[{"port": 1,"hostnames": ["route-1", "route-2"]}]`))
		})
	})

	Describe("DesiredLRPToResponse", func() {
		var desiredLRP models.DesiredLRP
		var securityRule models.SecurityGroupRule

		BeforeEach(func() {
			securityRule = models.SecurityGroupRule{
				Protocol:     "tcp",
				Destinations: []string{"0.0.0.0/0"},
				Ports:        []uint16{80, 443},
				Log:          true,
			}

			desiredLRP = models.DesiredLRP{
				ProcessGuid: "process-guid-0",
				Domain:      "domain-0",
				RootFS:      "root-fs-0",
				Instances:   127,
				EnvironmentVariables: []models.EnvironmentVariable{
					{Name: "ENV_VAR_NAME", Value: "value"},
				},
				Action:       &models.RunAction{Path: "/bin/true"},
				StartTimeout: 4,
				DiskMB:       126,
				MemoryMB:     1234,
				CPUWeight:    192,
				Privileged:   true,
				Ports: []uint16{
					456,
				},
				Routes:      routes,
				LogGuid:     "log-guid-0",
				LogSource:   "log-source-name-0",
				MetricsGuid: "metrics-guid-0",
				Annotation:  "annotation-0",
				EgressRules: []models.SecurityGroupRule{
					securityRule,
				},
				ModificationTag: models.ModificationTag{
					Epoch: "some-epoch",
					Index: 50,
				},
			}
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.DesiredLRPResponse{
				ProcessGuid: "process-guid-0",
				Domain:      "domain-0",
				RootFS:      "root-fs-0",
				Instances:   127,
				EnvironmentVariables: []receptor.EnvironmentVariable{
					{Name: "ENV_VAR_NAME", Value: "value"},
				},
				Action:       &models.RunAction{Path: "/bin/true"},
				StartTimeout: 4,
				DiskMB:       126,
				MemoryMB:     1234,
				CPUWeight:    192,
				Privileged:   true,
				Ports: []uint16{
					456,
				},
				Routes:      routingInfo,
				LogGuid:     "log-guid-0",
				LogSource:   "log-source-name-0",
				MetricsGuid: "metrics-guid-0",
				Annotation:  "annotation-0",
				EgressRules: []models.SecurityGroupRule{
					securityRule,
				},
				ModificationTag: receptor.ModificationTag{
					Epoch: "some-epoch",
					Index: 50,
				},
			}

			actualResponse := serialization.DesiredLRPToResponse(desiredLRP)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})
	})
})
