package command_factory_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
    "github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
)

var _ = Describe("CommandFactory", func() {
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

    Describe("logsCommand", func() {
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

    Describe("debugLogsCommand", func(){
        It("Tails logs from the lattice-debug stream",func(){
            commandFactory := command_factory.NewLogsCommandFactory(output.New(outputBuffer), fakeTailedLogsOutputter, exitHandler)
            tailLogsCommand := commandFactory.MakeDebugLogsCommand()

            test_helpers.AsyncExecuteCommandWithArgs(tailLogsCommand, []string{})

            Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
            Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal(reserved_app_ids.LatticeDebugLogStreamAppId))
        })
    })

})
