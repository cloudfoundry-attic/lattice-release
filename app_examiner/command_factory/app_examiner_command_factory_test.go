package command_factory_test

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"
	"github.com/dajulia3/cli"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/fake_app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/output/cursor"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
)

var _ = Describe("CommandFactory", func() {

	var (
		appExaminer  *fake_app_examiner.FakeAppExaminer
		outputBuffer *gbytes.Buffer
		timeProvider *faketimeprovider.FakeTimeProvider
		osSignalChan chan os.Signal
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		osSignalChan = make(chan os.Signal, 1)
		timeProvider = faketimeprovider.New(time.Now())
	})

	Describe("ListAppsCommand", func() {
		var listAppsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer), timeProvider, osSignalChan)
			listAppsCommand = commandFactory.MakeListAppCommand()
		})

		It("displays all the existing apps, making sure output spacing is correct", func() {
			listApps := []app_examiner.AppInfo{
				app_examiner.AppInfo{ProcessGuid: "process1", DesiredInstances: 21, ActualRunningInstances: 0, DiskMB: 100, MemoryMB: 50, Routes: []string{"alldaylong.com"}},
				app_examiner.AppInfo{ProcessGuid: "process2", DesiredInstances: 8, ActualRunningInstances: 9, DiskMB: 400, MemoryMB: 30, Routes: []string{"never.io"}},
				app_examiner.AppInfo{ProcessGuid: "process3", DesiredInstances: 5, ActualRunningInstances: 5, DiskMB: 600, MemoryMB: 90, Routes: []string{"allthetime.com", "herewego.org"}},
				app_examiner.AppInfo{ProcessGuid: "process4", DesiredInstances: 0, ActualRunningInstances: 0, DiskMB: 10, MemoryMB: 10, Routes: []string{}},
			}
			appExaminer.ListAppsReturns(listApps, nil)

			expectedOutput := gbytes.NewBuffer()
			w := new(tabwriter.Writer)
			w.Init(expectedOutput, 10+colors.ColorCodeLength, 8, 1, '\t', 0)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMb"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process1"), colors.Red("0/21"), colors.NoColor("100"), colors.NoColor("50"), colors.Cyan("alldaylong.com"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process2"), colors.Yellow("9/8"), colors.NoColor("400"), colors.NoColor("30"), colors.Cyan("never.io"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process3"), colors.Green("5/5"), colors.NoColor("600"), colors.NoColor("90"), colors.Cyan("allthetime.com herewego.org"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process4"), colors.Green("0/0"), colors.NoColor("10"), colors.NoColor("10"), colors.Cyan(""))
			w.Flush()

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer.Contents()).To(Equal(expectedOutput.Contents()))
		})

		It("alerts the user if there are no apps", func() {
			listApps := []app_examiner.AppInfo{}
			appExaminer.ListAppsReturns(listApps, nil)

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("No apps to display."))
		})

		It("alerts the user if fetching the list returns an error", func() {
			listApps := []app_examiner.AppInfo{}
			appExaminer.ListAppsReturns(listApps, errors.New("The list was lost"))

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Error listing apps: The list was lost"))
		})
	})

	Describe("VisualizeCommand", func() {
		var visualizeCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer), timeProvider, osSignalChan)
			visualizeCommand = commandFactory.MakeVisualizeCommand()
		})

		It("displays a visualization of cells", func() {
			listCells := []app_examiner.CellInfo{
				app_examiner.CellInfo{CellID: "cell-1", RunningInstances: 3, ClaimedInstances: 2},
				app_examiner.CellInfo{CellID: "cell-2", RunningInstances: 2, ClaimedInstances: 1},
				app_examiner.CellInfo{CellID: "cell-3", RunningInstances: 0, ClaimedInstances: 0},
			}
			appExaminer.ListCellsReturns(listCells, nil)

			test_helpers.ExecuteCommandWithArgs(visualizeCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("Distribution\n")))
			Expect(outputBuffer).To(test_helpers.Say("cell-1: " + colors.Green("•••") + colors.Yellow("••") + cursor.ClearToEndOfLine() + "\n"))
			Expect(outputBuffer).To(test_helpers.Say("cell-2: " + colors.Green("••") + colors.Yellow("•") + cursor.ClearToEndOfLine() + "\n"))
			Expect(outputBuffer).To(test_helpers.Say("cell-3: "))
			Expect(outputBuffer).To(test_helpers.SayNewLine())
		})

		It("alerts the user if fetching the cells returns an error", func() {
			appExaminer.ListCellsReturns(nil, errors.New("The list was lost"))

			test_helpers.ExecuteCommandWithArgs(visualizeCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Error visualizing: The list was lost"))
		})

		Context("When a rate flag is provided", func() {
			var closeChan chan struct{}

			It("dynamically displays the visualization", func() {
				setNumberOfRunningInstances := func(count int) {
					appExaminer.ListCellsReturns([]app_examiner.CellInfo{app_examiner.CellInfo{CellID: "cell-0", RunningInstances: count}, app_examiner.CellInfo{CellID: "cell-1", RunningInstances: count, Missing: true}}, nil)
				}

				setNumberOfRunningInstances(0)

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"-rate", "2s"})

				Eventually(outputBuffer).Should(test_helpers.Say("cell-0: " + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-1" + colors.Red("[MISSING]") + ": " + cursor.ClearToEndOfLine() + "\n"))

				setNumberOfRunningInstances(2)

				timeProvider.IncrementBySeconds(1)

				Consistently(outputBuffer).ShouldNot(test_helpers.Say("cell: \n"))

				timeProvider.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(2)))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-0: " + colors.Green("••") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-1" + colors.Red("[MISSING]") + ": " + colors.Green("••") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.ClearToEndOfDisplay()))
			})

			It("dynamically displays any errors", func() {
				appExaminer.ListCellsReturns(nil, errors.New("Spilled the Paint"))

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"-rate", "1s"})

				Eventually(outputBuffer).Should(test_helpers.Say("Error visualizing: Spilled the Paint" + cursor.ClearToEndOfLine() + "\n"))

				timeProvider.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(1)))
				Eventually(outputBuffer).Should(test_helpers.Say("Error visualizing: Spilled the Paint" + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.ClearToEndOfDisplay()))
			})

			It("Ensures the user's cursor is visible even if they interrupt ltc", func() {
				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"-rate=1s"})

				osSignalChan <- syscall.SIGHUP
				Consistently(closeChan).ShouldNot(BeClosed())

				osSignalChan <- os.Interrupt
				Eventually(closeChan).Should(BeClosed())
			})

			AfterEach(func() {
				osSignalChan <- os.Interrupt
				Eventually(closeChan).Should(BeClosed())
			})
		})

	})
})
