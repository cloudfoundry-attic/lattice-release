package main_test

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actual LRP API", func() {
	const lrpCount = 6

	var (
		evacuatingLRPKey models.ActualLRPKey
	)

	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)

		for i := 0; i < lrpCount; i++ {
			index := strconv.Itoa(i)
			lrpKey := models.NewActualLRPKey(
				"process-guid-"+index,
				int32(i),
				fmt.Sprintf("domain-%d", i/2),
			)
			instanceKey := models.NewActualLRPInstanceKey(
				"instance-guid-"+index,
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", models.NewPortMapping(uint32(1000+i), 80))
			err := bbsClient.StartActualLRP(&lrpKey, &instanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())
		}

		desiredLRP := &models.DesiredLRP{
			ProcessGuid: "process-guid-0",
			Domain:      "domain-0",
			Instances:   1,
			RootFs:      "some:rootfs",
			Ports:       []uint32{80},
			Action:      models.WrapAction(&models.RunAction{User: "me", Path: "/bin/true"}),
		}

		err := bbsClient.DesireLRP(desiredLRP)
		Expect(err).NotTo(HaveOccurred())

		evacuatingLRPKey = models.NewActualLRPKey("process-guid-0", 0, "domain-0")
		instanceKey := models.NewActualLRPInstanceKey("instance-guid-0", "cell-id")
		netInfo := models.NewActualLRPNetInfo("the-host", models.NewPortMapping(1000, 80))
		_, bbsErr := bbsClient.EvacuateRunningActualLRP(&evacuatingLRPKey, &instanceKey, &netInfo, 0)
		Expect(bbsErr).To(Equal(models.ErrUnknownError))
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("ActualLRPs", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPs()
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Expect(actualLRPResponses).To(HaveLen(lrpCount))
		})

		It("has the correct data from the BBS", func() {
			actualLRPGroups, err := bbsClient.ActualLRPGroups(models.ActualLRPFilter{})
			Expect(err).NotTo(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, lrpCount)
			for _, actualLRPGroup := range actualLRPGroups {
				actualLRP, evacuating := actualLRPGroup.Resolve()
				if actualLRP.ActualLRPKey.Equal(evacuatingLRPKey) {
					continue
				}

				expectedResponses = append(expectedResponses, serialization.ActualLRPProtoToResponse(actualLRP, evacuating))
			}

			Expect(actualLRPResponses).To(ConsistOf(expectedResponses))
		})
	})

	Describe("ActualLRPsByDomain", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		BeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPsByDomain("domain-0")
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches all of the actual lrps", func() {
			Expect(actualLRPResponses).To(HaveLen(2))
		})

		It("has the correct data from the bbs", func() {
			expectedResponses := []receptor.ActualLRPResponse{}

			instanceLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex("process-guid-1", 1)
			Expect(err).NotTo(HaveOccurred())
			expectedResponses = append(expectedResponses, serialization.ActualLRPProtoToResponse(instanceLRPGroup.Instance, false))

			evacuatingLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(evacuatingLRPKey.ProcessGuid, int(evacuatingLRPKey.Index))
			Expect(err).NotTo(HaveOccurred())
			expectedResponses = append(expectedResponses, serialization.ActualLRPProtoToResponse(evacuatingLRPGroup.Evacuating, true))

			Expect(actualLRPResponses).To(ConsistOf(expectedResponses))
		})
	})

	Describe("ActualLRPsByProcessGuid", func() {
		var actualLRPResponses []receptor.ActualLRPResponse
		var getErr error

		JustBeforeEach(func() {
			actualLRPResponses, getErr = client.ActualLRPsByProcessGuid("process-guid-0")
		})

		It("responds without an error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("fetches all of the actual lrps for the process guid", func() {
			Expect(actualLRPResponses).To(HaveLen(1))
		})

		It("has the correct data from the bbs", func() {
			evacuatingLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(evacuatingLRPKey.ProcessGuid, int(evacuatingLRPKey.Index))
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPResponses).To(ConsistOf(serialization.ActualLRPProtoToResponse(evacuatingLRPGroup.Evacuating, true)))
		})
	})

	Describe("ActualLRPsByProcessGuidAndIndex", func() {
		var actualLRPResponse receptor.ActualLRPResponse
		var getErr error
		var processGuid string
		var index int

		BeforeEach(func() {
			processGuid = "process-guid-0"
			index = 1

			lrpKey := models.NewActualLRPKey(
				processGuid,
				int32(index),
				"domain-0",
			)
			instanceKey := models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", models.NewPortMapping(2345, 80))
			err := bbsClient.StartActualLRP(&lrpKey, &instanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())

			actualLRPResponse, getErr = client.ActualLRPByProcessGuidAndIndex(processGuid, index)
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			actualLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(processGuid, index)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPResponse).To(Equal(serialization.ActualLRPProtoToResponse(actualLRPGroup.Instance, false)))
		})
	})
})
