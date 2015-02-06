package command_factory_test

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/fake_app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler/exit_codes"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler/fake_exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/output/cursor"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("CommandFactory", func() {

	var (
		appExaminer  *fake_app_examiner.FakeAppExaminer
		outputBuffer *gbytes.Buffer
		clock        *fakeclock.FakeClock
		osSignalChan chan os.Signal
		exitHandler  *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		osSignalChan = make(chan os.Signal, 1)
		clock = fakeclock.NewFakeClock(time.Now())
		exitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("ListAppsCommand", func() {
		var listAppsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer), clock, exitHandler)
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
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMB"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
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
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer), clock, exitHandler)
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
			Expect(outputBuffer).To(test_helpers.Say("cell-3: " + colors.Red("empty")))
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

				Eventually(outputBuffer).Should(test_helpers.Say("cell-0: " + colors.Red("empty") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-1" + colors.Red("[MISSING]") + ": " + cursor.ClearToEndOfLine() + "\n"))

				setNumberOfRunningInstances(2)

				clock.IncrementBySeconds(1)

				Consistently(outputBuffer).ShouldNot(test_helpers.Say("cell: \n"))

				clock.IncrementBySeconds(1)

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

				clock.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(1)))
				Eventually(outputBuffer).Should(test_helpers.Say("Error visualizing: Spilled the Paint" + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.ClearToEndOfDisplay()))
			})

			It("Ensures the user's cursor is visible even if they interrupt ltc", func() {
				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"-rate=1s"})

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))

				exitHandler.Exit(exit_codes.SigInt)

				Expect(outputBuffer).Should(test_helpers.Say(cursor.Show()))
			})

			AfterEach(func() {
				go exitHandler.Exit(exit_codes.SigInt)
				Eventually(closeChan).Should(BeClosed())
			})
		})

	})

	Describe("StatusCommand", func() {
		var statusCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer), clock, exitHandler)
			statusCommand = commandFactory.MakeStatusCommand()
		})

		It("emits a pretty representation of the DesiredLRP", func() {
			appExaminer.AppStatusReturns(
				app_examiner.AppInfo{
					ProcessGuid:            "wompy-app",
					DesiredInstances:       12,
					ActualRunningInstances: 1,
					Stack: "lucid64",
					EnvironmentVariables: []app_examiner.EnvironmentVariable{
						app_examiner.EnvironmentVariable{Name: "WOMPY_APP_PASSWORD", Value: "seekreet pass"},
						app_examiner.EnvironmentVariable{Name: "WOMPY_APP_USERNAME", Value: "mrbigglesworth54"},
					},
					StartTimeout: 600,
					DiskMB:       2048,
					MemoryMB:     256,
					CPUWeight:    100,
					Ports:        []uint32{8887, 9000},
					Routes:       []string{"wompy-app.my-fun-domain.com", "wompy-app.my-exuberant-domain.org"},
					LogGuid:      "a9s8dfa99023r",
					LogSource:    "wompy-app-logz",
					Annotation:   "I love this app. So wompy.",
					ActualInstances: []app_examiner.InstanceInfo{
						app_examiner.InstanceInfo{
							InstanceGuid: "a0s9f-u9a8sf-aasdioasdjoi",
							CellID:       "cell-12",
							Index:        3,
							Ip:           "10.85.12.100",
							Ports: []app_examiner.PortMapping{
								app_examiner.PortMapping{
									HostPort:      1234,
									ContainerPort: 3000,
								},
								app_examiner.PortMapping{
									HostPort:      5555,
									ContainerPort: 6666,
								},
							},
							State: "RUNNING",
							Since: 401120627 * 1e9,
						},
						app_examiner.InstanceInfo{
							Index:          4,
							State:          "UNCLAIMED",
							PlacementError: "insufficient resources.",
						},
					},
				}, nil)

			test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"wompy-app"})

			Expect(appExaminer.AppStatusCallCount()).To(Equal(1))
			Expect(appExaminer.AppStatusArgsForCall(0)).To(Equal("wompy-app"))

			Expect(outputBuffer).To(test_helpers.Say("wompy-app"))

			Expect(outputBuffer).To(test_helpers.Say("Instances"))
			Expect(outputBuffer).To(test_helpers.Say("1/12"))

			Expect(outputBuffer).To(test_helpers.Say("Stack"))
			Expect(outputBuffer).To(test_helpers.Say("lucid64"))

			Expect(outputBuffer).To(test_helpers.Say("Start Timeout"))
			Expect(outputBuffer).To(test_helpers.Say("600"))

			Expect(outputBuffer).To(test_helpers.Say("DiskMB"))
			Expect(outputBuffer).To(test_helpers.Say("2048"))

			Expect(outputBuffer).To(test_helpers.Say("MemoryMB"))
			Expect(outputBuffer).To(test_helpers.Say("256"))

			Expect(outputBuffer).To(test_helpers.Say("CPUWeight"))
			Expect(outputBuffer).To(test_helpers.Say("100"))

			Expect(outputBuffer).To(test_helpers.Say("Ports"))
			Expect(outputBuffer).To(test_helpers.Say("8887"))
			Expect(outputBuffer).To(test_helpers.Say("9000"))

			Expect(outputBuffer).To(test_helpers.Say("Routes"))
			Expect(outputBuffer).To(test_helpers.Say("wompy-app.my-fun-domain.com"))
			Expect(outputBuffer).To(test_helpers.Say("wompy-app.my-exuberant-domain.org"))

			Expect(outputBuffer).To(test_helpers.Say("Annotation"))
			Expect(outputBuffer).To(test_helpers.Say("I love this app. So wompy."))

			Expect(outputBuffer).To(test_helpers.Say("Environment"))
			Expect(outputBuffer).To(test_helpers.Say(`WOMPY_APP_PASSWORD="seekreet pass"`))
			Expect(outputBuffer).To(test_helpers.Say(`WOMPY_APP_USERNAME="mrbigglesworth54"`))

			Expect(outputBuffer).To(test_helpers.Say("Instance 3"))
			Expect(outputBuffer).To(test_helpers.Say("RUNNING"))

			Expect(outputBuffer).To(test_helpers.Say("InstanceGuid"))
			Expect(outputBuffer).To(test_helpers.Say("a0s9f-u9a8sf-aasdioasdjoi"))

			Expect(outputBuffer).To(test_helpers.Say("Cell ID"))
			Expect(outputBuffer).To(test_helpers.Say("cell-12"))

			Expect(outputBuffer).To(test_helpers.Say("Ip"))
			Expect(outputBuffer).To(test_helpers.Say("10.85.12.100"))

			Expect(outputBuffer).To(test_helpers.Say("Port Mapping"))
			Expect(outputBuffer).To(test_helpers.Say("1234:3000"))
			Expect(outputBuffer).To(test_helpers.Say("5555:6666"))

			Expect(outputBuffer).To(test_helpers.Say("Since"))

			prettyTimestamp := time.Unix(0, 401120627*1e9).Format(command_factory.TimestampDisplayLayout)
			Expect(outputBuffer).To(test_helpers.Say(prettyTimestamp))

			Expect(outputBuffer).To(test_helpers.Say("Placement Error"))
			Expect(outputBuffer).To(test_helpers.Say("insufficient resources."))
		})

		Context("when there is a placement error on an actualLRP", func() {
			It("Displays UNCLAIMED in red, and outputs only the placement error", func() {
				appExaminer.AppStatusReturns(
					app_examiner.AppInfo{
						ActualInstances: []app_examiner.InstanceInfo{
							app_examiner.InstanceInfo{
								Index:          7,
								State:          "UNCLAIMED",
								PlacementError: "insufficient resources.",
							},
						},
					}, nil)

				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"swanky-app"})

				Expect(outputBuffer).To(test_helpers.Say("Instance 7"))
				Expect(outputBuffer).To(test_helpers.Say("UNCLAIMED"))

				Expect(outputBuffer).ToNot(test_helpers.Say("InstanceGuid"))
				Expect(outputBuffer).ToNot(test_helpers.Say("Cell ID"))
				Expect(outputBuffer).ToNot(test_helpers.Say("Ip"))
				Expect(outputBuffer).ToNot(test_helpers.Say("Port Mapping"))
				Expect(outputBuffer).ToNot(test_helpers.Say("Since"))

				Expect(outputBuffer).To(test_helpers.Say("Placement Error"))
				Expect(outputBuffer).To(test_helpers.Say("insufficient resources."))

			})
		})

		Context("When no appName is specified", func() {
			It("Prints usage information", func() {
				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{})
				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
			})
		})

		It("prints any errors from app examiner", func() {
			appExaminer.AppStatusReturns(app_examiner.AppInfo{}, errors.New("You want the status?? ...YOU CAN'T HANDLE THE STATUS!!!"))

			test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"zany-app"})

			Expect(outputBuffer).To(test_helpers.Say("You want the status?? ...YOU CAN'T HANDLE THE STATUS!!!"))
		})

		Context("When Annotation is empty", func() {
			It("omits Annotation from the output", func() {
				appExaminer.AppStatusReturns(app_examiner.AppInfo{ProcessGuid: "jumpy-app"}, nil)

				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"jumpy-app"})

				Expect(outputBuffer).NotTo(test_helpers.Say("Annotation"))
			})
		})
	})
})
