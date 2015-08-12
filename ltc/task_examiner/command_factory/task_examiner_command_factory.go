package command_factory

import (
	"fmt"
	"text/tabwriter"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
)

type TaskExaminerCommandFactory struct {
	taskExaminer task_examiner.TaskExaminer
	ui           terminal.UI
	exitHandler  exit_handler.ExitHandler
}

func NewTaskExaminerCommandFactory(taskExaminer task_examiner.TaskExaminer, ui terminal.UI, exitHandler exit_handler.ExitHandler) *TaskExaminerCommandFactory {
	return &TaskExaminerCommandFactory{taskExaminer, ui, exitHandler}
}

func (factory *TaskExaminerCommandFactory) MakeTaskCommand() cli.Command {
	var taskCommand = cli.Command{
		Name:        "task",
		Aliases:     []string{"tk"},
		Usage:       "Displays the status of a given task",
		Description: "ltc task TASK_NAME",
		Action:      factory.task,
		Flags:       []cli.Flag{},
	}

	return taskCommand
}

func (factory *TaskExaminerCommandFactory) task(context *cli.Context) {
	taskName := context.Args().First()
	if taskName == "" {
		factory.ui.SayIncorrectUsage("")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	taskInfo, err := factory.taskExaminer.TaskStatus(taskName)
	if err != nil {
		if err.Error() == task_examiner.TaskNotFoundErrorMessage {
			factory.ui.SayLine(colors.Red(fmt.Sprintf("No task '%s' was found", taskName)))
			factory.exitHandler.Exit(exit_codes.CommandFailed)
			return
		}
		factory.ui.SayLine(colors.Red("Error fetching task result: " + err.Error()))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	w := tabwriter.NewWriter(factory.ui, 9, 8, 1, '\t', 0)

	fmt.Fprintf(w, "%s\t%s\n", "Task Name", taskInfo.TaskGuid)
	fmt.Fprintf(w, "%s\t%s\n", "Cell ID", taskInfo.CellID)
	if taskInfo.State == "PENDING" || taskInfo.State == "CLAIMED" || taskInfo.State == "RUNNING" {
		fmt.Fprintf(w, "%s\t%s\n", "Status", colors.Yellow(taskInfo.State))
	} else if (taskInfo.State == "COMPLETED" || taskInfo.State == "RESOLVING") && !taskInfo.Failed {
		fmt.Fprintf(w, "%s\t%s\n", "Status", colors.Green(taskInfo.State))
		fmt.Fprintf(w, "%s\t%s\n", "Result", taskInfo.Result)
	} else if taskInfo.Failed {
		fmt.Fprintf(w, "%s\t%s\n", "Status", colors.Red(taskInfo.State))
		fmt.Fprintf(w, "%s\t%s\n", "Failure Reason", taskInfo.FailureReason)
	}

	w.Flush()
}
