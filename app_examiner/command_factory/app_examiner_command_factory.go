package command_factory

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory/presentation"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/output/cursor"
	"github.com/pivotal-golang/clock"
)

const TimestampDisplayLayout = "2006-01-02 15:04:05 (MST)"

type AppExaminerCommandFactory struct {
	appExaminerCommand *appExaminerCommand
}

type exitHandler interface {
	OnExit(func())
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, output *output.Output, clock clock.Clock, exitHandler exitHandler) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{&appExaminerCommand{appExaminer, output, clock, exitHandler}}
}

func (commandFactory *AppExaminerCommandFactory) MakeListAppCommand() cli.Command {

	var startCommand = cli.Command{
		Name:        "list",
		ShortName:   "li",
		Description: "List all applications running on Lattice",
		Usage:       "ltc list",
		Action:      commandFactory.appExaminerCommand.listApps,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (commandFactory *AppExaminerCommandFactory) MakeVisualizeCommand() cli.Command {

	var visualizeFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "rate, r",
			Usage: "The rate at which to refresh the visualization.\n\te.g. -r=\".5s\"\n\te.g. -r=\"1000ns\"",
		},
	}

	var startCommand = cli.Command{
		Name:        "visualize",
		Description: "Visualize the workload distribution across the Lattice Cells",
		Usage:       "ltc visualize",
		Action:      commandFactory.appExaminerCommand.visualizeCells,
		Flags:       visualizeFlags,
	}

	return startCommand
}

func (commandFactory *AppExaminerCommandFactory) MakeStatusCommand() cli.Command {
	return cli.Command{
		Name:        "status",
		Description: "Displays detailed status information about the given application and its instances",
		Usage:       "ltc status APP_NAME",
		Action:      commandFactory.appExaminerCommand.appStatus,
		Flags:       []cli.Flag{},
	}
}

type appExaminerCommand struct {
	appExaminer app_examiner.AppExaminer
	output      *output.Output
	clock       clock.Clock
	exitHandler exitHandler
}

