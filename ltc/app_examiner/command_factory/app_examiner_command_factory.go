package command_factory

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/graphical"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/presentation"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/cursor"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/bytefmt"
	"github.com/pivotal-golang/clock"
)

const (
	minColumnWidth = 13
)

var (
	indentHeading = strings.Repeat(" ", minColumnWidth/2)
)

// IntSlice attaches the methods of sort.Interface to []uint16, sorting in increasing order.
type UInt16Slice []uint16

func (p UInt16Slice) Len() int           { return len(p) }
func (p UInt16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p UInt16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type AppExaminerCommandFactory struct {
	appExaminer         app_examiner.AppExaminer
	ui                  terminal.UI
	clock               clock.Clock
	exitHandler         exit_handler.ExitHandler
	graphicalVisualizer graphical.GraphicalVisualizer
	taskExaminer        task_examiner.TaskExaminer
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, ui terminal.UI, clock clock.Clock, exitHandler exit_handler.ExitHandler, graphicalVisualizer graphical.GraphicalVisualizer, taskExaminer task_examiner.TaskExaminer) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{appExaminer, ui, clock, exitHandler, graphicalVisualizer, taskExaminer}
}

func (factory *AppExaminerCommandFactory) MakeListAppCommand() cli.Command {

	var listCommand = cli.Command{
		Name:        "list",
		Aliases:     []string{"ls"},
		Usage:       "Lists applications & tasks running on lattice",
		Description: "ltc list",
		Action:      factory.listApps,
		Flags:       []cli.Flag{},
	}

	return listCommand
}

func (factory *AppExaminerCommandFactory) MakeVisualizeCommand() cli.Command {

	var visualizeFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "rate, r",
			Usage: "Visualization refresh rate (e.g., \".5s\" or \"10ms\")",
		},
		cli.BoolFlag{
			Name:  "graphical, g",
			Usage: "Visualize in a graphical screen",
		},
	}

	var visualizeCommand = cli.Command{
		Name:        "visualize",
		Aliases:     []string{"vz"},
		Usage:       "Visualizes the workload distribution across the lattice cells",
		Description: "ltc visualize [-r=DELAY] [-g]",
		Action:      factory.visualizeCells,
		Flags:       visualizeFlags,
	}

	return visualizeCommand
}

func (factory *AppExaminerCommandFactory) MakeStatusCommand() cli.Command {
	var statusFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "summary, s",
			Usage: "Summarizes the app instances",
		},
		cli.DurationFlag{
			Name:  "rate, r",
			Usage: "Status refresh rate (e.g., \".5s\" or \"10ms\")",
		},
	}

	return cli.Command{
		Name:        "status",
		Aliases:     []string{"st"},
		Usage:       "Shows details about a running app on lattice",
		Description: "ltc status APP_NAME",
		Action:      factory.appStatus,
		Flags:       statusFlags,
	}
}

func (factory *AppExaminerCommandFactory) MakeCellsCommand() cli.Command {
	return cli.Command{
		Name:    "cells",
		Aliases: []string{"ce"},
		Usage:   "Shows details about lattice cells",
		Description: `ltc cells APP_NAME
 
    Output format is:

    Cell	Zone	Memory	Apps (running/claimed)`,

		Action: factory.cells,
		Flags:  []cli.Flag{},
	}
}

func (factory *AppExaminerCommandFactory) cells(context *cli.Context) {
	cellList, err := factory.appExaminer.ListCells()
	if err != nil {
		factory.ui.SayLine(err.Error())
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	w := &tabwriter.Writer{}
	w.Init(factory.ui, 9, 8, 1, '\t', 0)

	fmt.Fprintln(w, "Cells\tZone\tMemory\tDisk\tApps")

	for _, cellInfo := range cellList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			cellInfo.CellID,
			cellInfo.Zone,
			fmt.Sprintf("%dM", cellInfo.MemoryMB),
			fmt.Sprintf("%dM", cellInfo.DiskMB),
			fmt.Sprintf("%d/%d", cellInfo.RunningInstances, cellInfo.ClaimedInstances),
		)
	}

	w.Flush()
}

