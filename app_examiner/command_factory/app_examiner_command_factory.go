package command_factory

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/output/cursor"
)

type AppExaminerCommandFactory struct {
	appExaminerCommand *appExaminerCommand
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, output *output.Output, timeProvider timeprovider.TimeProvider, signalChan chan os.Signal) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{&appExaminerCommand{appExaminer, output, timeProvider, signalChan}}
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

func (commandFactory *AppExaminerCommandFactory) MakeVisualizeCommand() cli.Command {

	var visualizeFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "rate, r",
			Usage: "The rate in seconds at which to refresh the visualization.",
		},
	}

	var startCommand = cli.Command{
		Name:        "visualize",
		Description: "Visualize Lattice Cells",
		Usage:       "ltc visualize",
		Action:      commandFactory.appExaminerCommand.visualizeCells,
		Flags:       visualizeFlags,
	}

	return startCommand
}

type appExaminerCommand struct {
	appExaminer  app_examiner.AppExaminer
	output       *output.Output
	timeProvider timeprovider.TimeProvider
	signalChan   chan os.Signal
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

	header := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMb"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
	fmt.Fprintln(w, header)

	for _, appInfo := range appList {
		routes := strings.Join(appInfo.Routes, " ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold(appInfo.ProcessGuid), colorInstances(appInfo), colors.NoColor(strconv.Itoa(appInfo.DiskMB)), colors.NoColor(strconv.Itoa(appInfo.MemoryMB)), colors.Cyan(routes))
	}
	w.Flush()
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
	go func() {
		for {
			select {
			case <-closeChan:
				return
			case <-cmd.timeProvider.NewTimer(rate).C():
				cmd.output.Say(cursor.Up(linesWritten))
				linesWritten = cmd.printDistribution()
			}
		}
	}()

	for signal := range cmd.signalChan {
		if signal == os.Interrupt {
			closeChan <- true
			cmd.output.Say(cursor.Show())
			return

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

		cmd.output.Say(colors.Green(strings.Repeat("•", cell.RunningInstances)))
		cmd.output.Say(colors.Yellow(strings.Repeat("•", cell.ClaimedInstances)))
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
