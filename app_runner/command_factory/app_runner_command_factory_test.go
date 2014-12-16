package command_factory_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"
	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/fake_app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner    *fake_app_runner.FakeAppRunner
		buffer       *gbytes.Buffer
		timeout      time.Duration = 10 * time.Second
		domain       string        = "192.168.11.11.xip.io"
		timeProvider *faketimeprovider.FakeTimeProvider
	)

	BeforeEach(func() {
		appRunner = &fake_app_runner.FakeAppRunner{}
		buffer = gbytes.NewBuffer()
	})

	Describe("startApp", func() {

		var startCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}

			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, env, timeProvider)
			startCommand = commandFactory.MakeStartAppCommand()
		})

		It("starts a Docker based  app as specified in the command via the AppRunner", func(done Done) {
			args := []string{
				"--memory-mb=12",
				"--disk-mb=12",
				"--port=3000",
				"--docker-image=docker:///fun/app",
				"--run-as-root=true",
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

			appRunner.IsAppUpReturns(true, nil)

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)
			Expect(err).NotTo(HaveOccurred())

			appRunner.IsAppUpReturns(true, nil)

			Expect(appRunner.StartDockerAppCallCount()).To(Equal(1))
			name, dockerImagePath, startCommand, appArgs, environmentVariables, privileged, memoryMB, diskMB, port := appRunner.StartDockerAppArgsForCall(0)
			Expect(name).To(Equal("cool-web-app"))
			Expect(startCommand).To(Equal("/start-me-please"))
			Expect(dockerImagePath).To(Equal("docker:///fun/app"))
			Expect(appArgs).To(Equal([]string{"AppArg0", "--appFlavor=\"purple\""}))
			Expect(environmentVariables).To(Equal(map[string]string{"TIMEZONE": "CST", "LANG": "\"Chicago English\"", "COLOR": "Blue", "UNSET": ""}))
			Expect(privileged).To(Equal(true))
			Expect(memoryMB).To(Equal(12))
			Expect(diskMB).To(Equal(12))
			Expect(port).To(Equal(3000))

			Expect(buffer).To(test_helpers.Say("Starting App: cool-web-app\n"))
			Expect(buffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
			Expect(buffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))

			close(done)
		})

		It("starts a Docker based app with sensible defaults", func(done Done) {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.IsAppUpReturns(true, nil)

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)
			Expect(err).NotTo(HaveOccurred())

			Expect(appRunner.StartDockerAppCallCount()).To(Equal(1))
			_, _, _, _, _, privileged, memoryMB, diskMB, port := appRunner.StartDockerAppArgsForCall(0)

			Expect(privileged).To(Equal(false))
			Expect(memoryMB).To(Equal(128))
			Expect(diskMB).To(Equal(1024))
			Expect(port).To(Equal(8080))

			close(done)
		})

		It("polls for the app to start", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.IsAppUpReturns(false, nil)

			go test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Eventually(buffer).Should(test_helpers.Say("Starting App: cool-web-app"))

			Expect(appRunner.IsAppUpCallCount()).To(Equal(1))
			Expect(appRunner.IsAppUpArgsForCall(0)).To(Equal("cool-web-app"))

			timeProvider.IncrementBySeconds(1)
			Eventually(buffer, 10).Should(test_helpers.Say("."))
			timeProvider.IncrementBySeconds(1)
			Eventually(buffer, 10).Should(test_helpers.Say("."))

			appRunner.IsAppUpReturns(true, nil)
			timeProvider.IncrementBySeconds(1)

			Eventually(buffer).Should(test_helpers.SayNewLine())
			Eventually(buffer).Should(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
			Eventually(buffer).Should(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))
		})

		It("alerts the user if the app does not start", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.IsAppUpReturns(false, nil)
			go test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Eventually(buffer).Should(test_helpers.Say("Starting App: cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(buffer).Should(test_helpers.SayNewLine())
			Eventually(buffer).Should(test_helpers.Say(colors.Red("cool-web-app took too long to start.")))
		})

		It("validates that the name is passed in", func(done Done) {
			args := []string{
				"--docker-image=docker:///fun/app",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

			close(done)
		})

		It("validates that the dockerImage is passed in", func(done Done) {
			args := []string{
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: Docker Image required"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

			close(done)
		})

		It("validates that the startCommand is passed in", func(done Done) {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: Start Command required"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

			close(done)
		})

		It("validates that the terminator -- is passed in", func(done Done) {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"not-the-terminator",
				"start-me-up",
			}
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: '--' Required before start command"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

			close(done)
		})

		It("validates that the full docker path is passed in", func(done Done) {
			args := []string{
				"--docker-image=fun/app",
				"cool-web-app",
				"--",
				"start-me-please",
			}
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: Docker Image should begin with: docker:///"))
			Expect(appRunner.StartDockerAppCallCount()).To(Equal(0))

			close(done)
		})

		It("outputs error messages", func(done Done) {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.StartDockerAppReturns(errors.New("Major Fault"))

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Error Starting App: Major Fault"))

			close(done)
		})
	})

	Describe("scaleApp", func() {

		var scaleCommand cli.Command
		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, []string{}, timeProvider)
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("starts a Docker based  app as specified in the command via the AppRunner", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(appRunner.ScaleAppCallCount()).To(Equal(1))

			name, instances := appRunner.ScaleAppArgsForCall(0)

			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(22))

			Expect(buffer).To(test_helpers.Say("App Scaled Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"--instances=22",
				"",
			}

			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			appRunner.ScaleAppReturns(errors.New("Major Fault"))
			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Error Scaling App to 22 instances: Major Fault"))
		})

	})

	Describe("stopApp", func() {
		var stopCommand cli.Command
		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, []string{}, timeProvider)
			stopCommand = commandFactory.MakeStopAppCommand()
		})

		It("scales an app to zero", func() {
			args := []string{
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(appRunner.ScaleAppCallCount()).To(Equal(1))

			name, instances := appRunner.ScaleAppArgsForCall(0)

			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(0))

			Expect(buffer).To(test_helpers.Say("App Scaled Successfully to 0 instances"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.ScaleAppReturns(errors.New("Major Fault"))
			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Error Scaling App to 0 instances: Major Fault"))
		})
	})

	Describe("removeApp", func() {
		var removeCommand cli.Command

		BeforeEach(func() {
			timeProvider = faketimeprovider.New(time.Now())
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, []string{}, timeProvider)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("stops a Docker based  app as specified in the command via the AppRunner", func(done Done) {
			args := []string{
				"cool",
			}

			appRunner.AppExistsReturns(true, nil)

			executeCommandDone := make(chan struct{})
			go func() {
				err := test_helpers.ExecuteCommandWithArgs(removeCommand, args) //RACE b/c prev write soon -> AppExists
				Expect(err).NotTo(HaveOccurred())
				close(executeCommandDone)
			}()

			Eventually(buffer).Should(test_helpers.Say("Removing cool"))

			Expect(appRunner.AppExistsCallCount()).To(Equal(1))
			Expect(appRunner.AppExistsArgsForCall(0)).To(Equal("cool"))

			timeProvider.IncrementBySeconds(1)
			Eventually(buffer, 10).Should(test_helpers.Say("."))
			timeProvider.IncrementBySeconds(1)
			Eventually(buffer, 10).Should(test_helpers.Say("."))

			timeProvider.IncrementBySeconds(1)
			appRunner.AppExistsReturns(false, nil) //WRITE TO APPEXISTSRETURNS

			Eventually(buffer).Should(test_helpers.SayNewLine())
			Eventually(buffer).Should(test_helpers.Say(colors.Green("Successfully Removed cool.")))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))

			<-executeCommandDone
			close(done)
		})

		It("alerts the user if the app does not remove", func() {
			appRunner.AppExistsReturns(true, nil)

			args := []string{
				"cool-web-app",
			}

			go test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(buffer).Should(test_helpers.Say("Removing cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(buffer).Should(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("alerts the user if DockerAppExists() returns an error", func() {
			appRunner.AppExistsReturns(false, errors.New("Something Bad"))

			args := []string{
				"cool-web-app",
			}

			go test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(buffer).Should(test_helpers.Say("Removing cool-web-app"))

			timeProvider.IncrementBySeconds(10)

			Eventually(buffer).Should(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			err := test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.RemoveAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.RemoveAppReturns(errors.New("Major Fault"))
			err := test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(test_helpers.Say("Error Stopping App: Major Fault"))
		})
	})
})
