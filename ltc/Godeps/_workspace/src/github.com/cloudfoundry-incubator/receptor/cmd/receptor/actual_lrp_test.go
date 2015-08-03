package main_test

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actual LRP API", func() {
	const lrpCount = 6

	var (
		evacuatingLRPKey    models.ActualLRPKey
		oldEvacuatingLRPKey oldmodels.ActualLRPKey
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
			_, err := bbsClient.StartActualLRP(&lrpKey, &instanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())
		}

		desiredLRP := oldmodels.DesiredLRP{
			ProcessGuid: "process-guid-0",
			Domain:      "domain-0",
			Instances:   1,
			RootFS:      "some:rootfs",
			Ports:       []uint16{80},
			Action:      &oldmodels.RunAction{User: "me", Path: "/bin/true"},
		}

		err := legacyBBS.DesireLRP(logger, desiredLRP)
		Expect(err).NotTo(HaveOccurred())

		evacuatingLRPKey = models.NewActualLRPKey("process-guid-0", 0, "domain-0")
		oldEvacuatingLRPKey = oldmodels.NewActualLRPKey("process-guid-0", 0, "domain-0")
		instanceKey := oldmodels.NewActualLRPInstanceKey("instance-guid-0", "cell-id")
		netInfo := oldmodels.NewActualLRPNetInfo("the-host", []oldmodels.PortMapping{{ContainerPort: 80, HostPort: 1000}})
		_, err = legacyBBS.EvacuateRunningActualLRP(logger, oldEvacuatingLRPKey, instanceKey, netInfo, 0)
		Expect(err).To(Equal(bbserrors.ErrServiceUnavailable))
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
			_, err := bbsClient.StartActualLRP(&lrpKey, &instanceKey, &netInfo)
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
