package main_test

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
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
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("desires an LRP in the BBS", func() {
			Eventually(func() ([]*models.DesiredLRP, error) { return bbsClient.DesiredLRPs(models.DesiredLRPFilter{}) }).Should(HaveLen(1))
			desiredLRPs, err := bbsClient.DesiredLRPs(models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPs[0].ProcessGuid).To(Equal(lrpToCreate.ProcessGuid))
		})

		Context("when the desired LRP already exists", func() {
			It("fails the request with an appropriate error", func() {
				err := client.CreateDesiredLRP(lrpToCreate)
				Expect(err).To(BeAssignableToTypeOf(receptor.Error{}))
				Expect(err.(receptor.Error).Type).To(Equal(receptor.DesiredLRPAlreadyExists))
			})
		})
	})

	Describe("GET /v1/desired_lrps/:process_guid", func() {
		var lrpRequest receptor.DesiredLRPCreateRequest
		var lrpResponse receptor.DesiredLRPResponse
		var getErr error
		var desiredLRP *models.DesiredLRP

		BeforeEach(func() {
			lrpRequest = newValidDesiredLRPCreateRequest()
			err := client.CreateDesiredLRP(lrpRequest)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err = bbsClient.DesiredLRPByProcessGuid(lrpRequest.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			lrpResponse, getErr = client.GetDesiredLRP(lrpRequest.ProcessGuid)
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches the desired lrp with the matching process guid", func() {
			expectedLRPResponse := serialization.DesiredLRPProtoToResponse(desiredLRP)
			actualJSON, err := json.Marshal(lrpResponse)
			Expect(err).NotTo(HaveOccurred())
			expectedJSON, err := json.Marshal(expectedLRPResponse)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualJSON).To(MatchJSON(expectedJSON))
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
			Expect(err).NotTo(HaveOccurred())

			update := receptor.DesiredLRPUpdateRequest{
				Instances:  &instances,
				Annotation: &annotation,
				Routes:     routingInfo,
			}

			updateErr = client.UpdateDesiredLRP(createLRPReq.ProcessGuid, update)
		})

		It("responds without an error", func() {
			Expect(updateErr).NotTo(HaveOccurred())
		})

		It("updates the LRP in the BBS", func() {
			Eventually(func() ([]*models.DesiredLRP, error) { return bbsClient.DesiredLRPs(models.DesiredLRPFilter{}) }).Should(HaveLen(1))
			desiredLRPs, err := bbsClient.DesiredLRPs(models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPs[0].Instances).To(BeNumerically("==", instances))
			Expect(desiredLRPs[0].Annotation).To(Equal(annotation))
			Expect(*desiredLRPs[0].Routes).To(BeEquivalentTo(map[string]*json.RawMessage(routingInfo)))
		})
	})

	Describe("DELETE /v1/desired_lrps/:process_guid", func() {
		var lrpRequest receptor.DesiredLRPCreateRequest
		var deleteErr error

		BeforeEach(func() {
			lrpRequest = newValidDesiredLRPCreateRequest()
			err := client.CreateDesiredLRP(lrpRequest)
			Expect(err).NotTo(HaveOccurred())

			deleteErr = client.DeleteDesiredLRP(lrpRequest.ProcessGuid)
		})

		It("responds without an error", func() {
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("deletes the desired lrp with the matching process guid", func() {
			_, getErr := client.GetDesiredLRP(lrpRequest.ProcessGuid)
			Expect(getErr).To(BeAssignableToTypeOf(receptor.Error{}))
			Expect(getErr.(receptor.Error).Type).To(Equal(receptor.DesiredLRPNotFound))
		})
	})

	Describe("GET /v1/desired_lrps", func() {
		var lrpResponses []receptor.DesiredLRPResponse
		const expectedLRPCount = 6
		var getErr error

		BeforeEach(func() {
			for i := 0; i < expectedLRPCount; i++ {
				err := client.CreateDesiredLRP(newValidDesiredLRPCreateRequest())
				Expect(err).NotTo(HaveOccurred())
			}
			lrpResponses, getErr = client.DesiredLRPs()
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches all of the desired lrps", func() {
			Expect(lrpResponses).To(HaveLen(expectedLRPCount))
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
				Expect(err).NotTo(HaveOccurred())
			}
			for i := 0; i < expectedLRPCount; i++ {
				lrp := newValidDesiredLRPCreateRequest()
				lrp.Domain = "wrong-domain"
				err := client.CreateDesiredLRP(lrp)
				Expect(err).NotTo(HaveOccurred())
			}
			lrpResponses, getErr = client.DesiredLRPsByDomain(expectedDomain)
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches all of the desired lrps", func() {
			Expect(lrpResponses).To(HaveLen(expectedLRPCount))
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
		Action: &oldmodels.RunAction{
			User: "me",
			Path: "/bin/bash",
		},
	}
}
