package graphical_test

import (
	"errors"
	"time"

	"github.com/gizak/termui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/graphical"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_noaa_consumer"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
)

var _ = Describe("Graphical", func() {
	var (
		fakeReceptorClient *fake_receptor.FakeClient
		fakeNoaaConsumer   *fake_noaa_consumer.FakeNoaaConsumer
		appExaminer        app_examiner.AppExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		fakeNoaaConsumer = &fake_noaa_consumer.FakeNoaaConsumer{}
		appExaminer = app_examiner.New(fakeReceptorClient, fakeNoaaConsumer)
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
	Describe("Initialize Termui", func() {
		It("checks when termui initialization returns an error", func() {

			graphical.InitTermUI = func() error {
				return errors.New("Unable to Initialize Termui")
			}
			rate := 100 * time.Millisecond

			err := graphical.PrintDistributionChart(appExaminer, rate)
			Expect(err).To(HaveOccurred())
			//Expect(err).NotTo(HaveOccurred())
		})
		It("checks when Label initialization returns an error", func() {

			graphical.InitTermUI = termui.Init
			graphical.Label = func(string) *termui.Par {
				return nil
			}
			rate := 100 * time.Millisecond

			err := graphical.PrintDistributionChart(appExaminer, rate)
			Expect(err).To(HaveOccurred())
		})
		It("checks when MBar initialization returns an error", func() {

			graphical.InitTermUI = termui.Init
			graphical.Label = termui.NewPar
			graphical.BarGraph = func() *termui.MBarChart {
				return nil
			}
			rate := 100 * time.Millisecond

			err := graphical.PrintDistributionChart(appExaminer, rate)
			Expect(err).To(HaveOccurred())
		})

	})
})
