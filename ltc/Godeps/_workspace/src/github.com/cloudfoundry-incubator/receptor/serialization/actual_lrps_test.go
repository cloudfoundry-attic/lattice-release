package serialization_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/diego_errors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP Serialization", func() {
	Describe("ActualLRPToResponse", func() {
		var actualLRP models.ActualLRP
		BeforeEach(func() {
			actualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-0",
					3,
					"some-domain",
				),
				ActualLRPContainerKey: models.NewActualLRPContainerKey(
					"instance-guid-0",
					"cell-id-0",
				),
				ActualLRPNetInfo: models.NewActualLRPNetInfo(
					"address-0",
					[]models.PortMapping{
						{
							ContainerPort: 2345,
							HostPort:      9876,
						},
					},
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

			actualResponse := serialization.ActualLRPToResponse(actualLRP, true)
			Ω(actualResponse).Should(Equal(expectedResponse))
		})

		It("maps model states to receptor states", func() {
			expectedStateMap := map[models.ActualLRPState]receptor.ActualLRPState{
				models.ActualLRPStateUnclaimed: receptor.ActualLRPStateUnclaimed,
				models.ActualLRPStateClaimed:   receptor.ActualLRPStateClaimed,
				models.ActualLRPStateRunning:   receptor.ActualLRPStateRunning,
				models.ActualLRPStateCrashed:   receptor.ActualLRPStateCrashed,
			}

			for modelState, jsonState := range expectedStateMap {
				actualLRP.State = modelState
				Ω(serialization.ActualLRPToResponse(actualLRP, false).State).Should(Equal(jsonState))
			}

			actualLRP.State = ""
			Ω(serialization.ActualLRPToResponse(actualLRP, false).State).Should(Equal(receptor.ActualLRPStateInvalid))
		})

		Context("when there is placement error", func() {
			BeforeEach(func() {
				actualLRP.State = models.ActualLRPStateUnclaimed
				actualLRP.PlacementError = diego_errors.INSUFFICIENT_RESOURCES_MESSAGE
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
					State:          receptor.ActualLRPStateUnclaimed,
					PlacementError: diego_errors.INSUFFICIENT_RESOURCES_MESSAGE,
					CrashCount:     42,
					Since:          99999999999,
					ModificationTag: receptor.ModificationTag{
						Epoch: "some-guid",
						Index: 50,
					},
				}

				actualResponse := serialization.ActualLRPToResponse(actualLRP, false)
				Ω(actualResponse).Should(Equal(expectedResponse))
			})
		})
	})

	Describe("ActualLRPFromResponse", func() {
		var actualLRPResponse receptor.ActualLRPResponse

		BeforeEach(func() {
			actualLRPResponse = receptor.ActualLRPResponse{
				ProcessGuid:  "process-guid-0",
				InstanceGuid: "instance-guid",
				CellID:       "cell-id",
				Domain:       "domain",
				Index:        0,
				Address:      "address",
				Ports:        []receptor.PortMapping{{ContainerPort: 10000, HostPort: 10000}},
				State:        receptor.ActualLRPStateRunning,
				Since:        99999999999,
				ModificationTag: receptor.ModificationTag{
					Epoch: "some-guid",
					Index: 50,
				},
			}
		})

		It("deserializes all the fields", func() {
			actualLRP := serialization.ActualLRPFromResponse(actualLRPResponse)
			Ω(actualLRP).Should(Equal(models.ActualLRP{
				ActualLRPKey:          models.NewActualLRPKey("process-guid-0", 0, "domain"),
				ActualLRPContainerKey: models.NewActualLRPContainerKey("instance-guid", "cell-id"),
				ActualLRPNetInfo:      models.NewActualLRPNetInfo("address", []models.PortMapping{{ContainerPort: 10000, HostPort: 10000}}),
				State:                 models.ActualLRPStateRunning,
				Since:                 99999999999,
				ModificationTag: models.ModificationTag{
					Epoch: "some-guid",
					Index: 50,
				},
			}))
		})

		It("maps receptor states to model states", func() {
			expectedStateMap := map[receptor.ActualLRPState]models.ActualLRPState{
				receptor.ActualLRPStateUnclaimed: models.ActualLRPStateUnclaimed,
				receptor.ActualLRPStateClaimed:   models.ActualLRPStateClaimed,
				receptor.ActualLRPStateRunning:   models.ActualLRPStateRunning,
			}

			for jsonState, modelState := range expectedStateMap {
				actualLRPResponse.State = jsonState
				Ω(serialization.ActualLRPFromResponse(actualLRPResponse).State).Should(Equal(modelState))
			}

			actualLRPResponse.State = ""
			Ω(serialization.ActualLRPFromResponse(actualLRPResponse).State).Should(Equal(models.ActualLRPState("")))
		})

		Context("when there is placement error", func() {
			BeforeEach(func() {
				actualLRPResponse.State = receptor.ActualLRPStateUnclaimed
				actualLRPResponse.PlacementError = diego_errors.INSUFFICIENT_RESOURCES_MESSAGE
			})

			It("deserializes all the fields", func() {
				actualLRP := serialization.ActualLRPFromResponse(actualLRPResponse)
				Ω(actualLRP).Should(Equal(models.ActualLRP{
					ActualLRPKey:          models.NewActualLRPKey("process-guid-0", 0, "domain"),
					ActualLRPContainerKey: models.NewActualLRPContainerKey("instance-guid", "cell-id"),
					ActualLRPNetInfo:      models.NewActualLRPNetInfo("address", []models.PortMapping{{ContainerPort: 10000, HostPort: 10000}}),
					State:                 models.ActualLRPStateUnclaimed,
					PlacementError:        diego_errors.INSUFFICIENT_RESOURCES_MESSAGE,
					Since:                 99999999999,
					ModificationTag: models.ModificationTag{
						Epoch: "some-guid",
						Index: 50,
					},
				}))
			})
		})
	})
})
