package command_factory

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
)

const (
	AttemptedToCreateLatticeDebugErrorMessage = reserved_app_ids.LatticeDebugLogStreamAppId + " is a reserved app name. It is used internally to stream debug logs for lattice components."
)

type TaskRunnerCommandFactory struct {
	taskRunner task_runner.TaskRunner
	ui         terminal.UI
}

func NewTaskRunnerCommandFactory(taskRunner task_runner.TaskRunner, ui terminal.UI) *TaskRunnerCommandFactory {
	return &TaskRunnerCommandFactory{
		taskRunner: taskRunner,
		ui:         ui,
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

func (factory *TaskRunnerCommandFactory) submitTask(context *cli.Context) {
	filePath := context.Args().First()
	if filePath == "" {
		factory.ui.Say("Path to JSON is required")
		return
	}

	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		factory.ui.Say("Error reading file: " + err.Error())
		return
	}

	taskName, err := factory.taskRunner.SubmitTask(jsonBytes)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error submitting %s: %s", taskName, err))
		return
	}
	factory.ui.Say(colors.Green("Successfully submitted "+taskName) + "\n")
}
