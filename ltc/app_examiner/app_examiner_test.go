package app_examiner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_noaa_consumer"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry/sonde-go/events"
)

var _ = Describe("AppExaminer", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		fakeNoaaConsumer   *fake_noaa_consumer.FakeNoaaConsumer
		appExaminer        app_examiner.AppExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		fakeNoaaConsumer = &fake_noaa_consumer.FakeNoaaConsumer{}
		appExaminer = app_examiner.New(fakeReceptorClient, fakeNoaaConsumer)
	})

	Describe("ListApps", func() {
		Context("with the receptor returning both desiredlrps and actuallrps", func() {
			BeforeEach(func() {
				desiredLrps := []receptor.DesiredLRPResponse{
					receptor.DesiredLRPResponse{
						ProcessGuid: "process2-scalingDown",
						Instances:   0,
						DiskMB:      564,
						MemoryMB:    200,
						Routes:      route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"ren", "stimpy"}}}.RoutingInfo(),
					},
					receptor.DesiredLRPResponse{
						ProcessGuid:          "process1-scalingUp",
						Instances:            2,
						DiskMB:               256,
						MemoryMB:             100,
						Routes:               route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"happy", "joy"}}}.RoutingInfo(),
						EnvironmentVariables: []receptor.EnvironmentVariable{},
						StartTimeout:         30,
						CPUWeight:            94,
						Ports:                []uint16{2378, 67},
						LogGuid:              "asdf-ojf93-9sdcsdk",
						LogSource:            "proc1-log",
						Annotation:           "Best process this side o' the Mississippi.",
					},
				}
				fakeReceptorClient.DesiredLRPsReturns(desiredLrps, nil)

				actualLrps := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{ProcessGuid: "process3-stopping", InstanceGuid: "guid4", Index: 1, State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{ProcessGuid: "process1-scalingUp", InstanceGuid: "guid1", Index: 1, State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{ProcessGuid: "process1-scalingUp", InstanceGuid: "guid2", Index: 2, State: receptor.ActualLRPStateClaimed},
					receptor.ActualLRPResponse{ProcessGuid: "process2-scalingDown", InstanceGuid: "guid3", Index: 1, State: receptor.ActualLRPStateRunning},
				}
				fakeReceptorClient.ActualLRPsReturns(actualLrps, nil)
			})

			It("returns a list of alphabetically sorted examined apps", func() {
				appList, err := appExaminer.ListApps()

				Expect(err).ToNot(HaveOccurred())
				Expect(appList).To(HaveLen(3))

				process1 := appList[0]
				Expect(process1.ProcessGuid).To(Equal("process1-scalingUp"))
				Expect(process1.DesiredInstances).To(Equal(2))
				Expect(process1.ActualRunningInstances).To(Equal(1))
				Expect(process1.DiskMB).To(Equal(256))
				Expect(process1.MemoryMB).To(Equal(100))
				Expect(process1.Routes).To(ConsistOf(route_helpers.AppRoute{Hostnames: []string{"happy", "joy"}}))
				Expect(process1.EnvironmentVariables).To(Equal([]app_examiner.EnvironmentVariable{}))
				Expect(process1.StartTimeout).To(Equal(uint(30)))
				Expect(process1.CPUWeight).To(Equal(uint(94)))
				Expect(process1.Ports).To(Equal([]uint16{2378, 67}))
				Expect(process1.LogGuid).To(Equal("asdf-ojf93-9sdcsdk"))
				Expect(process1.LogSource).To(Equal("proc1-log"))
				Expect(process1.Annotation).To(Equal("Best process this side o' the Mississippi."))

				process2 := appList[1]
				Expect(process2.ProcessGuid).To(Equal("process2-scalingDown"))
				Expect(process2.DesiredInstances).To(Equal(0))
				Expect(process2.ActualRunningInstances).To(Equal(1))
				Expect(process2.DiskMB).To(Equal(564))
				Expect(process2.MemoryMB).To(Equal(200))
				Expect(process2.Routes).To(ConsistOf(route_helpers.AppRoute{Hostnames: []string{"ren", "stimpy"}}))

				process3 := appList[2]
				Expect(process3.ProcessGuid).To(Equal("process3-stopping"))
				Expect(process3.DesiredInstances).To(Equal(0))
				Expect(process3.ActualRunningInstances).To(Equal(1))
				Expect(process3.DiskMB).To(Equal(0))
				Expect(process3.MemoryMB).To(Equal(0))
				Expect(process3.Routes).To(BeEmpty())
			})
		})

		Context("when the receptor returns errors", func() {
			It("returns errors from from fetching the DesiredLRPs", func() {
				fakeReceptorClient.DesiredLRPsReturns(nil, errors.New("You should go catch it."))
				_, err := appExaminer.ListApps()

				Expect(err).To(MatchError("You should go catch it."))
			})

			It("returns errors from fetching the ActualLRPs", func() {
				fakeReceptorClient.DesiredLRPsReturns(nil, nil)
				fakeReceptorClient.ActualLRPsReturns(nil, errors.New("Receptor is on fire!!"))
				_, err := appExaminer.ListApps()

				Expect(err).To(MatchError("Receptor is on fire!!"))
			})
		})
	})

	Describe("ListCells", func() {
		Context("receptor returns actual lrps that are all on existing cells", func() {
			BeforeEach(func() {
				actualLrps := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{CellID: "Cell-1", State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{CellID: "Cell-1", State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{CellID: "Cell-2", State: receptor.ActualLRPStateClaimed},
					receptor.ActualLRPResponse{CellID: "Cell-2", State: receptor.ActualLRPStateRunning},
				}
				fakeReceptorClient.ActualLRPsReturns(actualLrps, nil)

				cells := []receptor.CellResponse{
					receptor.CellResponse{CellID: "Cell-1", Zone: "z1", Capacity: receptor.CellCapacity{MemoryMB: 12394, DiskMB: 2349083, Containers: 512}},
					receptor.CellResponse{CellID: "Cell-2", Zone: "z1", Capacity: receptor.CellCapacity{MemoryMB: 12394, DiskMB: 2349083, Containers: 512}},
					receptor.CellResponse{CellID: "Cell-3", Zone: "z2", Capacity: receptor.CellCapacity{MemoryMB: 12394, DiskMB: 2349083, Containers: 512}},
				}
				fakeReceptorClient.CellsReturns(cells, nil)
			})

			It("returns a list of alphabetically sorted examined cells", func() {
				cellList, err := appExaminer.ListCells()
				Expect(err).ToNot(HaveOccurred())
				Expect(cellList).To(HaveLen(3))

				cell1 := cellList[0]
				Expect(cell1.CellID).To(Equal("Cell-1"))
				Expect(cell1.RunningInstances).To(Equal(2))
				Expect(cell1.ClaimedInstances).To(Equal(0))
				Expect(cell1.Zone).To(Equal("z1"))
				Expect(cell1.MemoryMB).To(Equal(12394))
				Expect(cell1.DiskMB).To(Equal(2349083))
				Expect(cell1.Containers).To(Equal(512))

				cell2 := cellList[1]
				Expect(cell2.CellID).To(Equal("Cell-2"))
				Expect(cell2.RunningInstances).To(Equal(1))
				Expect(cell2.ClaimedInstances).To(Equal(1))
				Expect(cell2.Zone).To(Equal("z1"))

				cell3 := cellList[2]
				Expect(cell3.CellID).To(Equal("Cell-3"))
				Expect(cell3.RunningInstances).To(Equal(0))
				Expect(cell3.ClaimedInstances).To(Equal(0))
				Expect(cell3.Zone).To(Equal("z2"))
			})
		})

		Context("receptor returns actual lrps, and some of their cells no longer exist", func() {
			BeforeEach(func() {
				actualLrps := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{CellID: "Cell-0", State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{CellID: "Cell-0", State: receptor.ActualLRPStateClaimed},
					receptor.ActualLRPResponse{CellID: "Cell-1", State: receptor.ActualLRPStateRunning},
				}
				fakeReceptorClient.ActualLRPsReturns(actualLrps, nil)
				cells := []receptor.CellResponse{
					receptor.CellResponse{CellID: "Cell-1"},
				}
				fakeReceptorClient.CellsReturns(cells, nil)
			})

			It("returns a list of alphabetically sorted examined cells", func() {
				cellList, err := appExaminer.ListCells()
				Expect(err).ToNot(HaveOccurred())
				Expect(cellList).To(HaveLen(2))

				cell0 := cellList[0]
				Expect(cell0.CellID).To(Equal("Cell-0"))
				Expect(cell0.Missing).To(BeTrue())
				Expect(cell0.RunningInstances).To(Equal(1))
				Expect(cell0.ClaimedInstances).To(Equal(1))
			})
		})

		Context("receptor returns unclaimed actual lrps", func() {
			BeforeEach(func() {
				actualLrps := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{State: receptor.ActualLRPStateUnclaimed},
					receptor.ActualLRPResponse{State: receptor.ActualLRPStateUnclaimed},
				}
				fakeReceptorClient.ActualLRPsReturns(actualLrps, nil)
				fakeReceptorClient.CellsReturns([]receptor.CellResponse{}, nil)
			})

			It("ignores unclaimed actual lrps", func() {
				cellList, err := appExaminer.ListCells()
				Expect(err).NotTo(HaveOccurred())
				Expect(cellList).To(HaveLen(0))
			})
		})

		Context("with the receptor returning errors", func() {
			It("returns errors from from fetching the Cells", func() {
				fakeReceptorClient.CellsReturns(nil, errors.New("You should go catch it."))

				_, err := appExaminer.ListCells()
				Expect(err).To(MatchError("You should go catch it."))
			})

			It("returns errors from fetching the ActualLRPs", func() {
				fakeReceptorClient.ActualLRPsReturns(nil, errors.New("Receptor is Running."))

				_, err := appExaminer.ListCells()
				Expect(err).To(MatchError("Receptor is Running."))
			})

		})
	})

	Describe("AppStatus", func() {
		var (
			getDesiredLRPResponse           receptor.DesiredLRPResponse
			actualLRPsByProcessGuidResponse []receptor.ActualLRPResponse
			containerMetrics                []*events.ContainerMetric
		)

		buildContainerMetric := func(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes, diskBytes uint64) *events.ContainerMetric {
			return &events.ContainerMetric{
				ApplicationId: &applicationId,
				InstanceIndex: &instanceIndex,
				CpuPercentage: &cpuPercentage,
				MemoryBytes:   &memoryBytes,
				DiskBytes:     &diskBytes,
			}
		}

		Context("When receptor successfully responds to all requests", func() {
			BeforeEach(func() {
				getDesiredLRPResponse = receptor.DesiredLRPResponse{
					ProcessGuid: "peekaboo-app",
					Domain:      "welp.org",
					RootFS:      "/var/root-fs",
					Instances:   4,
					EnvironmentVariables: []receptor.EnvironmentVariable{
						receptor.EnvironmentVariable{
							Name:  "API_TOKEN",
							Value: "98weufsa",
						},
						receptor.EnvironmentVariable{
							Name:  "PEEKABOO_APP_NICKNAME",
							Value: "Bugs McGee",
						},
					},
					StartTimeout: 5,
					DiskMB:       256,
					MemoryMB:     128,
					CPUWeight:    77,
					Ports:        []uint16{8765, 2300},
					Routes:       route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"peekaboo-one.example.com", "peekaboo-too.example.com"}}}.RoutingInfo(),
					LogGuid:      "9832-ur98j-idsckl",
					LogSource:    "peekaboo-lawgz",
					Annotation:   "best. game. ever.",
				}

				actualLRPsByProcessGuidResponse = []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{
						ProcessGuid:  "peekaboo-app",
						InstanceGuid: "aisu-8dfy8-9dhu",
						CellID:       "cell-3",
						Domain:       "welp.org",
						Index:        1,
						Address:      "212.38.11.83",
						Ports: []receptor.PortMapping{
							receptor.PortMapping{HostPort: 2983, ContainerPort: 2001},
						},
						State:      "CLAIMED",
						Since:      1982,
						CrashCount: 3,
					}, receptor.ActualLRPResponse{
						ProcessGuid:  "peekaboo-app",
						InstanceGuid: "98s98a-xcvcx4-93isl",
						CellID:       "cell-2",
						Domain:       "welp.org",
						Index:        0,
						Address:      "211.94.88.63",
						Ports: []receptor.PortMapping{
							receptor.PortMapping{HostPort: 2786, ContainerPort: 2020},
						},
						State: "RUNNING",
						Since: 2002,
					}, receptor.ActualLRPResponse{
						ProcessGuid:    "peekaboo-app",
						Index:          2,
						State:          "UNCLAIMED",
						PlacementError: "not enough resources. eek.",
					},
					receptor.ActualLRPResponse{
						ProcessGuid: "peekaboo-app",
						Index:       3,
						State:       "CRASHED",
						CrashCount:  7,
					},
				}

				containerMetrics = []*events.ContainerMetric{
					buildContainerMetric("peekaboo-app", 0, 0.018138574, 798729, 32768),
				}
			})

			It("returns a fully populated AppInfo with instances sorted by index", func() {
				fakeReceptorClient.GetDesiredLRPReturns(getDesiredLRPResponse, nil)
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLRPsByProcessGuidResponse, nil)
				fakeNoaaConsumer.GetContainerMetricsReturns(containerMetrics, nil)

				appInfo, err := appExaminer.AppStatus("peekaboo-app")
				Expect(err).ToNot(HaveOccurred())

				Expect(appInfo.ProcessGuid).To(Equal("peekaboo-app"))
				Expect(appInfo.DesiredInstances).To(Equal(4))
				Expect(appInfo.ActualRunningInstances).To(Equal(1))
				Expect(appInfo.EnvironmentVariables).To(ConsistOf(
					app_examiner.EnvironmentVariable{
						Name:  "API_TOKEN",
						Value: "98weufsa",
					},
					app_examiner.EnvironmentVariable{
						Name:  "PEEKABOO_APP_NICKNAME",
						Value: "Bugs McGee",
					},
				))
				Expect(appInfo.StartTimeout).To(Equal(uint(5)))
				Expect(appInfo.DiskMB).To(Equal(256))
				Expect(appInfo.MemoryMB).To(Equal(128))
				Expect(appInfo.CPUWeight).To(Equal(uint(77)))
				Expect(appInfo.Ports).To(ConsistOf(uint16(8765), uint16(2300)))
				Expect(appInfo.Routes).To(Equal(route_helpers.AppRoutes{
					route_helpers.AppRoute{Hostnames: []string{"peekaboo-one.example.com", "peekaboo-too.example.com"}}},
				))
				Expect(appInfo.LogGuid).To(Equal("9832-ur98j-idsckl"))
				Expect(appInfo.LogSource).To(Equal("peekaboo-lawgz"))
				Expect(appInfo.Annotation).To(Equal("best. game. ever."))

				Expect(appInfo.ActualInstances).To(HaveLen(4))

				instanceZero := appInfo.ActualInstances[0]
				Expect(instanceZero.InstanceGuid).To(Equal("98s98a-xcvcx4-93isl"))
				Expect(instanceZero.CellID).To(Equal("cell-2"))
				Expect(instanceZero.Index).To(Equal(0))
				Expect(instanceZero.Ip).To(Equal("211.94.88.63"))
				Expect(instanceZero.Ports).To(ConsistOf(app_examiner.PortMapping{
					HostPort:      2786,
					ContainerPort: 2020,
				}))
				Expect(instanceZero.State).To(Equal("RUNNING"))
				Expect(instanceZero.Since).To(Equal(int64(2002)))
				Expect(instanceZero.HasMetrics).To(BeTrue())
				Expect(instanceZero.Metrics).To(Equal(app_examiner.InstanceMetrics{
					CpuPercentage: 0.018138574,
					MemoryBytes:   798729,
					DiskBytes:     32768,
				}))

				instanceOne := appInfo.ActualInstances[1]
				Expect(instanceOne.InstanceGuid).To(Equal("aisu-8dfy8-9dhu"))
				Expect(instanceOne.CellID).To(Equal("cell-3"))
				Expect(instanceOne.Index).To(Equal(1))
				Expect(instanceOne.Ip).To(Equal("212.38.11.83"))
				Expect(instanceOne.Ports).To(ConsistOf(app_examiner.PortMapping{
					HostPort:      2983,
					ContainerPort: 2001,
				}))
				Expect(instanceOne.State).To(Equal("CLAIMED"))
				Expect(instanceOne.Since).To(Equal(int64(1982)))
				Expect(instanceOne.CrashCount).To(Equal(3))
				Expect(instanceOne.HasMetrics).To(BeFalse())

				instanceTwo := appInfo.ActualInstances[2]
				Expect(instanceTwo.Index).To(Equal(2))
				Expect(instanceTwo.Ports).To(BeEmpty())
				Expect(instanceTwo.State).To(Equal("UNCLAIMED"))
				Expect(instanceTwo.PlacementError).To(Equal("not enough resources. eek."))
				Expect(instanceTwo.HasMetrics).To(BeFalse())

				instanceThree := appInfo.ActualInstances[3]
				Expect(instanceThree.Index).To(Equal(3))
				Expect(instanceThree.Ports).To(BeEmpty())
				Expect(instanceThree.State).To(Equal("CRASHED"))
				Expect(instanceThree.CrashCount).To(Equal(7))
				Expect(instanceThree.HasMetrics).To(BeFalse())

				Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("peekaboo-app"))

				Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("peekaboo-app"))

				Expect(fakeNoaaConsumer.GetContainerMetricsCallCount()).To(Equal(1))
				appGuid, token := fakeNoaaConsumer.GetContainerMetricsArgsForCall(0)
				Expect(appGuid).To(Equal("peekaboo-app"))
				Expect(token).To(BeEmpty())
			})

			Context("when desired LRP is not found, but there are actual LRPs for the process GUID (App stopping)", func() {
				It("returns AppInfo that has ActualInstances, but is missing desiredlrp specific data", func() {
					fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, receptor.Error{
						Type:    receptor.DesiredLRPNotFound,
						Message: "Desired LRP with guid 'peekaboo-app' not found"},
					)
					fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLRPsByProcessGuidResponse, nil)
					fakeNoaaConsumer.GetContainerMetricsReturns(containerMetrics, nil)

					appInfo, err := appExaminer.AppStatus("peekaboo-app")
					Expect(err).NotTo(HaveOccurred())

					Expect(appInfo.ProcessGuid).To(Equal("peekaboo-app"))
					Expect(appInfo.ActualRunningInstances).To(Equal(1))

					Expect(appInfo.ActualInstances).To(HaveLen(4))

					instanceZero := appInfo.ActualInstances[0]
					Expect(instanceZero.InstanceGuid).To(Equal("98s98a-xcvcx4-93isl"))
					Expect(instanceZero.CellID).To(Equal("cell-2"))
					Expect(instanceZero.Index).To(Equal(0))
					Expect(instanceZero.Ip).To(Equal("211.94.88.63"))
					Expect(instanceZero.Ports).To(ConsistOf(app_examiner.PortMapping{
						HostPort:      2786,
						ContainerPort: 2020,
					}))
					Expect(instanceZero.State).To(Equal("RUNNING"))
					Expect(instanceZero.Since).To(Equal(int64(2002)))
					Expect(instanceZero.HasMetrics).To(BeTrue())
					Expect(instanceZero.Metrics).To(Equal(app_examiner.InstanceMetrics{
						CpuPercentage: 0.018138574,
						MemoryBytes:   798729,
						DiskBytes:     32768,
					}))

					instanceOne := appInfo.ActualInstances[1]
					Expect(instanceOne.InstanceGuid).To(Equal("aisu-8dfy8-9dhu"))
					Expect(instanceOne.CellID).To(Equal("cell-3"))
					Expect(instanceOne.Index).To(Equal(1))
					Expect(instanceOne.Ip).To(Equal("212.38.11.83"))
					Expect(instanceOne.Ports).To(ConsistOf(app_examiner.PortMapping{
						HostPort:      2983,
						ContainerPort: 2001,
					}))
					Expect(instanceOne.State).To(Equal("CLAIMED"))
					Expect(instanceOne.Since).To(Equal(int64(1982)))
					Expect(instanceOne.CrashCount).To(Equal(3))
					Expect(instanceOne.HasMetrics).To(BeFalse())

					instanceTwo := appInfo.ActualInstances[2]
					Expect(instanceTwo.Index).To(Equal(2))
					Expect(instanceTwo.Ports).To(BeEmpty())
					Expect(instanceTwo.State).To(Equal("UNCLAIMED"))
					Expect(instanceTwo.PlacementError).To(Equal("not enough resources. eek."))
					Expect(instanceTwo.HasMetrics).To(BeFalse())

					instanceThree := appInfo.ActualInstances[3]
					Expect(instanceThree.Index).To(Equal(3))
					Expect(instanceThree.Ports).To(BeEmpty())
					Expect(instanceThree.State).To(Equal("CRASHED"))
					Expect(instanceThree.CrashCount).To(Equal(7))
					Expect(instanceThree.HasMetrics).To(BeFalse())

					Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
					Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))

					Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("peekaboo-app"))
					Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("peekaboo-app"))

					Expect(fakeNoaaConsumer.GetContainerMetricsCallCount()).To(Equal(1))
					appGuid, token := fakeNoaaConsumer.GetContainerMetricsArgsForCall(0)
					Expect(appGuid).To(Equal("peekaboo-app"))
					Expect(token).To(BeEmpty())
				})
			})

			It("handles empty desiredLRP with empty actualLRP response", func() {
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, nil)
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(make([]receptor.ActualLRPResponse, 0), nil)

				appInfo, err := appExaminer.AppStatus("peekaboo-app")
				Expect(err).To(MatchError(app_examiner.AppNotFoundErrorMessage))
				Expect(appInfo).To(BeZero())

				Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("peekaboo-app"))

				Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("peekaboo-app"))

				Expect(fakeNoaaConsumer.GetContainerMetricsCallCount()).To(Equal(0))
			})

		})

		Context("when noaa returns container metrics without an associated actual lrp", func() {
			It("doesn't blow up", func() {
				getDesiredLRPResponse := receptor.DesiredLRPResponse{
					ProcessGuid: "peekaboo-app",
				}
				actualLRPs := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{
						ProcessGuid: "peekaboo-app",
						Index:       6,
					},
				}
				containerMetrics := []*events.ContainerMetric{
					buildContainerMetric("peekaboo-app", 42, 0.018138574, 798729, 32768),
				}
				fakeReceptorClient.GetDesiredLRPReturns(getDesiredLRPResponse, nil)
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLRPs, nil)
				fakeNoaaConsumer.GetContainerMetricsReturns(containerMetrics, nil)

				appInfo, err := appExaminer.AppStatus("peekaboo-app")
				Expect(err).NotTo(HaveOccurred())

				Expect(appInfo).ToNot(BeZero())
				Expect(appInfo.ActualInstances).To(HaveLen(1))
				Expect(appInfo.ActualInstances[0].HasMetrics).To(BeFalse())
			})
		})

		Context("when the receptor returns errors", func() {
			It("returns errors from from fetching the DesiredLRPs", func() {
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Oops."},
				)

				_, err := appExaminer.AppStatus("app-to-status")
				Expect(err).To(MatchError("Oops."))
			})

			It("returns errors from fetching the ActualLRPs", func() {
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, nil)
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(nil, receptor.Error{
					Type:    receptor.UnknownError,
					Message: "ABANDON SHIP!!!!"},
				)

				_, err := appExaminer.AppStatus("kiss-my-bumper")
				Expect(err).To(MatchError("ABANDON SHIP!!!!"))
			})
		})

		Context("when the noaa consumer returns errors", func() {
			It("returns an AppInfo without container metrics", func() {
				desiredLRPs := receptor.DesiredLRPResponse{
					ProcessGuid: "peekaboo-app",
				}
				actualLRPs := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{
						ProcessGuid: "peekaboo-app",
						Index:       6,
					},
				}
				fakeReceptorClient.GetDesiredLRPReturns(desiredLRPs, nil)
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLRPs, nil)
				fakeNoaaConsumer.GetContainerMetricsReturns(nil, errors.New("no metrics 4 you"))

				appInfo, err := appExaminer.AppStatus("peekaboo-app")
				Expect(err).NotTo(HaveOccurred())

				Expect(appInfo).ToNot(BeZero())
				Expect(appInfo.ActualInstances).To(HaveLen(1))
				Expect(appInfo.ActualInstances[0].HasMetrics).To(BeFalse())
			})
		})
	})

	Describe("NumOfRunningAppInstances", func() {
		It("returns the number of running instances for a given app guid", func() {
			actualLrpsResponse := []receptor.ActualLRPResponse{
				receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateRunning, Index: 1},
				receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateRunning, Index: 2},
				receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateClaimed, Index: 3},
			}
			fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLrpsResponse, nil)

			count, placementError, err := appExaminer.RunningAppInstancesInfo("americano-app")
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(2))
			Expect(placementError).To(BeFalse())

			Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("americano-app"))
		})

		It("returns errors from the receptor", func() {
			receptorError := errors.New("receptor did not like that request")
			fakeReceptorClient.ActualLRPsByProcessGuidReturns([]receptor.ActualLRPResponse{}, receptorError)

			_, _, err := appExaminer.RunningAppInstancesInfo("nescafe-app")
			Expect(err).To(MatchError(receptorError))
		})

		Context("when there are placement errors on an instance", func() {
			It("returns true for placementError", func() {
				actualLrpsResponse := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateRunning, Index: 1},
					receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateUnclaimed, Index: 2, PlacementError: "could not place!"},
					receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateRunning, Index: 3},
				}
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLrpsResponse, nil)

				count, placementError, err := appExaminer.RunningAppInstancesInfo("americano-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(2))
				Expect(placementError).To(BeTrue())
			})
		})
	})

	Describe("AppExists", func() {
		It("returns true if the docker app exists", func() {
			actualLRPs := []receptor.ActualLRPResponse{receptor.ActualLRPResponse{ProcessGuid: "americano-app"}}
			fakeReceptorClient.ActualLRPsReturns(actualLRPs, nil)

			exists, err := appExaminer.AppExists("americano-app")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns false if the docker app does not exist", func() {
			actualLRPs := []receptor.ActualLRPResponse{}
			fakeReceptorClient.ActualLRPsReturns(actualLRPs, nil)

			exists, err := appExaminer.AppExists("americano-app")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		Describe("returning errors from the receptor", func() {
			It("returns errors fetching the status", func() {
				actualLRPs := []receptor.ActualLRPResponse{}
				fakeReceptorClient.ActualLRPsReturns(actualLRPs, errors.New("Something Bad"))

				exists, err := appExaminer.AppExists("americano-app")
				Expect(err).To(MatchError("Something Bad"))
				Expect(exists).To(BeFalse())
			})
		})
	})
})
