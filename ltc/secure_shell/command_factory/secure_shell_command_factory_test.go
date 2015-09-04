package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/command_factory/fake_secure_shell"
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
		fakeSecureShell *fake_secure_shell.FakeSecureShell
	)

	BeforeEach(func() {
		config = config_package.New(nil)
		config.SetTarget("lattice.xip.io")

		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
		fakeSecureShell = &fake_secure_shell.FakeSecureShell{}
	})

	Describe("SSHCommand", func() {
		var sshCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewSSHCommandFactory(config, terminalUI, fakeExitHandler, fakeSecureShell)
			sshCommand = commandFactory.MakeSSHCommand()
		})

		It("should ssh to instance 0 given an app name", func() {
			test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"app-name"})

			Expect(outputBuffer).To(test_helpers.SayLine("Connecting to app-name/0 at %s", config.Target()))

			Expect(fakeSecureShell.ConnectToShellCallCount()).To(Equal(1))
			appName, instanceIndex, actualConfig := fakeSecureShell.ConnectToShellArgsForCall(0)
			Expect(appName).To(Equal("app-name"))
			Expect(instanceIndex).To(Equal(0))
			Expect(actualConfig).To(Equal(config))
		})

		It("should ssh to instance index specified", func() {
			test_helpers.ExecuteCommandWithArgs(sshCommand, []string{"--instance", "2", "app-name"})

			Expect(outputBuffer).To(test_helpers.SayLine("Connecting to app-name/2 at %s", config.Target()))

			Expect(fakeSecureShell.ConnectToShellCallCount()).To(Equal(1))
			appName, instanceIndex, actualConfig := fakeSecureShell.ConnectToShellArgsForCall(0)
			Expect(appName).To(Equal("app-name"))
			Expect(instanceIndex).To(Equal(2))
			Expect(actualConfig).To(Equal(config))
		})

		Context("when not given an app name", func() {
			It("prints an error", func() {
				test_helpers.ExecuteCommandWithArgs(sshCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())

				Expect(fakeSecureShell.ConnectToShellCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

	})
})
