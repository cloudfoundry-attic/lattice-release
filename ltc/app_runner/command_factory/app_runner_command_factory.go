package command_factory

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type pollingAction string

const (
	InvalidPortErrorMessage          = "Invalid port specified. Ports must be a comma-delimited list of integers between 0-65535."
	MalformedRouteErrorMessage       = "Malformed route. Routes must be of the format port:route"
	MustSetMonitoredPortErrorMessage = "Must set monitored-port when specifying multiple exposed ports unless --no-monitor is set."

	DefaultPollingTimeout time.Duration = 2 * time.Minute

	pollingStart pollingAction = "start"
	pollingScale pollingAction = "scale"
)

type AppRunnerCommandFactory struct {
	appRunner             docker_app_runner.AppRunner
	appExaminer           app_examiner.AppExaminer
	ui                    terminal.UI
	dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	domain                string
	env                   []string
	clock                 clock.Clock
	tailedLogsOutputter   console_tailed_logs_outputter.TailedLogsOutputter
	exitHandler           exit_handler.ExitHandler
}

type AppRunnerCommandFactoryConfig struct {
	AppRunner             docker_app_runner.AppRunner
	AppExaminer           app_examiner.AppExaminer
	UI                    terminal.UI
	DockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	Domain                string
	Env                   []string
	Clock                 clock.Clock
	Logger                lager.Logger
	TailedLogsOutputter   console_tailed_logs_outputter.TailedLogsOutputter
	ExitHandler           exit_handler.ExitHandler
}

func NewAppRunnerCommandFactory(config AppRunnerCommandFactoryConfig) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{
		appRunner:   config.AppRunner,
		appExaminer: config.AppExaminer,
		ui:          config.UI,
		dockerMetadataFetcher: config.DockerMetadataFetcher,
		domain:                config.Domain,
		env:                   config.Env,
		clock:                 config.Clock,
		tailedLogsOutputter:   config.TailedLogsOutputter,
		exitHandler:           config.ExitHandler,
	}
}

func (factory *AppRunnerCommandFactory) MakeCreateAppCommand() cli.Command {

	var createFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "Working directory for container (overrides Docker metadata)",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "run-as-root, r",
			Usage: "Runs in the context of the root user",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "Environment variables (can be passed multiple times)",
			Value: &cli.StringSlice{},
		},
		cli.IntFlag{
			Name:  "cpu-weight",
			Usage: "Relative CPU weight for the container (valid values: 1-100)",
			Value: 100,
		},
		cli.IntFlag{
			Name:  "memory-mb, m",
			Usage: "Memory limit for container in MB",
			Value: 128,
		},
		cli.IntFlag{
			Name:  "disk-mb, d",
			Usage: "Disk limit for container in MB",
			Value: 1024,
		},
		cli.StringFlag{
			Name:  "ports, p",
			Usage: "Ports to expose on the container",
		},
		cli.IntFlag{
			Name:  "monitored-port",
			Usage: "Selects which port is used to healthcheck the app. Required for multiple exposed ports",
		},
		cli.StringFlag{
			Name: "routes",
			Usage: "Route mappings to exposed ports as follows:\n\t\t" +
				"--routes=80:web,8080:api will route web to 80 and api to 8080",
		},
		cli.IntFlag{
			Name:  "instances",
			Usage: "Number of application instances to spawn on launch",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "no-monitor",
			Usage: "Disables healthchecking for the app",
		},
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "Polling timeout for app to start",
			Value: DefaultPollingTimeout,
		},
	}

	var createAppCommand = cli.Command{
		Name:    "create",
		Aliases: []string{"cr"},
		Usage:   "Creates a docker app on lattice",
		Description: `ltc create APP_NAME DOCKER_IMAGE

   APP_NAME is required and must be unique across the Lattice cluster
   DOCKER_IMAGE is required and must match the standard docker image format
   e.g.
   		1. "cloudfoundry/lattice-app"
   		2. "redis" - for official images; resolves to library/redis

   ltc will fetch the command associated with your Docker image.
   To provide a custom command:
   ltc create APP_NAME DOCKER_IMAGE <optional flags> -- START_COMMAND APP_ARG1 APP_ARG2 ...

   ltc will also fetch the working directory associated with your Docker image.
   If the image does not specify a working directory, ltc will default the working directory to "/"
   To provide a custom working directory:
   ltc create APP_NAME DOCKER_IMAGE --working-dir=/foo/app-folder -- START_COMMAND APP_ARG1 APP_ARG2 ...

   To specify environment variables:
   ltc create APP_NAME DOCKER_IMAGE -e FOO=BAR -e BAZ=WIBBLE`,
		Action: factory.createApp,
		Flags:  createFlags,
	}

	return createAppCommand
}