func (cmd *appExaminerCommand) listApps(context *cli.Context) {
	appList, err := cmd.appExaminer.ListApps()
	if err != nil {
		cmd.output.Say("Error listing apps: " + err.Error())
		return
	} else if len(appList) == 0 {
		cmd.output.Say("No apps to display.")
		return
	}

	w := &tabwriter.Writer{}
	w.Init(cmd.output, 10+colors.ColorCodeLength, 8, 1, '\t', 0)

	header := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMB"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
	fmt.Fprintln(w, header)

	for _, appInfo := range appList {
		routes := strings.Join(appInfo.Routes, " ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold(appInfo.ProcessGuid), colorInstances(appInfo), colors.NoColor(strconv.Itoa(appInfo.DiskMB)), colors.NoColor(strconv.Itoa(appInfo.MemoryMB)), colors.Cyan(routes))
	}
	w.Flush()
}

func printHorizontalRule(w io.Writer, pattern string) {
	header := strings.Repeat(pattern, 80) + "\n"
	fmt.Fprintf(w, header)
}

func (cmd *appExaminerCommand) appStatus(context *cli.Context) {
	if len(context.Args()) < 1 {
		cmd.output.IncorrectUsage("App Name required")
		return
	}

	appName := context.Args()[0]
	appInfo, err := cmd.appExaminer.AppStatus(appName)

	if err != nil {
		cmd.output.Say(err.Error())
		return
	}

	minColumnWidth := 13
	w := tabwriter.NewWriter(cmd.output, minColumnWidth, 8, 1, '\t', 0)

	headingPrefix := strings.Repeat(" ", minColumnWidth/2)

	titleBar := func(title string) {
		printHorizontalRule(w, "=")
		fmt.Fprintf(w, "%s%s\n", headingPrefix, title)
		printHorizontalRule(w, "-")
	}

	titleBar(colors.Bold(appName))

	printAppInfo(w, appInfo)

	fmt.Fprintln(w, "")
	printHorizontalRule(w, "=")

	printInstanceInfo(w, headingPrefix, appInfo.ActualInstances)
	w.Flush()
}

func printAppInfo(w io.Writer, appInfo app_examiner.AppInfo) {

	fmt.Fprintf(w, "%s\t%s\n", "Instances", colorInstances(appInfo))
	fmt.Fprintf(w, "%s\t%s\n", "Stack", appInfo.Stack)

	fmt.Fprintf(w, "%s\t%d\n", "Start Timeout", appInfo.StartTimeout)
	fmt.Fprintf(w, "%s\t%d\n", "DiskMB", appInfo.DiskMB)
	fmt.Fprintf(w, "%s\t%d\n", "MemoryMB", appInfo.MemoryMB)
	fmt.Fprintf(w, "%s\t%d\n", "CPUWeight", appInfo.CPUWeight)

	portStrings := make([]string, 0)
	for _, port := range appInfo.Ports {
		portStrings = append(portStrings, fmt.Sprint(port))
	}

	fmt.Fprintf(w, "%s\t%s\n", "Ports", strings.Join(portStrings, ","))
	fmt.Fprintf(w, "%s\t%s\n", "Routes", strings.Join(appInfo.Routes, " "))
	if appInfo.Annotation != "" {
		fmt.Fprintf(w, "%s\t%s\n", "Annotation", appInfo.Annotation)
	}

	printHorizontalRule(w, "-")
	var envVars string
	for _, envVar := range appInfo.EnvironmentVariables {
		envVars += envVar.Name + `="` + envVar.Value + `" ` + "\n"
	}
	fmt.Fprintf(w, "%s\n\n%s", "Environment", envVars)

}

func printInstanceInfo(w io.Writer, headingPrefix string, actualInstances []app_examiner.InstanceInfo) {
	instanceBar := func(index, state string) {
		fmt.Fprintf(w, "%sInstance %s  [%s]\n", headingPrefix, index, state)
		printHorizontalRule(w, "-")
	}

	for _, instance := range actualInstances {
		instanceBar(fmt.Sprint(instance.Index), presentation.ColorInstanceState(instance))

		if instance.PlacementError == "" {
			fmt.Fprintf(w, "%s\t%s\n", "InstanceGuid", instance.InstanceGuid)
			fmt.Fprintf(w, "%s\t%s\n", "Cell ID", instance.CellID)
			fmt.Fprintf(w, "%s\t%s\n", "Ip", instance.Ip)

			portMappingStrings := make([]string, 0)
			for _, portMapping := range instance.Ports {
				portMappingStrings = append(portMappingStrings, fmt.Sprintf("%d:%d", portMapping.HostPort, portMapping.ContainerPort))
			}
			fmt.Fprintf(w, "%s\t%s\n", "Port Mapping", strings.Join(portMappingStrings, ";"))

			fmt.Fprintf(w, "%s\t%s\n", "Since", fmt.Sprint(time.Unix(0, instance.Since).Format(TimestampDisplayLayout)))
		} else {
			fmt.Fprintf(w, "%s\t%s\n", "Placement Error:", instance.PlacementError)
		}
		printHorizontalRule(w, "-")
	}

}

func (cmd *appExaminerCommand) visualizeCells(context *cli.Context) {
	rate := context.Duration("rate")

	cmd.output.Say(colors.Bold("Distribution\n"))
	linesWritten := cmd.printDistribution()

	if rate == 0 {
		return
	}

	closeChan := make(chan bool)
	cmd.output.Say(cursor.Hide())

	cmd.exitHandler.OnExit(func() {
		closeChan <- true
		cmd.output.Say(cursor.Show())
	})

	for {
		select {
		case <-closeChan:
			return
		case <-cmd.clock.NewTimer(rate).C():
			cmd.output.Say(cursor.Up(linesWritten))
			linesWritten = cmd.printDistribution()
		}
	}
}

func (cmd *appExaminerCommand) printDistribution() int {
	defer cmd.output.Say(cursor.ClearToEndOfDisplay())

	cells, err := cmd.appExaminer.ListCells()
	if err != nil {
		cmd.output.Say("Error visualizing: " + err.Error())
		cmd.output.Say(cursor.ClearToEndOfLine())
		cmd.output.NewLine()
		return 1
	}

	for _, cell := range cells {
		cmd.output.Say(cell.CellID)
		if cell.Missing {
			cmd.output.Say(colors.Red("[MISSING]"))
		}
		cmd.output.Say(": ")

		if cell.RunningInstances == 0 && cell.ClaimedInstances == 0 && !cell.Missing {
			cmd.output.Say(colors.Red("empty"))
		} else {
			cmd.output.Say(colors.Green(strings.Repeat("•", cell.RunningInstances)))
			cmd.output.Say(colors.Yellow(strings.Repeat("•", cell.ClaimedInstances)))
		}
		cmd.output.Say(cursor.ClearToEndOfLine())
		cmd.output.NewLine()
	}

	return len(cells)
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
