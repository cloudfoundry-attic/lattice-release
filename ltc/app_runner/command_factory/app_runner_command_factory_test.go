package command_factory_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("AppRunner CommandFactory", func() {
	var (
		fakeAppRunner                 *fake_app_runner.FakeAppRunner
		fakeAppExaminer               *fake_app_examiner.FakeAppExaminer
		outputBuffer                  *gbytes.Buffer
		terminalUI                    terminal.UI
		fakeClock                     *fakeclock.FakeClock
		fakeTailedLogsOutputter       *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeExitHandler               *fake_exit_handler.FakeExitHandler
		appRunnerCommandFactoryConfig command_factory.AppRunnerCommandFactoryConfig
	)

	BeforeEach(func() {
		fakeAppRunner = &fake_app_runner.FakeAppRunner{}
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeClock = fakeclock.NewFakeClock(time.Now())
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("helper methods", func() {
		var factory *command_factory.AppRunnerCommandFactory

		BeforeEach(func() {
			appRunnerCommandFactoryConfig := command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				UI:          terminalUI,
				ExitHandler: fakeExitHandler,
				Env:         []string{"AAAAA=1", "AAA=2", "BBB=3"},
			}

			factory = command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
		})

		Describe("BuildEnvironment", func() {
			It("grabs values from the environment when not in its args", func() {
				envVars := factory.BuildEnvironment([]string{"AAAAA", "CCC=4"})

				aaaaaVar, found := envVars["AAAAA"]
				Expect(found).To(BeTrue())
				Expect(aaaaaVar).To(Equal("1"))
				cccVar, found := envVars["CCC"]
				Expect(found).To(BeTrue())
				Expect(cccVar).To(Equal("4"))
			})

			It("only uses exact key matches when grabbing from the environment", func() {
				envVars := factory.BuildEnvironment([]string{"AAA"})

				aaaVar, found := envVars["AAA"]
				Expect(found).To(BeTrue())
				Expect(aaaVar).To(Equal("2"))
				aaaaaVar, found := envVars["AAAAA"]
				Expect(found).To(BeFalse())
				Expect(aaaaaVar).To(BeEmpty())
			})
		})

		Describe("ParseTcpRoutes", func() {
			It("parses delimited tcp routes into the TcpRoutes struct", func() {
				tcpRoutes, err := factory.ParseTcpRoutes("50000:7777,50001:5222")
				Expect(err).NotTo(HaveOccurred())
				Expect(tcpRoutes).To(ContainExactly(app_runner.TcpRoutes{
					{ExternalPort: 50000, Port: 7777},
					{ExternalPort: 50001, Port: 5222},
				}))
			})

			Context("when a malformed tcp routes is passed", func() {
				It("errors out when the container port is not an int", func() {
					_, err := factory.ParseTcpRoutes("50000:woo")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the tcp route is incomplete", func() {
					_, err := factory.ParseTcpRoutes("5222,50000")
					Expect(err).To(MatchError(command_factory.MalformedTcpRouteErrorMessage))
				})
			})

			Context("when invalid port is passed in tcp routes", func() {
				It("errors out when the container port is a negative number", func() {
					_, err := factory.ParseTcpRoutes("50000:-1")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the container port is bigger than 65535", func() {
					_, err := factory.ParseTcpRoutes("50000:65536")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the external port is a negative number", func() {
					_, err := factory.ParseTcpRoutes("-1:6379")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the external port is bigger than 65535", func() {
					_, err := factory.ParseTcpRoutes("65536:6379")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})
			})

			Context("when a reserved external port is passed", func() {
				It("errors out", func() {
					_, err := factory.ParseTcpRoutes("7777:1234")
					Expect(err).To(MatchError("Port 7777 is reserved by Lattice.\nSee: http://lattice.cf/docs/troubleshooting#what-external-ports-are-unavailable-to-bind-as-tcp-routes"))
				})
			})
		})

		Describe("ParseRouteOverrides", func() {
			Context("when valid http route is passed", func() {
				It("returns a valid RouteOverrides", func() {
					routeOverrides, err := factory.ParseRouteOverrides("foo.com:8080, bar.com:8181")
					Expect(err).NotTo(HaveOccurred())
					Expect(routeOverrides).To(ContainExactly(
						app_runner.RouteOverrides{
							{HostnamePrefix: "foo.com", Port: 8080},
							{HostnamePrefix: "bar.com", Port: 8181},
						},
					))
				})
			})

			Context("when a malformed http route is passed", func() {
				It("errors out when the container port is not an int", func() {
					_, err := factory.ParseRouteOverrides("foo:bar")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the http route is incomplete", func() {
					_, err := factory.ParseRouteOverrides("foo.com,bar.com")
					Expect(err).To(MatchError(command_factory.MalformedRouteErrorMessage))
				})
			})

			Context("when invalid port is passed in https routes", func() {
				It("errors out when the container port is a negative number", func() {
					_, err := factory.ParseRouteOverrides("foo.com:-1")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the container port is bigger than 65535", func() {
					_, err := factory.ParseRouteOverrides("foo.com:65536")
					Expect(err).To(MatchError(command_factory.InvalidPortErrorMessage))
				})

				It("errors out when the host name is empty", func() {
					_, err := factory.ParseRouteOverrides(":6379")
					Expect(err).To(MatchError(command_factory.MalformedRouteErrorMessage))
				})
			})

		})
	})

	Describe("SubmitLrpCommand", func() {
		var submitLrpCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				UI:          terminalUI,
				ExitHandler: fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			submitLrpCommand = commandFactory.MakeSubmitLrpCommand()
		})

		Context("when the json file exists", func() {
			var (
				tmpDir  string
				tmpFile *os.File
			)
			BeforeEach(func() {
				tmpDir = os.TempDir()
				var err error
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_json")
				Expect(err).NotTo(HaveOccurred())

				Expect(ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)).To(Succeed())
			})

			It("creates an app from json", func() {
				fakeAppRunner.SubmitLrpReturns("my-json-app", nil)

				args := []string{tmpFile.Name()}
				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("Successfully submitted my-json-app.")))
				Expect(outputBuffer).To(test_helpers.SayLine("To view the status of your application: ltc status my-json-app"))
				Expect(fakeAppRunner.SubmitLrpCallCount()).To(Equal(1))
				Expect(fakeAppRunner.SubmitLrpArgsForCall(0)).To(Equal([]byte(`{"Value":"test value"}`)))
			})

			It("prints an error returned by the app_runner", func() {
				fakeAppRunner.SubmitLrpReturns("app-that-broke", errors.New("some error"))
				args := []string{tmpFile.Name()}
				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Error creating app-that-broke: some error"))
				Expect(fakeAppRunner.SubmitLrpCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		It("is an error when no path is passed in", func() {
			test_helpers.ExecuteCommandWithArgs(submitLrpCommand, []string{})

			Expect(outputBuffer).To(test_helpers.SayLine("Path to JSON is required"))
			Expect(fakeAppRunner.SubmitLrpCallCount()).To(BeZero())
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})

		Context("when the file cannot be read", func() {
			It("prints an error", func() {
				args := []string{"file-no-existy"}
				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine(fmt.Sprintf("Error reading file: open %s: no such file or directory", "file-no-existy")))
				Expect(fakeAppRunner.SubmitLrpCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
			})
		})
	})

	Describe("ScaleAppCommand", func() {
		var scaleCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				AppExaminer: fakeAppExaminer,
				UI:          terminalUI,
				Clock:       fakeClock,
				ExitHandler: fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("scales an with the specified number of instances", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(22, false, nil)

			args := []string{
				"cool-web-app",
				"22",
			}
			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).To(test_helpers.SayLine("Scaling cool-web-app to 22 instances"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("App Scaled Successfully")))

			Expect(fakeAppRunner.ScaleAppCallCount()).To(Equal(1))
			name, instances := fakeAppRunner.ScaleAppArgsForCall(0)
			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(22))
		})

		It("polls until the required number of instances are running", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)

			args := []string{
				"cool-web-app",
				"22",
			}
			doneChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.SayLine("Scaling cool-web-app to 22 instances"))

			Expect(fakeAppExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
			Expect(fakeAppExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

			fakeClock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			fakeClock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			fakeAppExaminer.RunningAppInstancesInfoReturns(22, false, nil)
			fakeClock.IncrementBySeconds(1)

			Eventually(doneChan, 3).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("App Scaled Successfully")))
		})

		Context("when the app does not scale before the timeout elapses", func() {
			It("alerts the user the app took too long to scale", func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)

				args := []string{
					"cool-web-app",
					"22",
				}
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.SayLine("Scaling cool-web-app to 22 instances"))

				fakeClock.IncrementBySeconds(120)

				Eventually(doneChan, 3).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayLine(colors.Red("Timed out waiting for the container to scale.")))
				Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still scaling your application in the background."))
				Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
			})
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				fakeAppRunner.ScaleAppReturns(errors.New("Major Fault"))

				args := []string{
					"cool-web-app",
					"22",
				}
				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Error Scaling App to 22 instances: Major Fault"))
				Expect(fakeAppRunner.ScaleAppCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{""}
				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(fakeAppRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the number of instances is passed in", func() {
				args := []string{"cool-web-app"}
				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(fakeAppRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the number of instances is an integer", func() {
				args := []string{
					"cool-web-app",
					"twenty-two",
				}
				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: Number of Instances must be an integer"))
				Expect(fakeAppRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when there is a placement error when polling for the app to scale", func() {
			It("prints an error message and exits", func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(0, false, nil)

				args := []string{
					"cool-web-app",
					"3",
				}
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.SayLine("Scaling cool-web-app to 3 instances"))

				Expect(fakeAppExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(fakeAppExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				fakeClock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

				fakeAppExaminer.RunningAppInstancesInfoReturns(2, true, nil)
				fakeClock.IncrementBySeconds(1)
				Eventually(doneChan, 3).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayLine(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
				Expect(outputBuffer).NotTo(test_helpers.Say("Timed out waiting for the container"))

				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))
			})
		})
	})

	Describe("UpdateCommand", func() {
		var updateCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				UI:          terminalUI,
				ExitHandler: fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			updateCommand = commandFactory.MakeUpdateCommand()
		})

		Context("when only http routes are passed", func() {
			It("updates the http routes and removes any tcp routes", func() {
				expectedRouteOverrides := app_runner.RouteOverrides{
					{HostnamePrefix: "foo.com", Port: 8080},
					{HostnamePrefix: "bar.com", Port: 9090},
				}

				args := []string{
					"cool-web-app",
					"--http-routes=foo.com:8080,bar.com:9090",
				}
				test_helpers.ExecuteCommandWithArgs(updateCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))
				Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(1))
				updateAppParams := fakeAppRunner.UpdateAppArgsForCall(0)
				Expect(updateAppParams.Name).To(Equal("cool-web-app"))
				Expect(updateAppParams.RouteOverrides).To(Equal(expectedRouteOverrides))
				Expect(updateAppParams.TcpRoutes).To(BeNil())
				Expect(updateAppParams.NoRoutes).To(BeFalse())
			})
		})

		Context("when only tcp routes are passed", func() {
			It("updates the tcp routes and removes the http routes", func() {
				expectedTcpRoutes := app_runner.TcpRoutes{
					{ExternalPort: 50000, Port: 8080},
					{ExternalPort: 51000, Port: 8181},
				}

				args := []string{
					"cool-web-app",
					"--tcp-routes=50000:8080,51000:8181",
				}
				test_helpers.ExecuteCommandWithArgs(updateCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))
				Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(1))
				updateAppParams := fakeAppRunner.UpdateAppArgsForCall(0)
				Expect(updateAppParams.Name).To(Equal("cool-web-app"))
				Expect(updateAppParams.TcpRoutes).To(Equal(expectedTcpRoutes))
				Expect(updateAppParams.RouteOverrides).To(BeNil())
				Expect(updateAppParams.NoRoutes).To(BeFalse())

			})
		})

		Context("when both http and tcp routes are passed", func() {
			It("updates both the http and tcp routes", func() {
				expectedRouteOverrides := app_runner.RouteOverrides{
					{HostnamePrefix: "foo.com", Port: 8080},
					{HostnamePrefix: "bar.com", Port: 9090},
				}
				expectedTcpRoutes := app_runner.TcpRoutes{
					{ExternalPort: 50000, Port: 5222},
					{ExternalPort: 51000, Port: 6379},
				}

				args := []string{
					"cool-web-app",
					"--http-routes=foo.com:8080,bar.com:9090",
					"--tcp-routes=50000:5222,51000:6379",
				}
				test_helpers.ExecuteCommandWithArgs(updateCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))
				Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(1))
				updateAppParams := fakeAppRunner.UpdateAppArgsForCall(0)
				Expect(updateAppParams.Name).To(Equal("cool-web-app"))
				Expect(updateAppParams.RouteOverrides).To(Equal(expectedRouteOverrides))
				Expect(updateAppParams.TcpRoutes).To(Equal(expectedTcpRoutes))
				Expect(updateAppParams.NoRoutes).To(BeFalse())
			})
		})

		Context("invalid syntax", func() {
			Context("when no app name is passed", func() {
				Context("when no other arguments are passed", func() {
					It("prints usage message", func() {
						test_helpers.ExecuteCommandWithArgs(updateCommand, []string{})

						Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
					})
				})

				Context("when http routes are passed", func() {
					It("returns usage message", func() {
						args := []string{"--http-routes=foo.com:8080,bar.com:9090"}
						test_helpers.ExecuteCommandWithArgs(updateCommand, args)

						Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
					})
				})

				Context("when tcp routes are passed", func() {
					It("returns usage message", func() {
						args := []string{"--tcp-routes=50000:5222,51000:6379"}
						test_helpers.ExecuteCommandWithArgs(updateCommand, args)

						Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
					})
				})

				Context("when both http and tcp routes are passed", func() {
					It("returns usage message", func() {
						args := []string{"--tcp-routes=50000:5222,51000:6379", "--http-routes=foo.com:8080,bar.com:9090"}
						test_helpers.ExecuteCommandWithArgs(updateCommand, args)

						Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
					})
				})

				Context("when no routes are passed", func() {
					It("returns usage message", func() {
						args := []string{"--no-routes"}
						test_helpers.ExecuteCommandWithArgs(updateCommand, args)

						Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
					})
				})
			})

			Context("when no arguments but app name are passed", func() {
				It("returns usage message", func() {
					args := []string{"cool-web-app"}
					test_helpers.ExecuteCommandWithArgs(updateCommand, args)

					Expect(outputBuffer).To(test_helpers.SayLine("Please enter 'ltc update APP_NAME' followed by at least one of: '--no-routes', '--http-routes' or '--tcp-routes' flag."))
				})
			})
		})

		Context("when the --no-routes flag is passed", func() {
			It("deregisters all the routes", func() {
				args := []string{
					"cool-web-app",
					"--no-routes",
				}
				test_helpers.ExecuteCommandWithArgs(updateCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))
				Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(1))
				updateAppParams := fakeAppRunner.UpdateAppArgsForCall(0)
				Expect(updateAppParams.Name).To(Equal("cool-web-app"))
				Expect(updateAppParams.NoRoutes).To(BeTrue())
				Expect(updateAppParams.TcpRoutes).To(BeNil())
				Expect(updateAppParams.RouteOverrides).To(BeNil())
			})
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				fakeAppRunner.UpdateAppReturns(errors.New("Major Fault"))

				args := []string{
					"cool-web-app",
					"--http-routes=foo.com:8080",
				}
				test_helpers.ExecuteCommandWithArgs(updateCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Error updating application: Major Fault"))
				Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("malformed route", func() {
			Context("when the http route is malformed", func() {
				It("returns invalid syntax error", func() {
					args := []string{
						"cool-web-app",
						"--http-routes=woo:aahh",
					}
					test_helpers.ExecuteCommandWithArgs(updateCommand, args)

					Expect(outputBuffer).To(test_helpers.SayLine(command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
			})

			Context("when the tcp route is malformed", func() {
				It("returns invalid syntax error", func() {
					args := []string{
						"cool-web-app",
						"--tcp-routes=-1:5222",
					}
					test_helpers.ExecuteCommandWithArgs(updateCommand, args)

					Expect(outputBuffer).To(test_helpers.SayLine(command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.UpdateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
			})
		})
	})

	Describe("RemoveAppCommand", func() {
		var removeCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				UI:          terminalUI,
				ExitHandler: fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("removes an app", func() {
			args := []string{"cool"}
			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.SayLine("Removing cool..."))
			Expect(fakeAppRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(fakeAppRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))
		})

		It("removes multiple apps", func() {
			args := []string{
				"app1",
				"app2",
				"app3",
			}
			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.SayLine("Removing app1..."))
			Expect(outputBuffer).To(test_helpers.SayLine("Removing app2..."))
			Expect(outputBuffer).To(test_helpers.SayLine("Removing app3..."))

			Expect(fakeAppRunner.RemoveAppCallCount()).To(Equal(3))
			Expect(fakeAppRunner.RemoveAppArgsForCall(0)).To(Equal("app1"))
			Expect(fakeAppRunner.RemoveAppArgsForCall(1)).To(Equal("app2"))
			Expect(fakeAppRunner.RemoveAppArgsForCall(2)).To(Equal("app3"))
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{}
				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: App Name required"))
				Expect(fakeAppRunner.RemoveAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when the receptor returns an error", func() {
			It("outputs error messages when trying to remove the app", func() {
				fakeAppRunner.RemoveAppReturns(errors.New("Major Fault"))

				args := []string{"cool-web-app"}
				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Error stopping cool-web-app: Major Fault"))
				Expect(fakeAppRunner.RemoveAppCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})

			It("outputs error messages when trying to remove the app", func() {
				fakeAppRunner.RemoveAppStub = func(name string) error {
					if name == "app2" {
						return errors.New("Major Fault")
					}
					return nil
				}

				args := []string{"app1", "app2", "app3"}
				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Removing app1..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app2..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Error stopping app2: Major Fault"))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app3..."))

				Expect(fakeAppRunner.RemoveAppCallCount()).To(Equal(3))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})
})