func (factory *AppRunnerCommandFactory) MakeCreateLrpCommand() cli.Command {
	var createLrpCommand = cli.Command{
		Name:        "create-lrp",
		Aliases:     []string{"cl"},
		Usage:       "Creates a docker app from JSON on lattice",
		Description: "ltc create-lrp /path/to/json",
		Action:      factory.createLrp,
	}

	return createLrpCommand
}

func (factory *AppRunnerCommandFactory) MakeScaleAppCommand() cli.Command {
	var scaleFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "Polling timeout for app to scale",
			Value: DefaultPollingTimeout,
		},
	}
	var scaleAppCommand = cli.Command{
		Name:        "scale",
		Aliases:     []string{"sc"},
		Usage:       "Scales a docker app on lattice",
		Description: "ltc scale APP_NAME NUM_INSTANCES",
		Action:      factory.scaleApp,
		Flags:       scaleFlags,
	}

	return scaleAppCommand
}

func (factory *AppRunnerCommandFactory) MakeUpdateRoutesCommand() cli.Command {
	var updateRoutesCommand = cli.Command{
		Name:        "update-routes",
		Aliases:     []string{"ur"},
		Usage:       "Updates the routes for a running app",
		Description: "ltc update-routes APP_NAME ROUTE,OTHER_ROUTE...", // TODO: route format?
		Action:      factory.updateAppRoutes,
	}

	return updateRoutesCommand
}

func (factory *AppRunnerCommandFactory) MakeRemoveAppCommand() cli.Command {

	var removeAppCommand = cli.Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "ltc remove APP1_NAME [APP2_NAME APP3_NAME...]",
		Usage:       "Stops and removes docker app(s) from lattice",
		Action:      factory.removeApp,
	}

	return removeAppCommand
}

