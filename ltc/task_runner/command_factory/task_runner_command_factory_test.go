package command_factory_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/fake_task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner/fake_task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("TaskRunner CommandFactory", func() {
	var (
		outputBuffer     *gbytes.Buffer
		terminalUI       terminal.UI
		fakeTaskRunner   *fake_task_runner.FakeTaskRunner
		fakeTaskExaminer *fake_task_examiner.FakeTaskExaminer
		fakeExitHandler  *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeTaskRunner = new(fake_task_runner.FakeTaskRunner)
		fakeTaskExaminer = new(fake_task_examiner.FakeTaskExaminer)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("SubmitTaskCommand", func() {
		var (
			submitTaskCommand cli.Command
			tmpFile           *os.File
			err               error
			jsonContents      []byte
		)

		BeforeEach(func() {
			commandFactory := command_factory.NewTaskRunnerCommandFactory(fakeTaskRunner, terminalUI, fakeExitHandler)
			submitTaskCommand = commandFactory.MakeSubmitTaskCommand()
		})

		Context("when the json file exists", func() {
			BeforeEach(func() {
				tmpFile, err = ioutil.TempFile("", "tmp_json")
				Expect(err).ToNot(HaveOccurred())

				jsonContents = []byte(`{"Value":"test value"}`)
				Expect(ioutil.WriteFile(tmpFile.Name(), jsonContents, 0700)).To(Succeed())
			})

			It("submits a task from json", func() {
				fakeTaskRunner.SubmitTaskReturns("some-task", nil)
				args := []string{tmpFile.Name()}

				test_helpers.ExecuteCommandWithArgs(submitTaskCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(colors.Green("Successfully submitted some-task")))
				Expect(fakeTaskRunner.SubmitTaskCallCount()).To(Equal(1))
				Expect(fakeTaskRunner.SubmitTaskArgsForCall(0)).To(Equal(jsonContents))
			})

			It("prints an error returned by the task_runner", func() {
				fakeTaskRunner.SubmitTaskReturns("some-task", errors.New("taskypoo"))
				args := []string{tmpFile.Name()}

				test_helpers.ExecuteCommandWithArgs(submitTaskCommand, args)

				Expect(fakeTaskRunner.SubmitTaskCallCount()).To(Equal(1))
				Expect(fakeTaskRunner.SubmitTaskArgsForCall(0)).To(Equal(jsonContents))

				Expect(outputBuffer).To(test_helpers.Say("Error submitting some-task: taskypoo"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})

		})

		It("is an error when no path is passed in", func() {
			test_helpers.ExecuteCommandWithArgs(submitTaskCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Path to JSON is required"))
			Expect(fakeTaskRunner.SubmitTaskCallCount()).To(BeZero())
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})

		Context("when the file cannot be read", func() {
			It("prints an error", func() {
				args := []string{filepath.Join(os.TempDir(), "file-no-existy")}

				test_helpers.ExecuteCommandWithArgs(submitTaskCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(fmt.Sprintf("Error reading file: open %s: no such file or directory", filepath.Join(os.TempDir(), "file-no-existy"))))
				Expect(fakeTaskRunner.SubmitTaskCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
			})
		})

	})

	Describe("DeleteTaskCommand", func() {
		var deleteTaskCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewTaskRunnerCommandFactory(fakeTaskRunner, terminalUI, fakeExitHandler)
			deleteTaskCommand = commandFactory.MakeDeleteTaskCommand()
		})

		It("Deletes the given task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    "COMPLETED",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)
			fakeTaskRunner.DeleteTaskReturns(nil)
			test_helpers.ExecuteCommandWithArgs(deleteTaskCommand, []string{"task-guid-1"})

			Expect(outputBuffer).To(test_helpers.Say(colors.Green("OK")))
		})

		It("returns error when fail to delete the task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    "COMPLETED",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)
			fakeTaskRunner.DeleteTaskReturns(errors.New("task in unknown state"))
			test_helpers.ExecuteCommandWithArgs(deleteTaskCommand, []string{"task-guid-1"})

			Expect(outputBuffer).To(test_helpers.Say("Error deleting task-guid-1: " + "task in unknown state"))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
		})

		It("fails with usage", func() {
			test_helpers.ExecuteCommandWithArgs(deleteTaskCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Please input a valid TASK_GUID"))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})
	})

	Describe("CancelTaskCommand", func() {
		var cancelTaskCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewTaskRunnerCommandFactory(fakeTaskRunner, terminalUI, fakeExitHandler)
			cancelTaskCommand = commandFactory.MakeCancelTaskCommand()
		})

		It("Cancels the given task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    "COMPLETED",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)
			fakeTaskRunner.CancelTaskReturns(nil)
			test_helpers.ExecuteCommandWithArgs(cancelTaskCommand, []string{"task-guid-1"})

			Expect(outputBuffer).To(test_helpers.Say(colors.Green("OK")))
		})

		It("returns error when fail to cancel the task", func() {
			taskInfo := task_examiner.TaskInfo{
				TaskGuid: "task-guid-1",
				State:    "COMPLETED",
			}
			fakeTaskExaminer.TaskStatusReturns(taskInfo, nil)
			fakeTaskRunner.CancelTaskReturns(errors.New("task in unknown state"))
			test_helpers.ExecuteCommandWithArgs(cancelTaskCommand, []string{"task-guid-1"})

			Expect(outputBuffer).To(test_helpers.Say("Error cancelling task-guid-1: " + "task in unknown state"))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
		})

		It("fails with usage", func() {
			test_helpers.ExecuteCommandWithArgs(cancelTaskCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Please input a valid TASK_GUID"))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
		})
	})
})
