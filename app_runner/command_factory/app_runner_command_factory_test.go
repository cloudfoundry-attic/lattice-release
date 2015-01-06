package command_factory_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"
	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher/fake_docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/fake_app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner             *fake_app_runner.FakeAppRunner
		outputBuffer          *gbytes.Buffer
		timeout               time.Duration = 10 * time.Second
		domain                string        = "192.168.11.11.xip.io"
		timeProvider          *faketimeprovider.FakeTimeProvider
		dockerMetadataFetcher *fake_docker_metadata_fetcher.FakeDockerMetadataFetcher
	)

	BeforeEach(func() {
		appRunner = &fake_app_runner.FakeAppRunner{}
		outputBuffer = gbytes.NewBuffer()
		dockerMetadataFetcher = &fake_docker_metadata_fetcher.FakeDockerMetadataFetcher{}
	})

	Describe("StartAppCommand", func() {

		var startCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}

			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, dockerMetadataFetcher, output.New(outputBuffer), timeout, domain, env, timeProvider)
			startCommand = commandFactory.MakeStartAppCommand()
		})

		It("starts a Docker based  app as specified in the command via the AppRunner", func() {
			args := []string{
				"--memory-mb=12",
				"--disk-mb=12",
				"--port=3000",
				"--working-dir=/applications",
				"--docker-image=docker:///fun/app",
				"--run-as-root=true",
				"--instances=22",
				"--env=TIMEZONE=CST",
				"--env=LANG=\"Chicago English\"",
				"--env=COLOR",
				"--env=UNSET",
				"cool-web-app",
				"--",
				"/start-me-please",
				"AppArg0",
				"--appFlavor=\"purple\"",
			}

			appRunner.NumOfRunningAppInstancesReturns(22, nil)

			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(appRunner.StartDockerAppCallCount()).To(Equal(1))
			startDockerAppParameters := appRunner.StartDockerAppArgsForCall(0)
			Expect(startDockerAppParameters.Name).To(Equal("cool-web-app"))
			Expect(startDockerAppParameters.StartCommand).To(Equal("/start-me-please"))
			Expect(startDockerAppParameters.DockerImagePath).To(Equal("docker:///fun/app"))
			Expect(startDockerAppParameters.AppArgs).To(Equal([]string{"AppArg0", "--appFlavor=\"purple\""}))
			Expect(startDockerAppParameters.Instances).To(Equal(22))
			Expect(startDockerAppParameters.EnvironmentVariables).To(Equal(map[string]string{"TIMEZONE": "CST", "LANG": "\"Chicago English\"", "COLOR": "Blue", "UNSET": ""}))
			Expect(startDockerAppParameters.Privileged).To(Equal(true))
			Expect(startDockerAppParameters.MemoryMB).To(Equal(12))
			Expect(startDockerAppParameters.DiskMB).To(Equal(12))
			Expect(startDockerAppParameters.Port).To(Equal(3000))
			Expect(startDockerAppParameters.WorkingDir).To(Equal("/applications"))

			Expect(outputBuffer).To(test_helpers.Say("Starting App: cool-web-app\n"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))
		})

		It("starts a Docker based app with sensible defaults", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.NumOfRunningAppInstancesReturns(1, nil)

			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(appRunner.StartDockerAppCallCount()).To(Equal(1))
			startDockerAppParamters := appRunner.StartDockerAppArgsForCall(0)

			Expect(startDockerAppParamters.Privileged).To(Equal(false))
			Expect(startDockerAppParamters.MemoryMB).To(Equal(128))
			Expect(startDockerAppParamters.DiskMB).To(Equal(1024))
			Expect(startDockerAppParamters.Port).To(Equal(8080))
			Expect(startDockerAppParamters.Instances).To(Equal(1))
			Expect(startDockerAppParamters.WorkingDir).To(Equal("/"))
		})

		Context("when no start command is provided", func() {
			var args = []string{
				"--docker-image=docker:///fun-org/app",
				"cool-web-app",
			}

			BeforeEach(func() {
				appRunner.NumOfRunningAppInstancesReturns(1, nil)
			})

			It("starts a Docker app with the start command retrieved from the docker image metadata", func() {
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{WorkingDir: "/this/directory/right/here", StartCommand: []string{"/fetch-start", "arg1", "arg2"}}, nil)

				test_helpers.ExecuteCommandWithArgs(startCommand, args)

				Expect(dockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))

				repoName, tag := dockerMetadataFetcher.FetchMetadataArgsForCall(0)
				Expect(repoName).To(Equal("fun-org/app"))
				Expect(tag).To(Equal("latest"))

				Expect(appRunner.StartDockerAppCallCount()).To(Equal(1))
				startDockerAppParameters := appRunner.StartDockerAppArgsForCall(0)

				Expect(startDockerAppParameters.StartCommand).To(Equal("/fetch-start"))
				Expect(startDockerAppParameters.AppArgs).To(Equal([]string{"arg1", "arg2"}))
				Expect(startDockerAppParameters.DockerImagePath).To(Equal("docker:///fun-org/app"))
				Expect(startDockerAppParameters.WorkingDir).To(Equal("/this/directory/right/here"))

				Expect(outputBuffer).To(test_helpers.Say("No start command specified, fetching metadata from the Dockerimage...\n"))
				Expect(outputBuffer).To(test_helpers.Say("Start command is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/fetch-start arg1 arg2\n"))

				Expect(outputBuffer).To(test_helpers.Say("Working directory is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/this/directory/right/here\n"))
			})

			It("starts a Docker app with the start command retrieved from the docker image metadata", func() {
				dockerMetadataFetcher.FetchMetadataReturns(nil, errors.New("Docker Says No."))

				test_helpers.ExecuteCommandWithArgs(startCommand, args)

				Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

				Expect(outputBuffer).To(test_helpers.Say("Error Fetching metadata: Docker Says No."))
			})

			It("does not ouput the working directory if it is not set", func() {
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{"/fetch-start"}}, nil)

				test_helpers.ExecuteCommandWithArgs(startCommand, args)

				Expect(outputBuffer).ToNot(test_helpers.Say("Working directory is:"))
			})
		})

		It("polls for the app to start with correct number of instances", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"--instances=10",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.NumOfRunningAppInstancesReturns(0, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(startCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Starting App: cool-web-app"))

			Expect(appRunner.NumOfRunningAppInstancesCallCount()).To(Equal(1))
			Expect(appRunner.NumOfRunningAppInstancesArgsForCall(0)).To(Equal("cool-web-app"))

			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer, 10).Should(test_helpers.Say("."))

			appRunner.NumOfRunningAppInstancesReturns(9, nil)
			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer, 10).Should(test_helpers.Say("."))
			Expect(commandFinishChan).ShouldNot(BeClosed())

			appRunner.NumOfRunningAppInstancesReturns(10, nil)
			timeProvider.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())
			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))
		})

		It("alerts the user if the app does not start", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.NumOfRunningAppInstancesReturns(0, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(startCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Starting App: cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("cool-web-app took too long to start.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
			}

			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))
		})

		It("validates that the dockerImage is passed in", func() {
			args := []string{
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Docker Image required"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))
		})

		It("validates that the terminator -- is passed in when a start command is specified", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"not-the-terminator",
				"start-me-up",
			}
			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: '--' Required before start command"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))
		})

		It("validates that the full docker path is passed in", func() {
			args := []string{
				"--docker-image=fun/app",
				"cool-web-app",
				"--",
				"start-me-please",
			}
			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Docker Image should begin with: docker:///"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.StartDockerAppReturns(errors.New("Major Fault"))

			test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Error Starting App: Major Fault"))
		})
	})

	Describe("ScaleAppCommand", func() {

		var scaleCommand cli.Command
		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, dockerMetadataFetcher, output.New(outputBuffer), timeout, domain, []string{}, timeProvider)
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("scales an with the specified number of instances", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			appRunner.NumOfRunningAppInstancesReturns(22, nil)

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
				"--instances=22",
				"cool-web-app",
			}

			appRunner.NumOfRunningAppInstancesReturns(1, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			Expect(appRunner.NumOfRunningAppInstancesCallCount()).To(Equal(1))
			Expect(appRunner.NumOfRunningAppInstancesArgsForCall(0)).To(Equal("cool-web-app"))

			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appRunner.NumOfRunningAppInstancesReturns(22, nil)
			timeProvider.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))
		})

		It("alerts the user if the app does not scale succesfully", func() {
			appRunner.NumOfRunningAppInstancesReturns(1, nil)

			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			timeProvider.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("cool-web-app took too long to scale.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			appRunner.ScaleAppReturns(errors.New("Major Fault"))
			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Error Scaling App to 22 instances: Major Fault"))
		})

		It("returns an error if instances is not set", func() {
			args := []string{
				"cool-web-app",
			}

			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Number of Instances Required"))
			Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
		})

	})

	Describe("StopAppCommand", func() {
		var stopCommand cli.Command
		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, dockerMetadataFetcher, output.New(outputBuffer), timeout, domain, []string{}, timeProvider)
			stopCommand = commandFactory.MakeStopAppCommand()
		})

		It("scales an app to zero", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.NumOfRunningAppInstancesReturns(0, nil)

			test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 0 instances"))
			Expect(outputBuffer).To(test_helpers.Say("App Scaled Successfully"))

			Expect(appRunner.ScaleAppCallCount()).To(Equal(1))
			name, instances := appRunner.ScaleAppArgsForCall(0)
			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(0))
		})

		It("polls the app until zero instances are running", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.NumOfRunningAppInstancesReturns(1, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(stopCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 0 instances"))

			Expect(appRunner.NumOfRunningAppInstancesCallCount()).To(Equal(1))
			Expect(appRunner.NumOfRunningAppInstancesArgsForCall(0)).To(Equal("cool-web-app"))

			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appRunner.NumOfRunningAppInstancesReturns(0, nil)
			timeProvider.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say("App Scaled Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.ScaleAppReturns(errors.New("Major Fault"))
			test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Error Scaling App to 0 instances: Major Fault"))
		})
	})

	Describe("RemoveAppCommand", func() {
		var removeCommand cli.Command

		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, dockerMetadataFetcher, output.New(outputBuffer), timeout, domain, []string{}, timeProvider)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("removes a app", func() {
			args := []string{
				"cool",
			}

			appRunner.AppExistsReturns(false, nil)

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))
			Eventually(outputBuffer).Should(test_helpers.Say(colors.Green("Successfully Removed cool.")))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))
		})

		It("polls until the app is removed", func() {
			args := []string{
				"cool",
			}

			appRunner.AppExistsReturns(true, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))

			Expect(appRunner.AppExistsCallCount()).To(Equal(1))
			Expect(appRunner.AppExistsArgsForCall(0)).To(Equal("cool"))

			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			timeProvider.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appRunner.AppExistsReturns(false, nil)
			timeProvider.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Eventually(outputBuffer).Should(test_helpers.SayNewLine())
			Eventually(outputBuffer).Should(test_helpers.Say(colors.Green("Successfully Removed cool.")))
		})

		It("alerts the user if the app does not remove", func() {
			appRunner.AppExistsReturns(true, nil)

			args := []string{
				"cool-web-app",
			}

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("alerts the user if the app runner returns an error", func() {
			appRunner.AppExistsReturns(false, errors.New("Something Bad"))

			args := []string{
				"cool-web-app",
			}

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.RemoveAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.RemoveAppReturns(errors.New("Major Fault"))
			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Error Stopping App: Major Fault"))
		})
	})
})