func (factory *AppRunnerCommandFactory) createApp(context *cli.Context) {
	workingDirFlag := context.String("working-dir")
	envVarsFlag := context.StringSlice("env")
	instancesFlag := context.Int("instances")
	cpuWeightFlag := uint(context.Int("cpu-weight"))
	memoryMBFlag := context.Int("memory-mb")
	diskMBFlag := context.Int("disk-mb")
	portsFlag := context.String("ports")
	monitoredPortFlag := context.Int("monitored-port")
	routesFlag := context.String("routes")
	noMonitorFlag := context.Bool("no-monitor")
	timeoutFlag := context.Duration("timeout")
	name := context.Args().Get(0)
	dockerImage := context.Args().Get(1)
	terminator := context.Args().Get(2)
	startCommand := context.Args().Get(3)

	var appArgs []string

	switch {
	case len(context.Args()) < 2:
		factory.ui.SayIncorrectUsage("APP_NAME and DOCKER_IMAGE are required")
		return
	case startCommand != "" && terminator != "--":
		factory.ui.SayIncorrectUsage("'--' Required before start command")
		return
	case len(context.Args()) > 4:
		appArgs = context.Args()[4:]
	case cpuWeightFlag < 1 || cpuWeightFlag > 100:
		factory.ui.SayIncorrectUsage("Invalid CPU Weight")
		return
	}

	imageMetadata, err := factory.dockerMetadataFetcher.FetchMetadata(dockerImage)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error fetching image metadata: %s", err))
		return
	}

	portConfig, err := factory.getPortConfigFromArgs(portsFlag, monitoredPortFlag, noMonitorFlag, imageMetadata)
	if err != nil {
		factory.ui.Say(err.Error())
		return
	}

	if workingDirFlag == "" {
		factory.ui.Say("No working directory specified, using working directory from the image metadata...\n")
		if imageMetadata.WorkingDir != "" {
			workingDirFlag = imageMetadata.WorkingDir
			factory.ui.Say("Working directory is:\n")
			factory.ui.Say(workingDirFlag + "\n")
		} else {
			workingDirFlag = "/"
		}
	}

	if !noMonitorFlag {
		factory.ui.Say(fmt.Sprintf("Monitoring the app on port %d...\n", portConfig.Monitored))
	} else {
		factory.ui.Say("No ports will be monitored.\n")
	}

	if startCommand == "" {
		if len(imageMetadata.StartCommand) == 0 {
			factory.ui.SayLine("Unable to determine start command from image metadata.")
			return
		}

		factory.ui.Say("No start command specified, using start command from the image metadata...\n")
		startCommand = imageMetadata.StartCommand[0]

		factory.ui.Say("Start command is:\n")
		factory.ui.Say(strings.Join(imageMetadata.StartCommand, " ") + "\n")

		appArgs = imageMetadata.StartCommand[1:]
	}

	routeOverrides, err := parseRouteOverrides(routesFlag)
	if err != nil {
		factory.ui.Say(err.Error())
		return
	}

	err = factory.appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
		Name:                 name,
		DockerImagePath:      dockerImage,
		StartCommand:         startCommand,
		AppArgs:              appArgs,
		EnvironmentVariables: factory.buildEnvironment(envVarsFlag, name),
		Privileged:           context.Bool("run-as-root"),
		Monitor:              !noMonitorFlag,
		Instances:            instancesFlag,
		CPUWeight:            cpuWeightFlag,
		MemoryMB:             memoryMBFlag,
		DiskMB:               diskMBFlag,
		Ports:                portConfig,
		WorkingDir:           workingDirFlag,
		RouteOverrides:       routeOverrides,
		Timeout:              timeoutFlag,
	})
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error creating app: %s", err))
		return
	}

	factory.ui.Say("Creating App: " + name + "\n")

	go factory.tailedLogsOutputter.OutputTailedLogs(name)
	defer factory.tailedLogsOutputter.StopOutputting()

	ok := factory.pollUntilAllInstancesRunning(timeoutFlag, name, instancesFlag, "start")

	if ok {
		factory.ui.Say(colors.Green(name + " is now running.\n"))
		factory.ui.Say("App is reachable at:\n")
	} else {
		factory.ui.Say("App will be reachable at:\n")
	}

	if routeOverrides != nil {
		for _, route := range strings.Split(routesFlag, ",") {
			factory.ui.Say(colors.Green(factory.urlForApp(strings.Split(route, ":")[1])))
		}
	} else {
		factory.ui.Say(colors.Green(factory.urlForApp(name)))
	}
}

func (factory *AppRunnerCommandFactory) createLrp(context *cli.Context) {

	filePath := context.Args().First()
	if filePath == "" {
		factory.ui.Say("Path to JSON is required")
		return
	}

	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}

	lrpName, err := factory.appRunner.CreateLrp(jsonBytes)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error creating %s: %s", lrpName, err.Error()))
		return
	}

	factory.ui.Say(colors.Green(fmt.Sprintf("Successfully submitted %s.", lrpName)) + "\n")
	factory.ui.Say(fmt.Sprintf("To view the status of your application: ltc status %s\n", lrpName))
}

func (factory *AppRunnerCommandFactory) scaleApp(c *cli.Context) {
	appName := c.Args().First()
	instancesArg := c.Args().Get(1)
	timeoutFlag := c.Duration("timeout")
	if appName == "" || instancesArg == "" {
		factory.ui.SayIncorrectUsage("Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'")
		return
	}

	instances, err := strconv.Atoi(instancesArg)
	if err != nil {
		factory.ui.SayIncorrectUsage("Number of Instances must be an integer")
		return
	}

	factory.setAppInstances(timeoutFlag, appName, instances)
}

