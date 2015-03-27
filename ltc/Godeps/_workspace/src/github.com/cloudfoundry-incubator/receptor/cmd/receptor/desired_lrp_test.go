package main_test

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Desired LRP API", func() {

	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("POST /v1/desired_lrps/", func() {
		var lrpToCreate receptor.DesiredLRPCreateRequest
		var createErr error

		BeforeEach(func() {
			lrpToCreate = newValidDesiredLRPCreateRequest()
			createErr = client.CreateDesiredLRP(lrpToCreate)
		})

		It("responds without an error", func() {
			Ω(createErr).ShouldNot(HaveOccurred())
		})

		It("desires an LRP in the BBS", func() {
			Eventually(bbs.DesiredLRPs).Should(HaveLen(1))
			desiredLRPs, err := bbs.DesiredLRPs()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(desiredLRPs[0].ProcessGuid).To(Equal(lrpToCreate.ProcessGuid))
		})

		Context("when the desired LRP already exists", func() {
			It("fails the request with an appropriate error", func() {
				err := client.CreateDesiredLRP(lrpToCreate)
				Ω(err).Should(BeAssignableToTypeOf(receptor.Error{}))
				Ω(err.(receptor.Error).Type).Should(Equal(receptor.DesiredLRPAlreadyExists))
			})
		})
	})

	Describe("GET /v1/desired_lrps/:process_guid", func() {
		var lrpRequest receptor.DesiredLRPCreateRequest
		var lrpResponse receptor.DesiredLRPResponse
		var getErr error
		var desiredLRP models.DesiredLRP

		BeforeEach(func() {
			lrpRequest = newValidDesiredLRPCreateRequest()
			err := client.CreateDesiredLRP(lrpRequest)
			Ω(err).ShouldNot(HaveOccurred())

			desiredLRP, err = bbs.DesiredLRPByProcessGuid(lrpRequest.ProcessGuid)
			Ω(err).ShouldNot(HaveOccurred())

			lrpResponse, getErr = client.GetDesiredLRP(lrpRequest.ProcessGuid)
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches the desired lrp with the matching process guid", func() {
			expectedLRPResponse := serialization.DesiredLRPToResponse(desiredLRP)
			Ω(lrpResponse).Should(Equal(expectedLRPResponse))
		})
	})

	Describe("PUT /v1/desired_lrps/:process_guid", func() {
		var updateErr error

		instances := 6
		annotation := "update-annotation"
		rawMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["updated-route"]}]`))

		routingInfo := receptor.RoutingInfo{
			"cf-router": &rawMessage,
		}

		BeforeEach(func() {
			createLRPReq := newValidDesiredLRPCreateRequest()
			err := client.CreateDesiredLRP(createLRPReq)
			Ω(err).ShouldNot(HaveOccurred())

			update := receptor.DesiredLRPUpdateRequest{
				Instances:  &instances,
				Annotation: &annotation,
				Routes:     routingInfo,
			}

			updateErr = client.UpdateDesiredLRP(createLRPReq.ProcessGuid, update)
		})

		It("responds without an error", func() {
			Ω(updateErr).ShouldNot(HaveOccurred())
		})

		It("updates the LRP in the BBS", func() {
			Eventually(bbs.DesiredLRPs).Should(HaveLen(1))
			desiredLRPs, err := bbs.DesiredLRPs()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(desiredLRPs[0].Instances).To(Equal(instances))
			Ω(desiredLRPs[0].Annotation).To(Equal(annotation))
			Ω(desiredLRPs[0].Routes).To(Equal(map[string]*json.RawMessage(routingInfo)))
		})
	})

	Describe("DELETE /v1/desired_lrps/:process_guid", func() {
		var lrpRequest receptor.DesiredLRPCreateRequest
		var deleteErr error

		BeforeEach(func() {
			lrpRequest = newValidDesiredLRPCreateRequest()
			err := client.CreateDesiredLRP(lrpRequest)
			Ω(err).ShouldNot(HaveOccurred())

			deleteErr = client.DeleteDesiredLRP(lrpRequest.ProcessGuid)
		})

		It("responds without an error", func() {
			Ω(deleteErr).ShouldNot(HaveOccurred())
		})

		It("deletes the desired lrp with the matching process guid", func() {
			_, getErr := client.GetDesiredLRP(lrpRequest.ProcessGuid)
			Ω(getErr).Should(BeAssignableToTypeOf(receptor.Error{}))
			Ω(getErr.(receptor.Error).Type).Should(Equal(receptor.DesiredLRPNotFound))
		})
	})

	Describe("GET /v1/desired_lrps", func() {
		var lrpResponses []receptor.DesiredLRPResponse
		const expectedLRPCount = 6
		var getErr error

		BeforeEach(func() {
			for i := 0; i < expectedLRPCount; i++ {
				err := client.CreateDesiredLRP(newValidDesiredLRPCreateRequest())
				Ω(err).ShouldNot(HaveOccurred())
			}
			lrpResponses, getErr = client.DesiredLRPs()
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the desired lrps", func() {
			Ω(lrpResponses).Should(HaveLen(expectedLRPCount))
		})
	})

	Describe("GET /v1/domains/:domain/desired_lrps", func() {
		const expectedDomain = "domain-1"
		const expectedLRPCount = 5
		var lrpResponses []receptor.DesiredLRPResponse
		var getErr error

		BeforeEach(func() {
			for i := 0; i < expectedLRPCount; i++ {
				lrp := newValidDesiredLRPCreateRequest()
				lrp.Domain = expectedDomain
				err := client.CreateDesiredLRP(lrp)
				Ω(err).ShouldNot(HaveOccurred())
			}
			for i := 0; i < expectedLRPCount; i++ {
				lrp := newValidDesiredLRPCreateRequest()
				lrp.Domain = "wrong-domain"
				err := client.CreateDesiredLRP(lrp)
				Ω(err).ShouldNot(HaveOccurred())
			}
			lrpResponses, getErr = client.DesiredLRPsByDomain(expectedDomain)
		})

		It("responds without an error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("fetches all of the desired lrps", func() {
			Ω(lrpResponses).Should(HaveLen(expectedLRPCount))
		})
	})
})

var processId int64

func newValidDesiredLRPCreateRequest() receptor.DesiredLRPCreateRequest {
	atomic.AddInt64(&processId, 1)

	return receptor.DesiredLRPCreateRequest{
		ProcessGuid: fmt.Sprintf("process-guid-%d", processId),
		Domain:      "test-domain",
		RootFS:      "some:rootfs",
		Instances:   1,
		Ports:       []uint16{1234, 5678},
		Action: &models.RunAction{
			Path: "/bin/bash",
		},
	}
}
