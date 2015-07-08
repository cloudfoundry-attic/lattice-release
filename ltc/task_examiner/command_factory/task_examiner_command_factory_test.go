package command_factory_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/fake_task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("TaskExaminer CommandFactory", func() {
	var (
		fakeTaskExaminer *fake_task_examiner.FakeTaskExaminer
		outputBuffer     *gbytes.Buffer
		terminalUI       terminal.UI
		fakeExitHandler  *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		fakeTaskExaminer = new(fake_task_examiner.FakeTaskExaminer)
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("TaskCommand", func() {
		var taskCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewTaskExaminerCommandFactory(fakeTaskExaminer, terminalUI, fakeExitHandler)
			taskCommand = commandFactory.MakeTaskCommand()
		})

		It("displays info for a pending task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid:      "boop",
				State:         "PENDING",
				CellID:        "cell-01",
				Failed:        false,
				FailureReason: "",
				Result:        "",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)

			test_helpers.ExecuteCommandWithArgs(taskCommand, []string{"boop"})

			Expect(outputBuffer).To(test_helpers.Say("Task Name"))
			Expect(outputBuffer).To(test_helpers.Say("boop"))
			Expect(outputBuffer).To(test_helpers.Say("Cell ID"))
			Expect(outputBuffer).To(test_helpers.Say("cell-01"))
			Expect(outputBuffer).To(test_helpers.Say("Status"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Yellow("PENDING")))
			Expect(outputBuffer).NotTo(test_helpers.Say("Result"))
			Expect(outputBuffer).NotTo(test_helpers.Say("Failure Reason"))

			Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
			Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("boop"))
		})

		It("displays result for a non-failed completed task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid:      "boop",
				State:         "COMPLETED",
				CellID:        "cell-01",
				Failed:        false,
				FailureReason: "",
				Result:        "some-result",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)

			test_helpers.ExecuteCommandWithArgs(taskCommand, []string{"boop"})

			Expect(outputBuffer).To(test_helpers.Say("Task Name"))
			Expect(outputBuffer).To(test_helpers.Say("boop"))
			Expect(outputBuffer).To(test_helpers.Say("Cell ID"))
			Expect(outputBuffer).To(test_helpers.Say("cell-01"))
			Expect(outputBuffer).To(test_helpers.Say("Status"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("COMPLETED")))
			Expect(outputBuffer).To(test_helpers.Say("Result"))
			Expect(outputBuffer).To(test_helpers.Say("some-result"))
			Expect(outputBuffer).NotTo(test_helpers.Say("Failure Reason"))

			Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
			Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("boop"))
		})

		It("displays failure reason for a failed task result", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid:      "boop",
				State:         "COMPLETED",
				CellID:        "cell-01",
				Failed:        true,
				FailureReason: "womp womp",
				Result:        "",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)

			test_helpers.ExecuteCommandWithArgs(taskCommand, []string{"boop"})

			Expect(outputBuffer).To(test_helpers.Say("Task Name"))
			Expect(outputBuffer).To(test_helpers.Say("boop"))
			Expect(outputBuffer).To(test_helpers.Say("Cell ID"))
			Expect(outputBuffer).To(test_helpers.Say("cell-01"))
			Expect(outputBuffer).To(test_helpers.Say("Status"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("COMPLETED")))
			Expect(outputBuffer).NotTo(test_helpers.Say("Result"))
			Expect(outputBuffer).To(test_helpers.Say("Failure Reason"))
			Expect(outputBuffer).To(test_helpers.Say("womp womp"))

			Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
			Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("boop"))
		})

		It("bails out when no task name passed", func() {
			test_helpers.ExecuteCommandWithArgs(taskCommand, []string{})

			Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
			Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(0))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})

		Context("when the task examiner returns errors", func() {
			It("prints no task found when error is tasknotfound", func() {
				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New(task_examiner.TaskNotFoundErrorMessage))

				test_helpers.ExecuteCommandWithArgs(taskCommand, []string{"boop"})

				Expect(outputBuffer).To(test_helpers.Say(colors.Red("No task 'boop' was found")))
				Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
				Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("boop"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})

			It("prints random errors", func() {
				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New("muhaha"))

				test_helpers.ExecuteCommandWithArgs(taskCommand, []string{"boop"})

				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error fetching task result: muhaha")))
				Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
				Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("boop"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})
})