func (factory *AppRunnerCommandFactory) updateAppRoutes(c *cli.Context) {
	appName := c.Args().First()
	userDefinedRoutes := c.Args().Get(1)

	if appName == "" || userDefinedRoutes == "" {
		factory.ui.SayIncorrectUsage("Please enter 'ltc update-routes APP_NAME NEW_ROUTES'")
		return
	}

	desiredRoutes, err := parseRouteOverrides(userDefinedRoutes)
	if err != nil {
		factory.ui.Say(err.Error())
		return
	}

	err = factory.appRunner.UpdateAppRoutes(appName, desiredRoutes)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error updating routes: %s", err))
		return
	}

	factory.ui.Say(fmt.Sprintf("Updating %s routes. You can check this app's current routes by running 'ltc status %s'", appName, appName))
}

func (factory *AppRunnerCommandFactory) setAppInstances(pollTimeout time.Duration, appName string, instances int) {
	err := factory.appRunner.ScaleApp(appName, instances)

	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error Scaling App to %d instances: %s", instances, err))
		return
	}

	factory.ui.Say(fmt.Sprintf("Scaling %s to %d instances \n", appName, instances))

	ok := factory.pollUntilAllInstancesRunning(pollTimeout, appName, instances, "scale")

	if ok {
		factory.ui.Say(colors.Green("App Scaled Successfully"))
	}
}

func (factory *AppRunnerCommandFactory) removeApp(c *cli.Context) {
	appNames := c.Args()

	if len(appNames) == 0 {
		factory.ui.SayIncorrectUsage("App Name required")
		return
	}

	for _, appName := range appNames {
		factory.ui.SayLine(fmt.Sprintf("Removing %s...", appName))
		err:= factory.appRunner.RemoveApp(appName)
		if err != nil {
			factory.ui.SayLine(fmt.Sprintf("Error stopping %s: %s", appName, err))
		}
	}
}

func (factory *AppRunnerCommandFactory) pollUntilSuccess(pollTimeout time.Duration, pollingFunc func() bool, outputProgress bool) (ok bool) {
	startingTime := factory.clock.Now()
	for startingTime.Add(pollTimeout).After(factory.clock.Now()) {
		if result := pollingFunc(); result {
			factory.ui.SayNewLine()
			return true
		} else if outputProgress {
			factory.ui.Say(".")
		}

		factory.clock.Sleep(1 * time.Second)
	}
	factory.ui.SayNewLine()
	return false
}

func (factory *AppRunnerCommandFactory) pollUntilAllInstancesRunning(pollTimeout time.Duration, appName string, instances int, action pollingAction) bool {
	placementErrorOccurred := false
	ok := factory.pollUntilSuccess(pollTimeout, func() bool {
		numberOfRunningInstances, placementError, _ := factory.appExaminer.RunningAppInstancesInfo(appName)
		if placementError {
			factory.ui.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity."))
			placementErrorOccurred = true
			return true
		}
		return numberOfRunningInstances == instances
	}, true)

	if placementErrorOccurred {
		factory.exitHandler.Exit(exit_codes.PlacementError)
		return false
	} else if !ok {
		if action == pollingStart {
			factory.ui.Say(colors.Red("Timed out waiting for the container to come up."))
			factory.ui.SayNewLine()
			factory.ui.SayLine("This typically happens because docker layers can take time to download.")
			factory.ui.SayLine("Lattice is still downloading your application in the background.")
		} else {
			factory.ui.Say(colors.Red("Timed out waiting for the container to scale."))
			factory.ui.SayNewLine()
			factory.ui.SayLine("Lattice is still scaling your application in the background.")
		}
		factory.ui.SayLine(fmt.Sprintf("To view logs:\n\tltc logs %s", appName))
		factory.ui.SayLine(fmt.Sprintf("To view status:\n\tltc status %s", appName))
		factory.ui.SayNewLine()
	}
	return ok
}

