package command_factory_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/command_factory/mocks"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("SSH CommandFactory", func() {
	var (
		config          *config_package.Config
		outputBuffer    *gbytes.Buffer
		terminalUI      terminal.UI
		fakeExitHandler *fake_exit_handler.FakeExitHandler
		fakeAppExaminer *fake_app_examiner.FakeAppExaminer
		fakeSSH         *mocks.FakeSSH
	)

	BeforeEach(func() {
		config = config_package.New(nil)
		config.SetTarget("lattice.xip.io")

		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		fakeSSH = &mocks.FakeSSH{}
	})

	Describe("SSHCommand", func() {
		var sshCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewSSHCommandFactory(config, terminalUI, fakeExitHandler, fakeAppExaminer, fakeSSH)
			sshCommand = commandFactory.MakeSSHCommand()
		})

		Context("when connecting fails", func() {
			It("should print an error", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)
				fakeSSH.ConnectReturns(errors.New("connection failed"))

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"good-name", "-L", "mrlocalhost:1234:remotehost:5678"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error connecting to good-name/0: connection failed"))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(1))
				Expect(fakeSSH.ForwardCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Describe("port forwarding", func() {
			It("should forward a local port to a remote host and port", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "-L", "mrlocalhost:1234:remotehost:5678"})

				Expect(outputBuffer).To(test_helpers.SayLine("Forwarding mrlocalhost:1234 to remotehost:5678 via app-name/0 at %s", config.Target()))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(1))
				appName, instanceIndex, actualConfig := fakeSSH.ConnectArgsForCall(0)
				Expect(appName).To(Equal("app-name"))
				Expect(instanceIndex).To(Equal(0))
				Expect(actualConfig).To(Equal(config))

				Expect(fakeSSH.ForwardCallCount()).To(Equal(1))
				localAddr, remoteAddr := fakeSSH.ForwardArgsForCall(0)
				Expect(localAddr).To(Equal("mrlocalhost:1234"))
				Expect(remoteAddr).To(Equal("remotehost:5678"))

				Expect(fakeAppExaminer.AppStatusCallCount()).To(Equal(1))
				Expect(fakeAppExaminer.AppStatusArgsForCall(0)).To(Equal("app-name"))
			})

			Context("when the local host address is not specified", func() {
				It("should default to localhost", func() {
					fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

					test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "-L", "1234:remotehost:5678"})

					Expect(outputBuffer).To(test_helpers.SayLine("Forwarding localhost:1234 to remotehost:5678 via app-name/0 at %s", config.Target()))

					Expect(fakeSSH.ForwardCallCount()).To(Equal(1))
					localAddr, _ := fakeSSH.ForwardArgsForCall(0)
					Expect(localAddr).To(Equal("localhost:1234"))
				})
			})

			Context("when forwarding fails", func() {
				It("should print an error", func() {
					fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)
					fakeSSH.ForwardReturns(errors.New("forwarding failed"))

					test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"good-name", "-L", "mrlocalhost:1234:remotehost:5678"})

					Expect(outputBuffer).To(test_helpers.SayLine("Forwarding mrlocalhost:1234 to remotehost:5678 via good-name/0 at %s", config.Target()))
					Expect(outputBuffer).To(test_helpers.SayLine("Error connecting to good-name/0: forwarding failed"))

					Expect(fakeSSH.ForwardCallCount()).To(Equal(1))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})

			It("should reject malformed local forward specs", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "-L", "9999:localhost:1234:remotehost:5678"})
				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: -L expects [localhost:]localport:remotehost:remoteport"))

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "-L", "remotehost:5678"})
				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: -L expects [localhost:]localport:remotehost:remoteport"))

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "-L", "5678"})
				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: -L expects [localhost:]localport:remotehost:remoteport"))

				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax, exit_codes.InvalidSyntax, exit_codes.InvalidSyntax}))

				Expect(fakeSSH.ForwardCallCount()).To(Equal(0))
			})
		})

		Describe("interactive shell", func() {
			It("should ssh to instance 0 given an app name", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name"})

				Expect(outputBuffer).To(test_helpers.SayLine("Connecting to app-name/0 at %s", config.Target()))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(1))
				appName, instanceIndex, actualConfig := fakeSSH.ConnectArgsForCall(0)
				Expect(appName).To(Equal("app-name"))
				Expect(instanceIndex).To(Equal(0))
				Expect(actualConfig).To(Equal(config))

				Expect(fakeSSH.ShellCallCount()).To(Equal(1))
				command := fakeSSH.ShellArgsForCall(0)
				Expect(command).To(BeEmpty())

				Expect(fakeAppExaminer.AppStatusCallCount()).To(Equal(1))
				Expect(fakeAppExaminer.AppStatusArgsForCall(0)).To(Equal("app-name"))
			})

			It("should ssh to instance index specified", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 3}, nil)

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"--instance", "2", "app-name"})

				Expect(outputBuffer).To(test_helpers.SayLine("Connecting to app-name/2 at %s", config.Target()))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(1))
				appName, instanceIndex, actualConfig := fakeSSH.ConnectArgsForCall(0)
				Expect(appName).To(Equal("app-name"))
				Expect(instanceIndex).To(Equal(2))
				Expect(actualConfig).To(Equal(config))

				Expect(fakeSSH.ShellCallCount()).To(Equal(1))
				command := fakeSSH.ShellArgsForCall(0)
				Expect(command).To(BeEmpty())
			})

			Context("when a command is provided", func() {
				It("should run a command remotely instead of the login shell", func() {
					fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(sshCommand, []string{"app-name", "echo", "1", "2", "3"})

					Eventually(doneChan, 3).Should(BeClosed())
					Expect(outputBuffer).NotTo(test_helpers.Say("Connecting to app-name"))

					Expect(fakeSSH.ShellCallCount()).To(Equal(1))
					command := fakeSSH.ShellArgsForCall(0)
					Expect(command).To(Equal("echo 1 2 3"))
				})

				It("should support -- delimiter for args", func() {
					fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

					test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name", "--", "/bin/ls", "-l"})

					Expect(fakeSSH.ShellCallCount()).To(Equal(1))
					command := fakeSSH.ShellArgsForCall(0)
					Expect(command).To(Equal("/bin/ls -l"))
				})
			})

			Context("when opening a shell fails", func() {
				It("should print an error", func() {
					fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)
					fakeSSH.ShellReturns(errors.New("shell failed"))

					test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"good-name"})

					Expect(outputBuffer).To(test_helpers.SayLine("Connecting to good-name/0 at %s", config.Target()))
					Expect(outputBuffer).To(test_helpers.SayLine("Error connecting to good-name/0: shell failed"))

					Expect(fakeSSH.ShellCallCount()).To(Equal(1))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})
		})

		Context("when not given an app name", func() {
			It("prints an error", func() {
				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())

				Expect(fakeSSH.ConnectCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when given a non-existent app name", func() {
			It("prints an error", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{}, errors.New("no app"))

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"bad-app"})

				Expect(outputBuffer).To(test_helpers.SayLine("App bad-app not found."))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when given an invalid instance index", func() {
			It("prints an error", func() {
				fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{ActualRunningInstances: 1}, nil)

				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"good-app", "-i", "1"})

				Expect(outputBuffer).To(test_helpers.SayLine("Instance good-app/1 does not exist."))

				Expect(fakeSSH.ConnectCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})
})
