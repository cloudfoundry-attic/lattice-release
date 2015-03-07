package command_factory

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_repository_name_formatter"
	"github.com/cloudfoundry-incubator/lattice/ltc/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/codegangsta/cli"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const (
	InvalidPortErrorMessage          = "Invalid port specified. Ports must be a comma-delimited list of integers between 0-65535."
	MalformedRouteErrorMessage       = "Malformed route. Routes must be of the format route:port"
	MustSetMonitoredPortErrorMessage = "Must set monitored-port when specifying multiple exposed ports unless --no-monitor is set."
)

type AppRunnerCommandFactory struct {
    appRunner             docker_app_runner.AppRunner
    dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
    output                *output.Output
    timeout               time.Duration
    domain                string
    env                   []string
    clock                 clock.Clock
    tailedLogsOutputter   console_tailed_logs_outputter.TailedLogsOutputter
    exitHandler           exit_handler.ExitHandler
}

type AppRunnerCommandFactoryConfig struct {
	AppRunner             docker_app_runner.AppRunner
	DockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	Output                *output.Output
	Timeout               time.Duration
	Domain                string
	Env                   []string
	Clock                 clock.Clock
	Logger                lager.Logger
	TailedLogsOutputter   console_tailed_logs_outputter.TailedLogsOutputter
	ExitHandler           exit_handler.ExitHandler
}

