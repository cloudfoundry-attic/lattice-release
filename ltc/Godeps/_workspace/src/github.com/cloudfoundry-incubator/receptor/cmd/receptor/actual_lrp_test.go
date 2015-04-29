package main_test

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
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
				i,
				fmt.Sprintf("domain-%d", i/2),
			)
			instanceKey := models.NewActualLRPInstanceKey(
				"instance-guid-"+index,
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", []models.PortMapping{{ContainerPort: 80, HostPort: uint16(1000 + i)}})
			err := bbs.StartActualLRP(logger, lrpKey, instanceKey, netInfo)
			Expect(err).NotTo(HaveOccurred())
		}

		desiredLRP := models.DesiredLRP{
			ProcessGuid: "process-guid-0",
			Domain:      "domain-0",
			Instances:   1,
			RootFS:      "some:rootfs",
			Ports:       []uint16{80},
			Action:      &models.RunAction{Path: "/bin/true"},
		}

		err := bbs.DesireLRP(logger, desiredLRP)
		Expect(err).NotTo(HaveOccurred())

		evacuatingLRPKey = models.NewActualLRPKey("process-guid-0", 0, "domain-0")
		instanceKey := models.NewActualLRPInstanceKey("instance-guid-0", "cell-id")
		netInfo := models.NewActualLRPNetInfo("the-host", []models.PortMapping{{ContainerPort: 80, HostPort: 1000}})
		_, err = bbs.EvacuateRunningActualLRP(logger, evacuatingLRPKey, instanceKey, netInfo, 0)
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

		It("has the correct data from the bbs", func() {
			actualLRPs, err := bbs.ActualLRPs()
			Expect(err).NotTo(HaveOccurred())

			expectedResponses := make([]receptor.ActualLRPResponse, 0, lrpCount)
			for _, actualLRP := range actualLRPs {
				if actualLRP.ActualLRPKey == evacuatingLRPKey {
					continue
				}

				expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(actualLRP, false))
			}

			evacuatingLRP, err := bbs.EvacuatingActualLRPByProcessGuidAndIndex(evacuatingLRPKey.ProcessGuid, evacuatingLRPKey.Index)
			Expect(err).NotTo(HaveOccurred())
			expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(evacuatingLRP, true))

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

			instanceLRPGroup, err := bbs.ActualLRPGroupByProcessGuidAndIndex("process-guid-1", 1)
			Expect(err).NotTo(HaveOccurred())
			expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(*instanceLRPGroup.Instance, false))

			evacuatingLRP, err := bbs.EvacuatingActualLRPByProcessGuidAndIndex(evacuatingLRPKey.ProcessGuid, evacuatingLRPKey.Index)
			Expect(err).NotTo(HaveOccurred())
			expectedResponses = append(expectedResponses, serialization.ActualLRPToResponse(evacuatingLRP, true))

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
			evacuatingLRP, err := bbs.EvacuatingActualLRPByProcessGuidAndIndex(evacuatingLRPKey.ProcessGuid, evacuatingLRPKey.Index)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualLRPResponses).To(ConsistOf(serialization.ActualLRPToResponse(evacuatingLRP, true)))
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
				index,
				"domain-0",
			)
			instanceKey := models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id",
			)
			netInfo := models.NewActualLRPNetInfo("the-host", []models.PortMapping{{ContainerPort: 80, HostPort: 2345}})
			err := bbs.StartActualLRP(logger, lrpKey, instanceKey, netInfo)
			Expect(err).NotTo(HaveOccurred())

			actualLRPResponse, getErr = client.ActualLRPByProcessGuidAndIndex(processGuid, index)
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct data from the bbs", func() {
			actualLRPGroup, err := bbs.ActualLRPGroupByProcessGuidAndIndex(processGuid, index)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPResponse).To(Equal(serialization.ActualLRPToResponse(*actualLRPGroup.Instance, false)))
		})
	})
})
