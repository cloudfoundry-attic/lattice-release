package command_factory_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/fake_task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("CommandFactory", func() {
	var (
		appExaminer             *fake_app_examiner.FakeAppExaminer
		taskExaminer            *fake_task_examiner.FakeTaskExaminer
		outputBuffer            *gbytes.Buffer
		terminalUI              terminal.UI
		fakeTailedLogsOutputter *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		signalChan              chan os.Signal
		fakeExitHandler         *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		taskExaminer = &fake_task_examiner.FakeTaskExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		signalChan = make(chan os.Signal)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("LogsCommand", func() {
		var logsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewLogsCommandFactory(appExaminer, taskExaminer, terminalUI, fakeTailedLogsOutputter, fakeExitHandler)
			logsCommand = commandFactory.MakeLogsCommand()
		})

		It("tails logs", func() {
			appExaminer.AppExistsReturns(true, nil)

			doneChan := test_helpers.AsyncExecuteCommandWithArgs(logsCommand, []string{"my-app-guid"})

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("my-app-guid"))

			Consistently(doneChan).ShouldNot(BeClosed())
		})

		It("handles invalid appguids", func() {
			test_helpers.ExecuteCommandWithArgs(logsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})

		It("handles non existent application", func() {
			appExaminer.AppExistsReturns(false, nil)
			taskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New("task not found"))

			doneChan := test_helpers.AsyncExecuteCommandWithArgs(logsCommand, []string{"non_existent_app"})

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("non_existent_app"))
			Expect(outputBuffer).To(test_helpers.Say("Application or task non_existent_app not found."))
			Expect(outputBuffer).To(test_helpers.Say("Tailing logs and waiting for non_existent_app to appear..."))

			Consistently(doneChan).ShouldNot(BeClosed())
		})

		It("handles tasks", func() {
			appExaminer.AppExistsReturns(false, nil)
			taskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, nil)

			doneChan := test_helpers.AsyncExecuteCommandWithArgs(logsCommand, []string{"task-guid"})

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("task-guid"))

			Consistently(doneChan).ShouldNot(BeClosed())
		})

		Context("when the receptor returns an error", func() {
			It("displays an error and exits", func() {
				appExaminer.AppExistsReturns(false, errors.New("can't log this"))

				test_helpers.ExecuteCommandWithArgs(logsCommand, []string{"non_existent_app"})

				Expect(outputBuffer).To(test_helpers.Say("Error: can't log this"))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(BeZero())
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})

	Describe("DebugLogsCommand", func() {
		var debugLogsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewLogsCommandFactory(appExaminer, taskExaminer, terminalUI, fakeTailedLogsOutputter, fakeExitHandler)
			debugLogsCommand = commandFactory.MakeDebugLogsCommand()
		})

		It("tails logs from the lattice-debug stream", func() {
			doneChan := test_helpers.AsyncExecuteCommandWithArgs(debugLogsCommand, []string{})

			Eventually(fakeTailedLogsOutputter.OutputDebugLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputDebugLogsArgsForCall(0)).To(BeTrue())

			Consistently(doneChan).ShouldNot(BeClosed())
		})

		Context("when the --raw flag is passed", func() {
			It("tails the debug logs without pretty print", func() {
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(debugLogsCommand, []string{"--raw"})

				Eventually(fakeTailedLogsOutputter.OutputDebugLogsCallCount).Should(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputDebugLogsArgsForCall(0)).To(BeFalse())

				Consistently(doneChan).ShouldNot(BeClosed())
			})
		})
	})
})
