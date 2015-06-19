package command_factory_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher/fake_docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner                     *fake_app_runner.FakeAppRunner
		appExaminer                   *fake_app_examiner.FakeAppExaminer
		outputBuffer                  *gbytes.Buffer
		terminalUI                    terminal.UI
		domain                        string = "192.168.11.11.xip.io"
		clock                         *fakeclock.FakeClock
		dockerMetadataFetcher         *fake_docker_metadata_fetcher.FakeDockerMetadataFetcher
		appRunnerCommandFactoryConfig command_factory.AppRunnerCommandFactoryConfig
		logger                        lager.Logger
		fakeTailedLogsOutputter       *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeExitHandler               *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appRunner = &fake_app_runner.FakeAppRunner{}
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		dockerMetadataFetcher = &fake_docker_metadata_fetcher.FakeDockerMetadataFetcher{}
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lager.NewLogger("ltc-test")
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("SubmitLrpCommand", func() {
		var (
			submitLrpCommand cli.Command
			tmpDir           string
			tmpFile          *os.File
			err              error
		)

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			submitLrpCommand = commandFactory.MakeSubmitLrpCommand()
		})

		Context("when the json file exists", func() {
			BeforeEach(func() {
				tmpDir = os.TempDir()
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_json")

				Expect(err).ToNot(HaveOccurred())
			})

			It("creates an app from json", func() {
				ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)
				args := []string{tmpFile.Name()}

				appRunner.SubmitLrpReturns("my-json-app", nil)

				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(appRunner.SubmitLrpCallCount()).To(Equal(1))
				Expect(appRunner.SubmitLrpArgsForCall(0)).To(Equal([]byte(`{"Value":"test value"}`)))
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("Successfully submitted my-json-app.")))
				Expect(outputBuffer).To(test_helpers.Say("To view the status of your application: ltc status my-json-app"))
			})

			It("prints an error returned by the app_runner", func() {
				args := []string{
					tmpFile.Name(),
				}
				appRunner.SubmitLrpReturns("app-that-broke", errors.New("some error"))

				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error creating app-that-broke: some error"))
				Expect(appRunner.SubmitLrpCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		It("is an error when no path is passed in", func() {
			test_helpers.ExecuteCommandWithArgs(submitLrpCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Path to JSON is required"))
			Expect(appRunner.SubmitLrpCallCount()).To(BeZero())
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})

		Context("when the file cannot be read", func() {
			It("prints an error", func() {
				args := []string{filepath.Join(tmpDir, "file-no-existy")}

				test_helpers.ExecuteCommandWithArgs(submitLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(fmt.Sprintf("Error reading file: open %s: no such file or directory", filepath.Join(tmpDir, "file-no-existy"))))
				Expect(appRunner.SubmitLrpCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
			})
		})
	})

	Describe("ScaleAppCommand", func() {

		var scaleCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("scales an with the specified number of instances", func() {
			args := []string{
				"cool-web-app",
				"22",
			}

			appExaminer.RunningAppInstancesInfoReturns(22, false, nil)

			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))

			Expect(appRunner.ScaleAppCallCount()).To(Equal(1))
			name, instances := appRunner.ScaleAppArgsForCall(0)
			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(22))
		})

		It("polls until the required number of instances are running", func() {
			args := []string{
				"cool-web-app",
				"22",
			}

			appExaminer.RunningAppInstancesInfoReturns(1, false, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
			Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appExaminer.RunningAppInstancesInfoReturns(22, false, nil)
			clock.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))
		})

		Context("when the app does not scale before the timeout elapses", func() {
			It("alerts the user the app took too long to scale", func() {
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				args := []string{
					"cool-web-app",
					"22",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))
				Expect(outputBuffer).To(test_helpers.SayNewLine())

				clock.IncrementBySeconds(120)

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Timed out waiting for the container to scale.")))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still scaling your application in the background."))
				Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
			})
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				args := []string{
					"cool-web-app",
					"22",
				}
				appRunner.ScaleAppReturns(errors.New("Major Fault"))

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error Scaling App to 22 instances: Major Fault"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{
					"",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the number of instances is passed in", func() {
				args := []string{
					"cool-web-app",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the number of instances is an integer", func() {
				args := []string{
					"cool-web-app",
					"twenty-two",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Number of Instances must be an integer"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when there is a placement error when polling for the app to scale", func() {
			It("Prints an error message and exits", func() {
				args := []string{
					"cool-web-app",
					"3",
				}

				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appExaminer.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 3 instances"))

				Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				clock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

				appExaminer.RunningAppInstancesInfoReturns(2, true, nil)
				clock.IncrementBySeconds(1)
				Eventually(commandFinishChan).Should(BeClosed())

				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))

				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
				Expect(outputBuffer).ToNot(test_helpers.Say("Timed out waiting for the container"))
			})
		})
	})

	Describe("UpdateRoutesCommand", func() {
		var updateRoutesCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			updateRoutesCommand = commandFactory.MakeUpdateRoutesCommand()
		})

		It("updates the routes", func() {
			args := []string{
				"cool-web-app",
				"8080:foo.com,9090:bar.com",
			}

			expectedRouteOverrides := app_runner.RouteOverrides{
				app_runner.RouteOverride{
					HostnamePrefix: "foo.com",
					Port:           8080,
				},
				app_runner.RouteOverride{
					HostnamePrefix: "bar.com",
					Port:           9090,
				},
			}

			test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))

			Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(1))

			name, routeOverrides := appRunner.UpdateAppRoutesArgsForCall(0)

			Expect(name).To(Equal("cool-web-app"))
			Expect(routeOverrides).To(Equal(expectedRouteOverrides))
		})

		Context("when the --no-routes flag is passed", func() {
			It("deregisters all the routes", func() {
				args := []string{
					"cool-web-app",
					"--no-routes",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(1))
				name, routeOverrides := appRunner.UpdateAppRoutesArgsForCall(0)

				Expect(name).To(Equal("cool-web-app"))
				Expect(routeOverrides).To(Equal(app_runner.RouteOverrides{}))

				Expect(outputBuffer).To(test_helpers.Say("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))
			})
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				args := []string{
					"cool-web-app",
					"8080:foo.com",
				}

				appRunner.UpdateAppRoutesReturns(errors.New("Major Fault"))
				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error updating routes: Major Fault"))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{
					"",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc update-routes APP_NAME NEW_ROUTES' or pass '--no-routes' flag."))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the routes are passed in", func() {
				args := []string{
					"cool-web-app",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc update-routes APP_NAME NEW_ROUTES' or pass '--no-routes' flag."))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("malformed route", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"woo:aahh",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("errors out when there is no colon", func() {
				args := []string{
					"cool-web-app",
					"8888",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

	})

	Describe("RemoveAppCommand", func() {
		var removeCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("removes an app", func() {
			args := []string{
				"cool",
			}

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))
		})

		It("removes multiple apps", func() {

			args := []string{
				"app1",
				"app2",
				"app3",
			}

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app1..."))
			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app2..."))
			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app3..."))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(3))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("app1"))
			Expect(appRunner.RemoveAppArgsForCall(1)).To(Equal("app2"))
			Expect(appRunner.RemoveAppArgsForCall(2)).To(Equal("app3"))
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{}

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
				Expect(appRunner.RemoveAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when the receptor returns an error", func() {
			It("outputs error messages when trying to remove the app", func() {
				args := []string{
					"cool-web-app",
				}
				appRunner.RemoveAppReturns(errors.New("Major Fault"))

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error stopping cool-web-app: Major Fault"))
				Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})

			It("outputs error messages when trying to remove the app", func() {
				args := []string{
					"app1",
					"app2",
					"app3",
				}

				appRunner.RemoveAppStub = func(name string) error {
					if name == "app2" {
						return errors.New("Major Fault")
					}
					return nil
				}

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Removing app1..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app2..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Error stopping app2: Major Fault"))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app3..."))

				Expect(appRunner.RemoveAppCallCount()).To(Equal(3))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

	})
})
