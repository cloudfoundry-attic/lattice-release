package command_factory

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type AppExaminerCommandFactory struct {
	appExaminerCommand *appExaminerCommand
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, output *output.Output) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{&appExaminerCommand{appExaminer, output}}
}

func (commandFactory *AppExaminerCommandFactory) MakeListAppCommand() cli.Command {

	var startCommand = cli.Command{
		Name:        "list",
		ShortName:   "li",
		Description: "List all apps on lattice.",
		Usage:       "ltc list",
		Action:      commandFactory.appExaminerCommand.listApps,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

type appExaminerCommand struct {
	appExaminer app_examiner.AppExaminer
	output      *output.Output
}

func (cmd *appExaminerCommand) listApps(context *cli.Context) {
	appList, err := cmd.appExaminer.ListApps()
	if err != nil {
		cmd.output.Say("Error listing apps: " + err.Error())

	} else if len(appList) == 0 {
		cmd.output.Say("No apps to display.")
		return
	}

	w := &tabwriter.Writer{}
	w.Init(cmd.output, 10+len(colors.NoColor("")), 8, 1, '\t', 0)

	header := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMb"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
	fmt.Fprintln(w, header)

	for _, appInfo := range appList {
		routes := strings.Join(appInfo.Routes, " ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold(appInfo.ProcessGuid), colorInstances(appInfo), colors.NoColor(strconv.Itoa(appInfo.DiskMB)), colors.NoColor(strconv.Itoa(appInfo.MemoryMB)), colors.Cyan(routes))
	}
	w.Flush()
}

func colorInstances(appInfo app_examiner.AppInfo) string {
	instances := fmt.Sprintf("%d/%d", appInfo.ActualRunningInstances, appInfo.DesiredInstances)
	if appInfo.ActualRunningInstances == appInfo.DesiredInstances {
		return colors.Green(instances)
	} else if appInfo.ActualRunningInstances == 0 {
		return colors.Red(instances)
	}

	return colors.Yellow(instances)
}
