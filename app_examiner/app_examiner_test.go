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
			It("returns errors from fetching the AcutalLRPs", func() {
				fakeReceptorClient.ActualLRPsReturns([]receptor.ActualLRPResponse{}, errors.New("Receptor is Running."))
				_, err := appExaminer.ListApps()

				Expect(err).To(HaveOccurred())
			})

			It("returns errors from from fetching the DesiredLRPs", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, errors.New("You should go catch it."))
				_, err := appExaminer.ListApps()

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