func (factory *AppExaminerCommandFactory) listApps(context *cli.Context) {
	appList, err := factory.appExaminer.ListApps()
	if err == nil {
		w := &tabwriter.Writer{}
		w.Init(factory.ui, 10+colors.ColorCodeLength, 8, 1, '\t', 0)
		appTableHeader := strings.Repeat("-", 30) + "= Apps =" + strings.Repeat("-", 31)
		fmt.Fprintln(w, appTableHeader)
		if len(appList) != 0 {
			header := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMB"), colors.Bold("MemoryMB"), colors.Bold("Route"))
			fmt.Fprintln(w, header)

			for _, appInfo := range appList {
				var displayedRoute string
				if appInfo.Routes != nil && len(appInfo.Routes) > 0 {
					arbitraryPort := appInfo.Ports[0]
					displayedRoute = fmt.Sprintf("%s => %d", strings.Join(appInfo.Routes.HostnamesByPort()[arbitraryPort], ", "), arbitraryPort)
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold(appInfo.ProcessGuid), colorInstances(appInfo), colors.NoColor(strconv.Itoa(appInfo.DiskMB)), colors.NoColor(strconv.Itoa(appInfo.MemoryMB)), colors.Cyan(displayedRoute))
			}

		} else {
			fmt.Fprintf(w, "No apps to display."+"\n")
		}
		w.Flush()
	} else {
		factory.ui.SayLine("Error listing apps: " + err.Error())
		factory.exitHandler.Exit(exit_codes.CommandFailed)
	}
	taskList, err := factory.taskExaminer.ListTasks()
	if err == nil {
		wTask := &tabwriter.Writer{}
		wTask.Init(factory.ui, 10+colors.ColorCodeLength, 8, 1, '\t', 0)
		factory.ui.SayNewLine()
		taskTableHeader := strings.Repeat("-", 30) + "= Tasks =" + strings.Repeat("-", 30)
		fmt.Fprintln(wTask, taskTableHeader)
		if len(taskList) != 0 {
			taskHeader := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t", colors.Bold("Task Name"), colors.Bold("Cell ID"), colors.Bold("Status"), colors.Bold("Result"), colors.Bold("Failure Reason"))
			fmt.Fprintln(wTask, taskHeader)

			for _, taskInfo := range taskList {
				if taskInfo.CellID == "" {
					taskInfo.CellID = "N/A"
				}
				if taskInfo.Result == "" {
					taskInfo.Result = "N/A"
				}
				if taskInfo.FailureReason == "" {
					taskInfo.FailureReason = "N/A"
				}
				coloumnInfo := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t", colors.Bold(taskInfo.TaskGuid), colors.NoColor(taskInfo.CellID), colors.NoColor(taskInfo.State), colors.NoColor(taskInfo.Result), colors.NoColor(taskInfo.FailureReason))
				fmt.Fprintln(wTask, coloumnInfo)
			}

		} else {
			fmt.Fprintf(wTask, "No tasks to display.\n")
		}
		wTask.Flush()
	} else {
		factory.ui.SayLine("Error listing tasks: " + err.Error())
		factory.exitHandler.Exit(exit_codes.CommandFailed)
	}
}

