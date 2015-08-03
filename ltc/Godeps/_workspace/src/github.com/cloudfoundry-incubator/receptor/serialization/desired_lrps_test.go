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
					User: "me",
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
			Expect(desiredLRP.ProcessGuid).To(Equal("the-process-guid"))
			Expect(desiredLRP.Domain).To(Equal("the-domain"))
			Expect(desiredLRP.RootFS).To(Equal("the-rootfs"))
			Expect(desiredLRP.Annotation).To(Equal("foo"))
			Expect(desiredLRP.StartTimeout).To(Equal(uint(4)))
			Expect(desiredLRP.Ports).To(HaveLen(2))
			Expect(desiredLRP.Ports[0]).To(Equal(uint16(2345)))
			Expect(desiredLRP.Ports[1]).To(Equal(uint16(6789)))
			Expect(desiredLRP.Privileged).To(BeTrue())
			Expect(desiredLRP.EgressRules).To(HaveLen(1))
			Expect(desiredLRP.EgressRules[0].Protocol).To(Equal(securityRule.Protocol))
			Expect(desiredLRP.EgressRules[0].PortRange).To(Equal(securityRule.PortRange))
			Expect(desiredLRP.EgressRules[0].Destinations).To(Equal(securityRule.Destinations))
			Expect(desiredLRP.Routes).To(HaveLen(1))
			Expect(desiredLRP.LogGuid).To(Equal("log-guid-0"))
			Expect(desiredLRP.LogSource).To(Equal("log-source-name-0"))
			Expect(desiredLRP.MetricsGuid).To(Equal("metrics-guid-0"))
			Expect([]byte(*desiredLRP.Routes["cf-router"])).To(MatchJSON(`[{"port": 1,"hostnames": ["route-1", "route-2"]}]`))
		})
	})

})
