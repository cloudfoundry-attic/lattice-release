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
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/presentation"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/cursor"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock"
)

const TimestampDisplayLayout = "2006-01-02 15:04:05 (MST)"

// IntSlice attaches the methods of sort.Interface to []uint16, sorting in increasing order.
type UInt16Slice []uint16

func (p UInt16Slice) Len() int           { return len(p) }
func (p UInt16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p UInt16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type AppExaminerCommandFactory struct {
	appExaminer app_examiner.AppExaminer
	ui          terminal.UI
	clock       clock.Clock
	exitHandler exit_handler.ExitHandler
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, ui terminal.UI, clock clock.Clock, exitHandler exit_handler.ExitHandler) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{appExaminer, ui, clock, exitHandler}
}

func (factory *AppExaminerCommandFactory) MakeListAppCommand() cli.Command {

	var listCommand = cli.Command{
		Name:        "list",
		Aliases:     []string{"li"},
		Usage:       "Lists applications running on lattice",
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
	}

	var visualizeCommand = cli.Command{
		Name:        "visualize",
		Aliases:     []string{"vz"},
		Usage:       "Shows a visualization of the workload distribution across the lattice cells",
		Description: "ltc visualize [-r=DELAY]",
		Action:      factory.visualizeCells,
		Flags:       visualizeFlags,
	}

	return visualizeCommand
}

func (factory *AppExaminerCommandFactory) MakeStatusCommand() cli.Command {
	return cli.Command{
		Name:        "status",
		Aliases:     []string{"st"},
		Usage:       "Shows details about a running app on lattice",
		Description: "ltc status APP_NAME",
		Action:      factory.appStatus,
		Flags:       []cli.Flag{},
	}
}

func (factory *AppExaminerCommandFactory) listApps(context *cli.Context) {
	appList, err := factory.appExaminer.ListApps()
	if err != nil {
		factory.ui.Say("Error listing apps: " + err.Error())
		return
	} else if len(appList) == 0 {
		factory.ui.Say("No apps to display.")
		return
	}

	w := &tabwriter.Writer{}
	w.Init(factory.ui, 10+colors.ColorCodeLength, 8, 1, '\t', 0)

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

	w.Flush()
}

func printHorizontalRule(w io.Writer, pattern string) {
	header := strings.Repeat(pattern, 80) + "\n"
	fmt.Fprintf(w, header)
}

func (factory *AppExaminerCommandFactory) appStatus(context *cli.Context) {
	if len(context.Args()) < 1 {
		factory.ui.IncorrectUsage("App Name required")
		return
	}

	appName := context.Args()[0]
	appInfo, err := factory.appExaminer.AppStatus(appName)

	if err != nil {
		factory.ui.Say(err.Error())
		return
	}

	minColumnWidth := 13
	w := tabwriter.NewWriter(factory.ui, minColumnWidth, 8, 1, '\t', 0)

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

func printInstanceInfo(w io.Writer, headingPrefix string, actualInstances []app_examiner.InstanceInfo) {
	instanceBar := func(index, state string) {
		fmt.Fprintf(w, "%sInstance %s  [%s]\n", headingPrefix, index, state)
		printHorizontalRule(w, "-")
	}

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

			fmt.Fprintf(w, "%s\t%s\n", "Since", fmt.Sprint(time.Unix(0, instance.Since).Format(TimestampDisplayLayout)))

		} else if instance.State != "CRASHED" {
			fmt.Fprintf(w, "%s\t%s\n", "Placement Error", instance.PlacementError)
		}
		fmt.Fprintf(w, "%s \t%d \n", "Crash Count", instance.CrashCount)
		printHorizontalRule(w, "-")
	}
}

func (factory *AppExaminerCommandFactory) visualizeCells(context *cli.Context) {
	rate := context.Duration("rate")

	factory.ui.Say(colors.Bold("Distribution\n"))
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
		factory.ui.NewLine()
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
		factory.ui.NewLine()
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
