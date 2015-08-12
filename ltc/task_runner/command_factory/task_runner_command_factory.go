package command_factory

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
)

const AttemptedToCreateLatticeDebugErrorMessage = reserved_app_ids.LatticeDebugLogStreamAppId + " is a reserved app name. It is used internally to stream debug logs for lattice components."

type TaskRunnerCommandFactory struct {
	taskRunner  task_runner.TaskRunner
	ui          terminal.UI
	exitHandler exit_handler.ExitHandler
}

func NewTaskRunnerCommandFactory(taskRunner task_runner.TaskRunner, ui terminal.UI, exitHandler exit_handler.ExitHandler) *TaskRunnerCommandFactory {
	return &TaskRunnerCommandFactory{
		taskRunner:  taskRunner,
		ui:          ui,
		exitHandler: exitHandler,
	}
}

func (factory *TaskRunnerCommandFactory) MakeSubmitTaskCommand() cli.Command {
	var submitTaskCommand = cli.Command{
		Name:        "submit-task",
		Aliases:     []string{"su"},
		Usage:       "Submits a task from JSON on lattice",
		Description: "ltc submit-task /path/to/json",
		Action:      factory.submitTask,
	}

	return submitTaskCommand
}

func (factory *TaskRunnerCommandFactory) MakeDeleteTaskCommand() cli.Command {
	var taskDeleteCommand = cli.Command{
		Name:        "delete-task",
		Aliases:     []string{"dt"},
		Usage:       "Deletes the given task",
		Description: "ltc delete-task TASK_NAME",
		Action:      factory.deleteTask,
		Flags:       []cli.Flag{},
	}
	return taskDeleteCommand
}

func (factory *TaskRunnerCommandFactory) MakeCancelTaskCommand() cli.Command {
	var taskDeleteCommand = cli.Command{
		Name:        "cancel-task",
		Aliases:     []string{"ct"},
		Usage:       "Cancels the given task",
		Description: "ltc cancel-task TASK_NAME",
		Action:      factory.cancelTask,
		Flags:       []cli.Flag{},
	}
	return taskDeleteCommand
}

func (factory *TaskRunnerCommandFactory) submitTask(context *cli.Context) {
	filePath := context.Args().First()
	if filePath == "" {
		factory.ui.SayLine("Path to JSON is required")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		factory.ui.SayLine("Error reading file: " + err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	taskName, err := factory.taskRunner.SubmitTask(jsonBytes)
	if err != nil {
		factory.ui.SayLine(fmt.Sprintf("Error submitting %s: %s", taskName, err))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}
	factory.ui.SayLine(colors.Green("Successfully submitted " + taskName))
}

func (factory *TaskRunnerCommandFactory) deleteTask(context *cli.Context) {
	taskGuid := context.Args().First()
	if taskGuid == "" {
		factory.ui.SayIncorrectUsage("Please input a valid TASK_GUID")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	if err := factory.taskRunner.DeleteTask(taskGuid); err != nil {
		factory.ui.SayLine(fmt.Sprintf(colors.Red("Error deleting %s: %s"), taskGuid, err.Error()))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}
	factory.ui.SayLine(colors.Green("OK"))
}

func (factory *TaskRunnerCommandFactory) cancelTask(context *cli.Context) {
	taskGuid := context.Args().First()
	if taskGuid == "" {
		factory.ui.SayIncorrectUsage("Please input a valid TASK_GUID")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	if err := factory.taskRunner.CancelTask(taskGuid); err != nil {
		factory.ui.SayLine(fmt.Sprintf(colors.Red("Error cancelling %s: %s"), taskGuid, err.Error()))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}
	factory.ui.SayLine(colors.Green("OK"))
}
