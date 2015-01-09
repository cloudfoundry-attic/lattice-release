package app_examiner_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
)

var _ = Describe("AppRunner", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		appExaminer        app_examiner.AppExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		appExaminer = app_examiner.New(fakeReceptorClient)

	})

	Describe("ListApps", func() {
		Context("with the receptor returning both desiredlrps and actuallrps", func() {
			BeforeEach(func() {
				desiredLrps := []receptor.DesiredLRPResponse{
					receptor.DesiredLRPResponse{ProcessGuid: "process2", Instances: 0, DiskMB: 564, MemoryMB: 200, Routes: []string{"ren", "stimpy"}},
					receptor.DesiredLRPResponse{ProcessGuid: "process1", Instances: 2, DiskMB: 256, MemoryMB: 100, Routes: []string{"happy", "joy"}},
				}
				fakeReceptorClient.DesiredLRPsReturns(desiredLrps, nil)

				actualLrps := []receptor.ActualLRPResponse{
					receptor.ActualLRPResponse{ProcessGuid: "process3", InstanceGuid: "guid4", Index: 1, State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{ProcessGuid: "process1", InstanceGuid: "guid1", Index: 1, State: receptor.ActualLRPStateRunning},
					receptor.ActualLRPResponse{ProcessGuid: "process1", InstanceGuid: "guid2", Index: 2, State: receptor.ActualLRPStateClaimed},
					receptor.ActualLRPResponse{ProcessGuid: "process2", InstanceGuid: "guid3", Index: 1, State: receptor.ActualLRPStateRunning},
				}
				fakeReceptorClient.ActualLRPsReturns(actualLrps, nil)
			})

			It("returns a list of alphabetically sorted examined apps", func() {
				appList, err := appExaminer.ListApps()

				Expect(err).ToNot(HaveOccurred())
				Expect(len(appList)).To(Equal(3))

				process1 := appList[0]
				Expect(process1.ProcessGuid).To(Equal("process1"))
				Expect(process1.DesiredInstances).To(Equal(2))
				Expect(process1.ActualRunningInstances).To(Equal(1))
				Expect(process1.DiskMB).To(Equal(256))
				Expect(process1.MemoryMB).To(Equal(100))
				Expect(process1.Routes).To(Equal([]string{"happy", "joy"}))

				process2 := appList[1]
				Expect(process2.ProcessGuid).To(Equal("process2"))
				Expect(process2.DesiredInstances).To(Equal(0))
				Expect(process2.ActualRunningInstances).To(Equal(1))
				Expect(process2.DiskMB).To(Equal(564))
				Expect(process2.MemoryMB).To(Equal(200))
				Expect(process2.Routes).To(Equal([]string{"ren", "stimpy"}))

				process3 := appList[2]
				Expect(process3.ProcessGuid).To(Equal("process3"))
				Expect(process3.DesiredInstances).To(Equal(0))
				Expect(process3.ActualRunningInstances).To(Equal(1))
				Expect(process3.DiskMB).To(Equal(0))
				Expect(process3.MemoryMB).To(Equal(0))
				Expect(process3.Routes).To(BeEmpty())
			})
		})

		Context("with the receptor returning errors", func() {
			It("returns errors from fetching the ActualLRPs", func() {
				fakeReceptorClient.ActualLRPsReturns(nil, errors.New("Receptor is Running."))
				_, err := appExaminer.ListApps()

				Expect(err).To(HaveOccurred())
			})

			It("returns errors from from fetching the DesiredLRPs", func() {
				fakeReceptorClient.DesiredLRPsReturns(nil, errors.New("You should go catch it."))
				_, err := appExaminer.ListApps()

				Expect(err).To(HaveOccurred())
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
					receptor.CellResponse{CellID: "Cell-1"},
					receptor.CellResponse{CellID: "Cell-2"},
					receptor.CellResponse{CellID: "Cell-3"},
				}
				fakeReceptorClient.CellsReturns(cells, nil)
			})

			It("returns a list of alphabetically sorted examined cells", func() {
				cellList, err := appExaminer.ListCells()

				Expect(err).ToNot(HaveOccurred())
				Expect(len(cellList)).To(Equal(3))

				cell1 := cellList[0]
				Expect(cell1.CellID).To(Equal("Cell-1"))
				Expect(cell1.RunningInstances).To(Equal(2))
				Expect(cell1.ClaimedInstances).To(Equal(0))

				cell2 := cellList[1]
				Expect(cell2.CellID).To(Equal("Cell-2"))
				Expect(cell2.RunningInstances).To(Equal(1))
				Expect(cell2.ClaimedInstances).To(Equal(1))

				cell3 := cellList[2]
				Expect(cell3.CellID).To(Equal("Cell-3"))
				Expect(cell3.RunningInstances).To(Equal(0))
				Expect(cell3.ClaimedInstances).To(Equal(0))
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
				Expect(len(cellList)).To(Equal(2))

				cell0 := cellList[0]
				Expect(cell0.CellID).To(Equal("Cell-0"))
				Expect(cell0.Missing).To(Equal(true))
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

				Expect(err).ToNot(HaveOccurred())
				Expect(len(cellList)).To(Equal(0))
			})
		})

		Context("with the receptor returning errors", func() {
			It("returns errors from from fetching the Cells", func() {
				fakeReceptorClient.CellsReturns(nil, errors.New("You should go catch it."))
				_, err := appExaminer.ListCells()

				Expect(err).To(HaveOccurred())
			})

			It("returns errors from fetching the ActualLRPs", func() {
				fakeReceptorClient.ActualLRPsReturns(nil, errors.New("Receptor is Running."))
				_, err := appExaminer.ListCells()

				Expect(err).To(HaveOccurred())
			})

		})
	})
})
