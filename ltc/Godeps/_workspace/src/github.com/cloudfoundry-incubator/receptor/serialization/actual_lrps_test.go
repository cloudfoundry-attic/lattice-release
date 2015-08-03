package serialization_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP Serialization", func() {
	Describe("ActualLRPProtoToResponse", func() {
		var actualLRP *models.ActualLRP

		BeforeEach(func() {
			actualLRP = &models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-0",
					3,
					"some-domain",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-0",
					"cell-id-0",
				),
				ActualLRPNetInfo: models.NewActualLRPNetInfo(
					"address-0",
					models.NewPortMapping(9876, 2345),
				),
				State:      models.ActualLRPStateRunning,
				CrashCount: 42,
				Since:      99999999999,
				ModificationTag: models.ModificationTag{
					Epoch: "some-guid",
					Index: 50,
				},
			}
		})

		It("serializes all the fields", func() {
			expectedResponse := receptor.ActualLRPResponse{
				ProcessGuid:  "process-guid-0",
				InstanceGuid: "instance-guid-0",
				CellID:       "cell-id-0",
				Domain:       "some-domain",
				Index:        3,
				Address:      "address-0",
				Ports: []receptor.PortMapping{
					{
						ContainerPort: 2345,
						HostPort:      9876,
					},
				},
				State:      receptor.ActualLRPStateRunning,
				CrashCount: 42,
				Since:      99999999999,
				Evacuating: true,
				ModificationTag: receptor.ModificationTag{
					Epoch: "some-guid",
					Index: 50,
				},
			}

			actualResponse := serialization.ActualLRPProtoToResponse(actualLRP, true)
			Expect(actualResponse).To(Equal(expectedResponse))
		})

		It("maps model states to receptor states", func() {
			expectedStateMap := map[string]receptor.ActualLRPState{
				models.ActualLRPStateUnclaimed: receptor.ActualLRPStateUnclaimed,
				models.ActualLRPStateClaimed:   receptor.ActualLRPStateClaimed,
				models.ActualLRPStateRunning:   receptor.ActualLRPStateRunning,
				models.ActualLRPStateCrashed:   receptor.ActualLRPStateCrashed,
			}

			for modelState, jsonState := range expectedStateMap {
				actualLRP.State = modelState
				Expect(serialization.ActualLRPProtoToResponse(actualLRP, false).State).To(Equal(jsonState))
			}

			actualLRP.State = ""
			Expect(serialization.ActualLRPProtoToResponse(actualLRP, false).State).To(Equal(receptor.ActualLRPStateInvalid))
		})

		Context("when there is placement error", func() {
			BeforeEach(func() {
				actualLRP.State = models.ActualLRPStateUnclaimed
				actualLRP.PlacementError = "some-error"
			})

			It("includes the placement error", func() {
				actualResponse := serialization.ActualLRPProtoToResponse(actualLRP, false)
				Expect(actualResponse.PlacementError).To(Equal("some-error"))
			})
		})

		Context("when there is a crash reason", func() {
			BeforeEach(func() {
				actualLRP.State = models.ActualLRPStateCrashed
				actualLRP.CrashReason = "crashed"
			})

			It("includes the placement error", func() {
				actualResponse := serialization.ActualLRPProtoToResponse(actualLRP, false)
				Expect(actualResponse.CrashReason).To(Equal("crashed"))
			})
		})
	})
})