func (factory *AppExaminerCommandFactory) appStatus(context *cli.Context) {

	summaryFlag := context.Bool("summary")
	rateFlag := context.Duration("rate")

	if len(context.Args()) < 1 {
		factory.ui.SayIncorrectUsage("App Name required")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	appName := context.Args()[0]

	appInfo, err := factory.appExaminer.AppStatus(appName)
	if err != nil {
		factory.ui.SayLine(err.Error())
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	factory.printAppInfo(appInfo)

	if summaryFlag || rateFlag != 0 {
		factory.printInstanceSummary(appInfo.ActualInstances)
	} else {
		factory.printInstanceInfo(appInfo.ActualInstances)
	}

	if rateFlag == 0 {
		return
	}

	linesWritten := appStatusLinesWritten(appInfo)
	closeChan := make(chan struct{})
	defer factory.ui.Say(cursor.Show())
	factory.ui.Say(cursor.Hide())

	factory.exitHandler.OnExit(func() {
		closeChan <- struct{}{}
		factory.ui.Say(cursor.Show())
	})

	for {
		select {
		case <-closeChan:
			return
		case <-factory.clock.NewTimer(rateFlag).C():
			appInfo, err = factory.appExaminer.AppStatus(appName)
			if err != nil {
				factory.ui.SayLine("Error getting status: " + err.Error())
				return
			}
			factory.ui.Say(cursor.Up(linesWritten))
			factory.printAppInfo(appInfo)
			factory.printInstanceSummary(appInfo.ActualInstances)
			linesWritten = appStatusLinesWritten(appInfo)
		}
	}
}

func (factory *AppExaminerCommandFactory) printAppInfo(appInfo app_examiner.AppInfo) {
	factory.ui.Say(cursor.ClearToEndOfDisplay())

	w := tabwriter.NewWriter(factory.ui, minColumnWidth, 8, 1, '\t', 0)

	titleBar := func(title string) {
		printHorizontalRule(w, "=")
		fmt.Fprintf(w, "%s%s\n", indentHeading, title)
		printHorizontalRule(w, "-")
	}

	titleBar(colors.Bold(appInfo.ProcessGuid))

	fmt.Fprintf(w, "%s\t%s\n", "Instances", colorInstances(appInfo))
	fmt.Fprintf(w, "%s\t%d\n", "Start Timeout", appInfo.StartTimeout)
	fmt.Fprintf(w, "%s\t%d\n", "DiskMB", appInfo.DiskMB)
	fmt.Fprintf(w, "%s\t%d\n", "MemoryMB", appInfo.MemoryMB)
	fmt.Fprintf(w, "%s\t%d\n", "CPUWeight", appInfo.CPUWeight)

	portStrings := make([]string, 0)
	for _, port := range appInfo.Ports {
		portStrings = append(portStrings, fmt.Sprint(port))
	}

	fmt.Fprintf(w, "%s\t%s\n", "Ports", strings.Join(portStrings, ","))

	printAppRoutes(w, appInfo)

	if appInfo.Annotation != "" {
		fmt.Fprintf(w, "%s\t%s\n", "Annotation", appInfo.Annotation)
	}

	printHorizontalRule(w, "-")
	var envVars string
	for _, envVar := range appInfo.EnvironmentVariables {
		envVars += envVar.Name + `="` + envVar.Value + `" ` + "\n"
	}
	fmt.Fprintf(w, "%s\n\n%s", "Environment", envVars)

	fmt.Fprintln(w, "")

	w.Flush()
}

func appStatusLinesWritten(appInfo app_examiner.AppInfo) int {
	linesWritten := 9
	for _, appRoute := range appInfo.Routes {
		linesWritten += len(appRoute.Hostnames)
	}
	if appInfo.Annotation != "" {
		linesWritten += 1
	}
	linesWritten += 3
	linesWritten += len(appInfo.EnvironmentVariables)
	linesWritten += 4
	linesWritten += len(appInfo.ActualInstances)
	return linesWritten
}

func (factory *AppExaminerCommandFactory) printInstanceSummary(actualInstances []app_examiner.InstanceInfo) {
	w := tabwriter.NewWriter(factory.ui, minColumnWidth, 8, 1, '\t', 0)

	printHorizontalRule(w, "=")
	fmt.Fprintf(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n",
		"Instance",
		colors.NoColor("State")+"    ",
		"Crashes",
		"CPU",
		"Memory",
		"Uptime"),
	)
	printHorizontalRule(w, "-")

	for _, instance := range actualInstances {
		metricsSlice := []string{"N/A", "N/A"}
		if instance.HasMetrics {
			metricsSlice = []string{
				fmt.Sprintf("%.2f%%", instance.Metrics.CpuPercentage),
				fmt.Sprintf("%s", bytefmt.ByteSize(instance.Metrics.MemoryBytes)),
			}
		}
		if instance.PlacementError == "" && instance.State != "CRASHED" {
			uptime := factory.clock.Now().Sub(time.Unix(0, instance.Since))
			roundedUptime := uptime - (uptime % time.Second)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				strconv.Itoa(instance.Index),
				presentation.PadAndColorInstanceState(instance),
				strconv.Itoa(instance.CrashCount),
				strings.Join(metricsSlice, "\t"),
				fmt.Sprint(roundedUptime),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				strconv.Itoa(instance.Index),
				presentation.PadAndColorInstanceState(instance),
				strconv.Itoa(instance.CrashCount),
				strings.Join(metricsSlice, "\t"),
				"N/A",
			)
		}
	}

	w.Flush()
}