func NewAppRunnerCommandFactory(config AppRunnerCommandFactoryConfig) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{
        appRunner:             config.AppRunner,
        dockerMetadataFetcher: config.DockerMetadataFetcher,
        output:                config.Output,
        timeout:               config.Timeout,
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
			Usage: "working directory to assign to the running process.",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "run-as-root, r",
			Usage: "run process as a privileged user (root)",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "environment variables to set, NAME[=VALUE]",
			Value: &cli.StringSlice{},
		},
		cli.IntFlag{
			Name:  "memory-mb, m",
			Usage: "container memory limit in MB",
			Value: 128,
		},
		cli.IntFlag{
			Name:  "disk-mb, d",
			Usage: "container disk limit in MB",
			Value: 1024,
		},
		cli.StringFlag{
			Name:  "ports, p",
			Usage: "ports that the running process will listen on",
		},
		cli.IntFlag{
			Name:  "monitored-port",
			Usage: "the port that lattice will monitor to ensure the app is up and running. Required if specifying multiple exposed ports.",
		},
		cli.StringFlag{
			Name: "routes",
			Usage: `mapping of port to hostname. If specified, lattice will not generate any default routes.
                    eg routes=8080:foo,443:bar
                    will generate routes foo.<SystemIp>.xip.io => container port 8080 and bar.<SystemIp>.xip.io => container port 443
            `,
		},
		cli.IntFlag{
			Name:  "instances",
			Usage: "number of container instances to launch",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "no-monitor",
			Usage: "if set, lattice will not monitor that the app is listening on its port, and thus will not know if an app is running.",
		},
	}

	var createAppCommand = cli.Command{
		Name:      "create",
		ShortName: "c",
		Usage:     "ltc create APP_NAME DOCKER_IMAGE",
		Description: `Create a docker app on lattice
   
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

func (factory *AppRunnerCommandFactory) MakeScaleAppCommand() cli.Command {
	var scaleAppCommand = cli.Command{
		Name:        "scale",
        ShortName:   "sc",
		Description: "Scale a docker app on lattice",
		Usage:       "ltc scale APP_NAME NUM_INSTANCES",
		Action:      factory.scaleApp,
	}

	return scaleAppCommand
}

func (factory *AppRunnerCommandFactory) MakeUpdateRoutesCommand() cli.Command {
	var updateRoutesCommand = cli.Command{
		Name:        "update-routes",
		Description: "Updates the routes for a running app",
		Usage:       "ltc update-routes APP_NAME ROUTE,OTHER_ROUTE...", // TODO: route format?
		Action:      factory.updateAppRoutes,
	}

	return updateRoutesCommand
}

func (factory *AppRunnerCommandFactory) MakeRemoveAppCommand() cli.Command {
	var removeAppCommand = cli.Command{
		Name:        "remove",
        ShortName:   "r",
		Description: "Stop and remove a docker app from lattice",
		Usage:       "ltc remove APP_NAME",
		Action:      factory.removeApp,
	}

	return removeAppCommand
}

func (factory *AppRunnerCommandFactory) createApp(context *cli.Context) {
	workingDirFlag := context.String("working-dir")
	envVarsFlag := context.StringSlice("env")
	instancesFlag := context.Int("instances")
	memoryMBFlag := context.Int("memory-mb")
	diskMBFlag := context.Int("disk-mb")
	portsFlag := context.String("ports")
	monitoredPortFlag := context.Int("monitored-port")
	routesFlag := context.String("routes")
	noMonitorFlag := context.Bool("no-monitor")
	name := context.Args().Get(0)
	dockerImage := context.Args().Get(1)
	terminator := context.Args().Get(2)
	startCommand := context.Args().Get(3)

	var appArgs []string

	switch {
	case len(context.Args()) < 2:
		factory.output.IncorrectUsage("APP_NAME and DOCKER_IMAGE are required")
		return
	case startCommand != "" && terminator != "--":
		factory.output.IncorrectUsage("'--' Required before start command")
		return
	case len(context.Args()) > 4:
		appArgs = context.Args()[4:]
	}

	repoName, tag := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImage)
	imageMetadata, err := factory.dockerMetadataFetcher.FetchMetadata(repoName, tag)
	if err != nil {
		factory.output.Say(fmt.Sprintf("Error fetching image metadata: %s", err))
		return
	}

	portConfig, err := factory.getPortConfigFromArgs(portsFlag, monitoredPortFlag, noMonitorFlag, imageMetadata)
	if err != nil {
		factory.output.Say(err.Error())
		return
	}

	if workingDirFlag == "" {
		factory.output.Say("No working directory specified, using working directory from the image metadata...\n")
		if imageMetadata.WorkingDir != "" {
			workingDirFlag = imageMetadata.WorkingDir
			factory.output.Say("Working directory is:\n")
			factory.output.Say(workingDirFlag + "\n")
		} else {
			workingDirFlag = "/"
		}
	}

	if !noMonitorFlag {
		factory.output.Say(fmt.Sprintf("Monitoring the app on port %d...\n", portConfig.Monitored))
	} else {
		factory.output.Say("No ports will be monitored.\n")
	}

	if startCommand == "" {
		factory.output.Say("No start command specified, using start command from the image metadata...\n")

		startCommand = imageMetadata.StartCommand[0]
		factory.output.Say("Start command is:\n")
		factory.output.Say(strings.Join(imageMetadata.StartCommand, " ") + "\n")

		appArgs = imageMetadata.StartCommand[1:]
	}

	routeOverrides, err := parseRouteOverrides(routesFlag)
	if err != nil {
		factory.output.Say(err.Error())
		return
	}

	err = factory.appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
		Name:                 name,
		DockerImagePath:      dockerImage,
		StartCommand:         startCommand,
		AppArgs:              appArgs,
		EnvironmentVariables: factory.buildEnvironment(envVarsFlag),
		Privileged:           context.Bool("run-as-root"),
		Monitor:              !noMonitorFlag,
		Instances:            instancesFlag,
		MemoryMB:             memoryMBFlag,
		DiskMB:               diskMBFlag,
		Ports:                portConfig,
		WorkingDir:           workingDirFlag,
		RouteOverrides:       routeOverrides,
	})
	if err != nil {
		factory.output.Say(fmt.Sprintf("Error Creating App: %s", err))
		return
	}

	factory.output.Say("Creating App: " + name + "\n")

	go factory.tailedLogsOutputter.OutputTailedLogs(name)
	defer factory.tailedLogsOutputter.StopOutputting()

	ok := factory.pollUntilAllInstancesRunning(name, instancesFlag, "start")

	if ok {
		factory.output.Say(colors.Green(name + " is now running.\n"))
	}

	if routeOverrides != nil {
		for _, route := range strings.Split(routesFlag, ",") {
			factory.output.Say(colors.Green(factory.urlForApp(strings.Split(route, ":")[1])))
		}
	} else {
		factory.output.Say(colors.Green(factory.urlForApp(name)))
	}
}

func (factory *AppRunnerCommandFactory) scaleApp(c *cli.Context) {
	appName := c.Args().First()
	instancesArg := c.Args().Get(1)

	if appName == "" || instancesArg == "" {
		factory.output.IncorrectUsage("Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'")
		return
	}

	instances, err := strconv.Atoi(instancesArg)
	if err != nil {
		factory.output.IncorrectUsage("Number of Instances must be an integer")
		return
	}

	factory.setAppInstances(appName, instances)
}

func (factory *AppRunnerCommandFactory) updateAppRoutes(c *cli.Context) {
	appName := c.Args().First()
	userDefinedRoutes := c.Args().Get(1)

	if appName == "" || userDefinedRoutes == "" {
		factory.output.IncorrectUsage("Please enter 'ltc update-routes APP_NAME NEW_ROUTES'")
		return
	}

	desiredRoutes, err := parseRouteOverrides(userDefinedRoutes)
	if err != nil {
		factory.output.Say(err.Error())
		return
	}

	err = factory.appRunner.UpdateAppRoutes(appName, desiredRoutes)
	if err != nil {
		factory.output.Say(fmt.Sprintf("Error updating routes: %s", err))
		return
	}

	factory.output.Say(fmt.Sprintf("Updating %s routes. You can check this app's current routes by running 'ltc status %s'", appName, appName))
}

func (factory *AppRunnerCommandFactory) setAppInstances(appName string, instances int) {
	err := factory.appRunner.ScaleApp(appName, instances)

	if err != nil {
		factory.output.Say(fmt.Sprintf("Error Scaling App to %d instances: %s", instances, err))
		return
	}

	factory.output.Say(fmt.Sprintf("Scaling %s to %d instances \n", appName, instances))

	ok := factory.pollUntilAllInstancesRunning(appName, instances, "scale")

	if ok {
		factory.output.Say(colors.Green("App Scaled Successfully"))
    }
}

func (factory *AppRunnerCommandFactory) pollUntilAllInstancesRunning(appName string, instances int, action string) bool {
	placementErrorOccurred := false
	ok := factory.pollUntilSuccess(func() bool {
		numberOfRunningInstances, placementError, _ := factory.appRunner.RunningAppInstancesInfo(appName)
		if placementError {
			factory.output.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity."))
			placementErrorOccurred = true
			return true
		}
		return numberOfRunningInstances == instances
	}, true)

	if placementErrorOccurred {
		factory.exitHandler.Exit(exit_codes.PlacementError)
		return false
	} else if !ok {
        factory.output.Say(colors.Red(appName + " took too long to " + action + "."))
	}
    return ok

}

func (factory *AppRunnerCommandFactory) removeApp(c *cli.Context) {
	appName := c.Args().First()
	if appName == "" {
		factory.output.IncorrectUsage("App Name required")
		return
	}

	err := factory.appRunner.RemoveApp(appName)
	if err != nil {
		factory.output.Say(fmt.Sprintf("Error Stopping App: %s", err))
		return
	}

	factory.output.Say(fmt.Sprintf("Removing %s", appName))
	ok := factory.pollUntilSuccess(func() bool {
		appExists, err := factory.appRunner.AppExists(appName)
		return err == nil && !appExists
	}, true)

	if ok {
		factory.output.Say(colors.Green("Successfully Removed " + appName + "."))
	} else {
		factory.output.Say(colors.Red(fmt.Sprintf("Failed to remove %s.", appName)))
	}
}

func (factory *AppRunnerCommandFactory) pollUntilSuccess(pollingFunc func() bool, outputProgress bool) (ok bool) {
	startingTime := factory.clock.Now()
	for startingTime.Add(factory.timeout).After(factory.clock.Now()) {
		if result := pollingFunc(); result {
			factory.output.NewLine()
			return true
		} else if outputProgress {
			factory.output.Say(".")
		}

		factory.clock.Sleep(1 * time.Second)
	}
	factory.output.NewLine()
	return false
}

func (factory *AppRunnerCommandFactory) urlForApp(name string) string {
	return fmt.Sprintf("http://%s.%s\n", name, factory.domain)
}

func (factory *AppRunnerCommandFactory) buildEnvironment(envVars []string) map[string]string {
	environment := make(map[string]string)

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

		factory.output.Say(fmt.Sprintf("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: %s\n", strings.Join(portStrs, ", ")))
		portConfig = imageMetadata.Ports
	} else if portsFlag == "" && imageMetadata.Ports.IsEmpty() && noMonitorFlag {
		portConfig = docker_app_runner.PortConfig{
			Monitored: 0,
			Exposed:   []uint16{8080},
		}
	} else if portsFlag == "" && imageMetadata.Ports.IsEmpty() {
		factory.output.Say(fmt.Sprintf("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))
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
