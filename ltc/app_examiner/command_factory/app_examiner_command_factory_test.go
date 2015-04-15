package command_factory_test

import (
	"errors"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/cursor"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
)

const TerminalEsc = "\033["

var _ = Describe("CommandFactory", func() {

	var (
		appExaminer  *fake_app_examiner.FakeAppExaminer
		outputBuffer *gbytes.Buffer
		terminalUI   terminal.UI
		clock        *fakeclock.FakeClock
		osSignalChan chan os.Signal
		exitHandler  *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		osSignalChan = make(chan os.Signal, 1)
		clock = fakeclock.NewFakeClock(time.Now())
		exitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("ListAppsCommand", func() {
		var listAppsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, terminalUI, clock, exitHandler)
			listAppsCommand = commandFactory.MakeListAppCommand()
		})

		It("displays all the existing apps, making sure output spacing is correct", func() {
			listApps := []app_examiner.AppInfo{
				app_examiner.AppInfo{ProcessGuid: "process1", DesiredInstances: 21, ActualRunningInstances: 0, DiskMB: 100, MemoryMB: 50, Ports: []uint16{54321}, Routes: route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"alldaylong.com"}, Port: 54321}}},
				app_examiner.AppInfo{ProcessGuid: "process2", DesiredInstances: 8, ActualRunningInstances: 9, DiskMB: 400, MemoryMB: 30, Ports: []uint16{1234}, Routes: route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"never.io"}, Port: 1234}}},
				app_examiner.AppInfo{ProcessGuid: "process3", DesiredInstances: 5, ActualRunningInstances: 5, DiskMB: 600, MemoryMB: 90, Ports: []uint16{1234}, Routes: route_helpers.AppRoutes{route_helpers.AppRoute{Hostnames: []string{"allthetime.com", "herewego.org"}, Port: 1234}}},
				app_examiner.AppInfo{ProcessGuid: "process4", DesiredInstances: 0, ActualRunningInstances: 0, DiskMB: 10, MemoryMB: 10, Routes: route_helpers.AppRoutes{}},
			}
			appExaminer.ListAppsReturns(listApps, nil)

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("App Name")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("Instances")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("DiskMB")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("MemoryMB")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("Route")))

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("process1")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("0/21")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("100")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("50")))
			Expect(outputBuffer).To(test_helpers.Say("alldaylong.com => 54321"))

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("process2")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Yellow("9/8")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("400")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("30")))
			Expect(outputBuffer).To(test_helpers.Say("never.io => 1234"))

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("process3")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("5/5")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("600")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("90")))
			Expect(outputBuffer).To(test_helpers.Say("allthetime.com, herewego.org => 1234"))

			Expect(outputBuffer).To(test_helpers.Say(colors.Bold("process4")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("0/0")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("10")))
			Expect(outputBuffer).To(test_helpers.Say(colors.NoColor("10")))

		})

		It("alerts the user if there are no apps", func() {
			listApps := []app_examiner.AppInfo{}
			appExaminer.ListAppsReturns(listApps, nil)

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("No apps to display."))
		})

		Context("when the app examiner returns an error", func() {
			It("alerts the user fetching the list returns an error", func() {
				listApps := []app_examiner.AppInfo{}
				appExaminer.ListAppsReturns(listApps, errors.New("The list was lost"))

				test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Error listing apps: The list was lost"))
			})
		})
	})

	Describe("VisualizeCommand", func() {
		var visualizeCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, terminalUI, clock, exitHandler)
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

		Context("when the app examiner returns an error", func() {
			It("alerts the user fetching the cells returns an error", func() {
				appExaminer.ListCellsReturns(nil, errors.New("The list was lost"))

				test_helpers.ExecuteCommandWithArgs(visualizeCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Error visualizing: The list was lost"))
			})
		})
/*
		Context("When a rate flag is provided", func() {
			var closeChan chan struct{}

			AfterEach(func() {
				go exitHandler.Exit(exit_codes.SigInt)
				Eventually(closeChan).Should(BeClosed())
			})

			It("dynamically displays the visualization", func() {
				setNumberOfRunningInstances := func(count int) {
					appExaminer.ListCellsReturns([]app_examiner.CellInfo{app_examiner.CellInfo{CellID: "cell-0", RunningInstances: count}, app_examiner.CellInfo{CellID: "cell-1", RunningInstances: count, Missing: true}}, nil)
				}

				setNumberOfRunningInstances(0)

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"--rate", "2s"})

				Eventually(outputBuffer).Should(test_helpers.Say("cell-0: " + colors.Red("empty") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-1" + colors.Red("[MISSING]") + ": " + cursor.ClearToEndOfLine() + "\n"))

				setNumberOfRunningInstances(2)

				clock.IncrementBySeconds(1)

				Consistently(outputBuffer).ShouldNot(test_helpers.Say("cell: \n")) // TODO: how would this happen

				clock.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(2)))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-0: " + colors.Green("••") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("cell-1" + colors.Red("[MISSING]") + ": " + colors.Green("••") + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.ClearToEndOfDisplay()))
			})

			It("dynamically displays any errors", func() {
				appExaminer.ListCellsReturns(nil, errors.New("Spilled the Paint"))

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"--rate", "1s"})

				Eventually(outputBuffer).Should(test_helpers.Say("Error visualizing: Spilled the Paint" + cursor.ClearToEndOfLine() + "\n"))

				clock.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(1)))
				Eventually(outputBuffer).Should(test_helpers.Say("Error visualizing: Spilled the Paint" + cursor.ClearToEndOfLine() + "\n"))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.ClearToEndOfDisplay()))
			})

			It("ensures the user's cursor is visible even if they interrupt ltc", func() {
				closeChan = test_helpers.AsyncExecuteCommandWithArgs(visualizeCommand, []string{"--rate=1s"})

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))

				exitHandler.Exit(exit_codes.SigInt)

				Eventually(closeChan).Should(BeClosed())
				Expect(outputBuffer).Should(test_helpers.Say(cursor.Show()))
			})
		})
*/
	})

	Describe("StatusCommand", func() {
		var (
			statusCommand cli.Command
			sampleAppInfo app_examiner.AppInfo
		)

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, terminalUI, clock, exitHandler)
			statusCommand = commandFactory.MakeStatusCommand()

			sampleAppInfo = app_examiner.AppInfo{
				ProcessGuid:            "wompy-app",
				DesiredInstances:       12,
				ActualRunningInstances: 1,
				EnvironmentVariables: []app_examiner.EnvironmentVariable{
					app_examiner.EnvironmentVariable{Name: "WOMPY_APP_PASSWORD", Value: "seekreet pass"},
					app_examiner.EnvironmentVariable{Name: "WOMPY_APP_USERNAME", Value: "mrbigglesworth54"},
				},
				StartTimeout: 600,
				DiskMB:       2048,
				MemoryMB:     256,
				CPUWeight:    100,
				Ports:        []uint16{8887, 9000},
				Routes: route_helpers.AppRoutes{
					route_helpers.AppRoute{
						Port:      9090,
						Hostnames: []string{"route-me.my-fun-domain.com"},
					},
					route_helpers.AppRoute{
						Port:      8080,
						Hostnames: []string{"wompy-app.my-fun-domain.com", "cranky-app.my-fun-domain.com"},
					},
				},
				LogGuid:    "a9s8dfa99023r",
				LogSource:  "wompy-app-logz",
				Annotation: "I love this app. So wompy.",
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
						CrashCount:     2,
					},
					app_examiner.InstanceInfo{
						Index:      5,
						State:      "CRASHED",
						CrashCount: 7,
					},
				},
			}
		})

		It("emits a pretty representation of the DesiredLRP", func() {
			appExaminer.AppStatusReturns(sampleAppInfo, nil)

			test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"wompy-app"})

			Expect(appExaminer.AppStatusCallCount()).To(Equal(1))
			Expect(appExaminer.AppStatusArgsForCall(0)).To(Equal("wompy-app"))

			Expect(outputBuffer).To(test_helpers.Say("wompy-app"))

			Expect(outputBuffer).To(test_helpers.Say("Instances"))
			Expect(outputBuffer).To(test_helpers.Say("1/12"))

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
			Expect(outputBuffer).To(test_helpers.Say("wompy-app.my-fun-domain.com => 8080"))
			Expect(outputBuffer).To(test_helpers.Say("cranky-app.my-fun-domain.com => 8080"))
			Expect(outputBuffer).To(test_helpers.Say("route-me.my-fun-domain.com => 9090"))

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

			Expect(outputBuffer).To(test_helpers.Say("Instance 4"))
			Expect(outputBuffer).To(test_helpers.Say("UNCLAIMED"))

			Expect(outputBuffer).NotTo(test_helpers.Say("InstanceGuid"))
			Expect(outputBuffer).To(test_helpers.Say("Placement Error"))
			Expect(outputBuffer).To(test_helpers.Say("insufficient resources."))
			Expect(outputBuffer).To(test_helpers.Say("Crash Count"))
			Expect(outputBuffer).To(test_helpers.Say("2"))

			Expect(outputBuffer).To(test_helpers.Say("Instance 5"))
			Expect(outputBuffer).To(test_helpers.Say("CRASHED"))

			Expect(outputBuffer).NotTo(test_helpers.Say("InstanceGuid"))
			Expect(outputBuffer).To(test_helpers.Say("Crash Count"))
			Expect(outputBuffer).To(test_helpers.Say("7"))

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

		Context("when the --summary flag is passed", func() {
			It("prints the instance info in summary mode", func() {
				appExaminer.AppStatusReturns(sampleAppInfo, nil)

				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"wompy-app", "--summary"})

				Expect(appExaminer.AppStatusCallCount()).To(Equal(1))
				Expect(appExaminer.AppStatusArgsForCall(0)).To(Equal("wompy-app"))

				Expect(outputBuffer).To(test_helpers.Say("Instance"))
				Expect(outputBuffer).To(test_helpers.Say("State"))
				Expect(outputBuffer).To(test_helpers.Say("Crashes"))
				Expect(outputBuffer).To(test_helpers.Say("Since"))

				Expect(outputBuffer).To(test_helpers.Say("3"))
				Expect(outputBuffer).To(test_helpers.Say("RUNNING"))
				Expect(outputBuffer).To(test_helpers.Say("0"))
				prettyTimestamp := time.Unix(0, 401120627*1e9).Format(command_factory.TimestampDisplayLayout)
				Expect(outputBuffer).To(test_helpers.Say(prettyTimestamp))

				Expect(outputBuffer).To(test_helpers.Say("4"))
				Expect(outputBuffer).To(test_helpers.Say("UNCLAIMED"))
				Expect(outputBuffer).To(test_helpers.Say("2"))

				Expect(outputBuffer).To(test_helpers.Say("5"))
				Expect(outputBuffer).To(test_helpers.Say("CRASHED"))
				Expect(outputBuffer).To(test_helpers.Say("7"))

			})
		})

		Context("when a rate flag is passed", func() {

			var closeChan chan struct{}

			AfterEach(func() {
				go exitHandler.Exit(exit_codes.SigInt)
				Eventually(closeChan).Should(BeClosed())

				_, err := fmt.Print(cursor.Show())
				Expect(err).ToNot(HaveOccurred())
			})

			It("refreshes for the designated time", func() {
				appExaminer.AppStatusReturns(sampleAppInfo, nil)

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(statusCommand, []string{"wompy-app", "--rate", "2s"})

				Eventually(outputBuffer).Should(test_helpers.Say("wompy-app"))
				prettyTimestamp := time.Unix(0, 401120627*1e9).Format(command_factory.TimestampDisplayLayout)
				Eventually(outputBuffer).Should(test_helpers.Say(prettyTimestamp))

				clock.IncrementBySeconds(1)

				Consistently(outputBuffer).ShouldNot(test_helpers.Say("wompy-app"))

				refreshAppInfo := app_examiner.AppInfo{
					ProcessGuid:            "wompy-app",
					DesiredInstances:       1,
					ActualRunningInstances: 1,
					ActualInstances: []app_examiner.InstanceInfo{
						app_examiner.InstanceInfo{
							InstanceGuid: "a0s9f-u9a8sf-aasdioasdjoi",
							Index:        1,
							State:        "RUNNING",
							Since:        405234567 * 1e9,
						},
					},
				}
				appExaminer.AppStatusReturns(refreshAppInfo, nil)

				clock.IncrementBySeconds(1)

				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))
				Eventually(outputBuffer).Should(test_helpers.Say(cursor.Up(24)))
				Eventually(outputBuffer).Should(test_helpers.Say("wompy-app"))
				prettyTimestamp = time.Unix(0, 405234567*1e9).Format(command_factory.TimestampDisplayLayout)
				Eventually(outputBuffer).Should(test_helpers.Say(prettyTimestamp))
			})

			It("dynamically displays any errors", func() {
				appExaminer.AppStatusReturns(sampleAppInfo, nil)

				closeChan = test_helpers.AsyncExecuteCommandWithArgs(statusCommand, []string{"wompy-app", "--rate", "1s"})

				Eventually(outputBuffer).Should(test_helpers.Say("wompy-app"))
				Expect(outputBuffer).ToNot(test_helpers.Say("Error getting status"))

				appExaminer.AppStatusReturns(app_examiner.AppInfo{}, errors.New("error fetching status"))

				clock.IncrementBySeconds(1)

				Eventually(closeChan).Should(BeClosed())

				Expect(outputBuffer).ToNot(test_helpers.Say(TerminalEsc + "\\d+A"))
				Expect(outputBuffer).To(test_helpers.Say("Error getting status: error fetching status"))
				Expect(outputBuffer).To(test_helpers.Say(cursor.Show()))
			})

			Context("when the user interrupts ltc status with ctrl-c", func() {
				It("ensures the user's cursor is still visible", func() {
					closeChan = test_helpers.AsyncExecuteCommandWithArgs(statusCommand, []string{"wompy-app", "--rate=1s"})

					Eventually(outputBuffer).Should(test_helpers.Say(cursor.Hide()))

					exitHandler.Exit(exit_codes.SigInt)

					Eventually(closeChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.Say(cursor.Show()))
				})
			})
		})

		Context("when annotation is empty", func() {
			It("omits annotation from the output", func() {
				appExaminer.AppStatusReturns(app_examiner.AppInfo{ProcessGuid: "jumpy-app"}, nil)

				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"jumpy-app"})

				Expect(outputBuffer).NotTo(test_helpers.Say("Annotation"))
			})
		})

		Context("when no app name is specified", func() {
			It("prints usage information", func() {
				test_helpers.ExecuteCommandWithArgs(statusCommand, []string{})
				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
			})
		})

		It("prints any errors from app examiner", func() {
			appExaminer.AppStatusReturns(app_examiner.AppInfo{}, errors.New("You want the status?? ...YOU CAN'T HANDLE THE STATUS!!!"))

			test_helpers.ExecuteCommandWithArgs(statusCommand, []string{"zany-app"})

			Expect(outputBuffer).To(test_helpers.Say("You want the status?? ...YOU CAN'T HANDLE THE STATUS!!!"))
		})

	})
})
