package command_factory

import (
	"fmt"
	"text/tabwriter"

	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
)

type TaskExaminerCommandFactory struct {
	taskExaminer task_examiner.TaskExaminer
	ui           terminal.UI
}

func NewTaskExaminerCommandFactory(taskExaminer task_examiner.TaskExaminer, ui terminal.UI) *TaskExaminerCommandFactory {
	return &TaskExaminerCommandFactory{taskExaminer, ui}
}

func (factory *TaskExaminerCommandFactory) MakeTaskCommand() cli.Command {

	var taskCommand = cli.Command{
		Name:        "task",
		Aliases:     []string{},
		Usage:       "Displays the status of a given task",
		Description: "ltc task TASK_NAME",
		Action:      factory.task,
		Flags:       []cli.Flag{},
	}

	return taskCommand
}

func (factory *TaskExaminerCommandFactory) MakeTaskDeleteCommand() cli.Command {
	var taskDeleteCommand = cli.Command{
		Name:        "delete-task",
		Aliases:     []string{"dt"},
		Usage:       "Deletes the given task",
		Description: "ltc delete-task TASK_NAME",
		Action:      factory.taskDelete,
		Flags:       []cli.Flag{},
	}
	return taskDeleteCommand
}

func (factory *TaskExaminerCommandFactory) task(context *cli.Context) {
	taskName := context.Args().First()
	if taskName == "" {
		factory.ui.SayIncorrectUsage("")
		return
	}

	taskInfo, err := factory.taskExaminer.TaskStatus(taskName)
	if err != nil {
		if err.Error() == task_examiner.TaskNotFoundErrorMessage {
			factory.ui.Say(colors.Red(fmt.Sprintf("No task '%s' was found", taskName)))
			return
		}
		factory.ui.Say(colors.Red("Error fetching task result: " + err.Error()))
		return
	}

	w := tabwriter.NewWriter(factory.ui, 9, 8, 1, '\t', 0)

	fmt.Fprintf(w, "%s\t%s\n", "Task Name", taskInfo.TaskGuid)
	fmt.Fprintf(w, "%s\t%s\n", "Cell ID", taskInfo.CellID)
	fmt.Fprintf(w, "%s\t%s\n", "Status", taskInfo.State)
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

func (factory *TaskExaminerCommandFactory) taskDelete(context *cli.Context) {
	taskGuid := context.Args().First()
	if taskGuid == "" {
		factory.ui.SayIncorrectUsage("Please input a valid TASK_GUID")
		return
	}
	factory.ui.Say("Deleting the task " + colors.Bold(taskGuid) + "\n")
	err := factory.taskExaminer.TaskDelete(taskGuid)
	if err != nil {
		factory.ui.Say("Error Deleting the task " + colors.Bold(taskGuid) + "\n")
		factory.ui.Say("Failiure Reason :" + colors.Red(err.Error()) + "\n")
		return
	}
	factory.ui.Say(colors.Green("OK"))
}