func (factory *AppExaminerCommandFactory) printInstanceInfo(actualInstances []app_examiner.InstanceInfo) {
	w := tabwriter.NewWriter(factory.ui, minColumnWidth, 8, 1, '\t', 0)

	instanceBar := func(index, state string) {
		fmt.Fprintf(w, "%sInstance %s  [%s]\n", indentHeading, index, state)
		printHorizontalRule(w, "-")
	}

	printHorizontalRule(w, "=")

	for _, instance := range actualInstances {

		instanceBar(fmt.Sprint(instance.Index), presentation.ColorInstanceState(instance))

		if instance.PlacementError == "" && instance.State != "CRASHED" {
			fmt.Fprintf(w, "%s\t%s\n", "InstanceGuid", instance.InstanceGuid)
			fmt.Fprintf(w, "%s\t%s\n", "Cell ID", instance.CellID)
			fmt.Fprintf(w, "%s\t%s\n", "Ip", instance.Ip)

			portMappingStrings := make([]string, 0)
			for _, portMapping := range instance.Ports {
				portMappingStrings = append(portMappingStrings, fmt.Sprintf("%d:%d", portMapping.HostPort, portMapping.ContainerPort))
			}
			fmt.Fprintf(w, "%s\t%s\n", "Port Mapping", strings.Join(portMappingStrings, ";"))
			uptime := factory.clock.Now().Sub(time.Unix(0, instance.Since))
			roundedUptime := uptime - (uptime % time.Second)
			fmt.Fprintf(w, "%s\t%s\n", "Uptime", fmt.Sprint(roundedUptime))

		} else if instance.State != "CRASHED" {
			fmt.Fprintf(w, "%s\t%s\n", "Placement Error", instance.PlacementError)
		}

		fmt.Fprintf(w, "%s \t%d \n", "Crash Count", instance.CrashCount)

		if instance.HasMetrics {
			fmt.Fprintf(w, "%s \t%.2f%% \n", "CPU", instance.Metrics.CpuPercentage)
			fmt.Fprintf(w, "%s \t%s \n", "Memory", bytefmt.ByteSize(instance.Metrics.MemoryBytes))
		}
		printHorizontalRule(w, "-")
	}

	w.Flush()
}

func (factory *AppExaminerCommandFactory) visualizeCells(context *cli.Context) {
	rate := context.Duration("rate")
	graphicalFlag := context.Bool("graphical")

	if graphicalFlag {
		err := factory.graphicalVisualizer.PrintDistributionChart(rate)
		if err != nil {
			factory.ui.SayLine("Error Visualization: " + err.Error())
			factory.exitHandler.Exit(exit_codes.CommandFailed)
		}
		return
	}

	factory.ui.SayLine(colors.Bold("Distribution"))
	linesWritten := factory.printDistribution()

	if rate == 0 {
		return
	}

	closeChan := make(chan struct{})
	factory.ui.Say(cursor.Hide())

	factory.exitHandler.OnExit(func() {
		closeChan <- struct{}{}
		factory.ui.Say(cursor.Show())

	})

	for {
		select {
		case <-closeChan:
			return
		case <-factory.clock.NewTimer(rate).C():
			factory.ui.Say(cursor.Up(linesWritten))
			linesWritten = factory.printDistribution()
		}
	}
}

func (factory *AppExaminerCommandFactory) printDistribution() int {
	defer factory.ui.Say(cursor.ClearToEndOfDisplay())

	cells, err := factory.appExaminer.ListCells()
	if err != nil {
		factory.ui.Say("Error visualizing: " + err.Error())
		factory.ui.Say(cursor.ClearToEndOfLine())
		factory.ui.SayNewLine()
		return 1
	}

	for _, cell := range cells {
		factory.ui.Say(cell.CellID)
		if cell.Missing {
			factory.ui.Say(colors.Red("[MISSING]"))
		}
		factory.ui.Say(": ")

		if cell.RunningInstances == 0 && cell.ClaimedInstances == 0 && !cell.Missing {
			factory.ui.Say(colors.Red("empty"))
		} else {
			factory.ui.Say(colors.Green(strings.Repeat("•", cell.RunningInstances)))
			factory.ui.Say(colors.Yellow(strings.Repeat("•", cell.ClaimedInstances)))
		}
		factory.ui.Say(cursor.ClearToEndOfLine())
		factory.ui.SayNewLine()
	}

	return len(cells)
}

func printHorizontalRule(w io.Writer, pattern string) {
	header := strings.Repeat(pattern, 90) + "\n"
	fmt.Fprintf(w, header)
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

func printAppRoutes(w io.Writer, appInfo app_examiner.AppInfo) {
	formatRoute := func(hostname string, port uint16) string {
		return colors.Cyan(fmt.Sprintf("%s => %d", hostname, port))
	}

	routeStringsByPort := appInfo.Routes.HostnamesByPort()
	ports := make(UInt16Slice, 0, len(routeStringsByPort))
	for port, _ := range routeStringsByPort {
		ports = append(ports, port)
	}
	sort.Sort(ports)

	for portIndex, port := range ports {
		routeStrs, _ := routeStringsByPort[uint16(port)]
		for routeIndex, routeStr := range routeStrs {
			if routeIndex == 0 && portIndex == 0 {
				fmt.Fprintf(w, "%s\t%s\n", "Routes", formatRoute(routeStrs[0], port))
			} else {
				fmt.Fprintf(w, "\t%s\n", formatRoute(routeStr, port))
			}
		}
	}
}
