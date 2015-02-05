package command_factory_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler/fake_exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/command_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"
)

var _ = Describe("CommandFactory", func() {
	Describe("logsCommand", func() {
		var (
			outputBuffer            *gbytes.Buffer
			fakeTailedLogsOutputter *fake_tailed_logs_outputter.FakeTailedLogsOutputter
			signalChan              chan os.Signal
			exitHandler             exit_handler.ExitHandler
		)

		BeforeEach(func() {
			outputBuffer = gbytes.NewBuffer()
			fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
			signalChan = make(chan os.Signal)
			exitHandler = &fake_exit_handler.FakeExitHandler{}
		})

		It("Tails logs", func() {
			args := []string{
				"my-app-guid",
			}

			commandFactory := command_factory.NewLogsCommandFactory(output.New(outputBuffer), fakeTailedLogsOutputter, exitHandler)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			test_helpers.AsyncExecuteCommandWithArgs(tailLogsCommand, args)

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("my-app-guid"))

		})

		It("Handles invalid appguids", func() {
			args := []string{}

			commandFactory := command_factory.NewLogsCommandFactory(output.New(outputBuffer), fakeTailedLogsOutputter, exitHandler)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			test_helpers.ExecuteCommandWithArgs(tailLogsCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage"))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(0))

		})

	})
})