func (factory *AppRunnerCommandFactory) urlForApp(name string) string {
	return fmt.Sprintf("http://%s.%s\n", name, factory.domain)
}

func (factory *AppRunnerCommandFactory) buildEnvironment(envVars []string, appName string) map[string]string {
	environment := make(map[string]string)
	environment["PROCESS_GUID"] = appName

	for _, envVarPair := range envVars {
		name, value := parseEnvVarPair(envVarPair)

		if value == "" {
			value = factory.grabVarFromEnv(name)
		}

		environment[name] = value
	}
	return environment
}

func (factory *AppRunnerCommandFactory) grabVarFromEnv(name string) string {
	for _, envVarPair := range factory.env {
		if strings.HasPrefix(envVarPair, name) {
			_, value := parseEnvVarPair(envVarPair)
			return value
		}
	}
	return ""
}

func (factory *AppRunnerCommandFactory) getPortConfigFromArgs(portsFlag string, monitoredPortFlag int, noMonitorFlag bool, imageMetadata *docker_metadata_fetcher.ImageMetadata) (docker_app_runner.PortConfig, error) {

	var portConfig docker_app_runner.PortConfig
	if portsFlag == "" && !imageMetadata.Ports.IsEmpty() {
		portStrs := make([]string, 0)
		for _, port := range imageMetadata.Ports.Exposed {
			portStrs = append(portStrs, strconv.Itoa(int(port)))
		}

		factory.ui.Say(fmt.Sprintf("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: %s\n", strings.Join(portStrs, ", ")))
		portConfig = imageMetadata.Ports
	} else if portsFlag == "" && imageMetadata.Ports.IsEmpty() && noMonitorFlag {
		portConfig = docker_app_runner.PortConfig{
			Monitored: 0,
			Exposed:   []uint16{8080},
		}
	} else if portsFlag == "" && imageMetadata.Ports.IsEmpty() {
		factory.ui.Say(fmt.Sprintf("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))
		portConfig = docker_app_runner.PortConfig{
			Monitored: 8080,
			Exposed:   []uint16{8080},
		}
	} else {
		portStrings := strings.Split(portsFlag, ",")
		if len(portStrings) > 1 && monitoredPortFlag == 0 && !noMonitorFlag {
			return docker_app_runner.PortConfig{}, errors.New(MustSetMonitoredPortErrorMessage)
		}

		sort.Strings(portStrings)

		convertedPorts := []uint16{}

		for _, p := range portStrings {
			intPort, err := strconv.Atoi(p)
			if err != nil || intPort > 65535 {
				return docker_app_runner.PortConfig{}, errors.New(InvalidPortErrorMessage)
			}
			convertedPorts = append(convertedPorts, uint16(intPort))
		}

		var monitoredPort uint16
		if len(portStrings) > 1 {
			monitoredPort = uint16(monitoredPortFlag)
		} else {
			monitoredPort = convertedPorts[0]
		}

		portConfig = docker_app_runner.PortConfig{
			Monitored: monitoredPort,
			Exposed:   convertedPorts,
		}
	}

	return portConfig, nil
}

func parseRouteOverrides(routes string) (docker_app_runner.RouteOverrides, error) {
	var routeOverrides docker_app_runner.RouteOverrides

	for _, route := range strings.Split(routes, ",") {
		if route == "" {
			continue
		}
		routeArr := strings.Split(route, ":")
		maybePort, err := strconv.Atoi(routeArr[0])
		if err != nil || len(routeArr) < 2 {
			return nil, errors.New(MalformedRouteErrorMessage)
		}

		port := uint16(maybePort)
		hostnamePrefix := routeArr[1]
		routeOverrides = append(routeOverrides, docker_app_runner.RouteOverride{HostnamePrefix: hostnamePrefix, Port: port})
	}

	return routeOverrides, nil
}

func parseEnvVarPair(envVarPair string) (name, value string) {
	s := strings.Split(envVarPair, "=")
	if len(s) > 1 {
		return s[0], s[1]
	} else {
		return s[0], ""
	}
}
