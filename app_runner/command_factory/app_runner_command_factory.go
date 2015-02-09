package command_factory

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/console_tailed_logs_outputter"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
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
}

func NewAppRunnerCommandFactory(config AppRunnerCommandFactoryConfig) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{
		&appRunnerCommand{
			appRunner:             config.AppRunner,
			dockerMetadataFetcher: config.DockerMetadataFetcher,
			output:                config.Output,
			timeout:               config.Timeout,
			domain:                config.Domain,
			env:                   config.Env,
			clock:                 config.Clock,
			tailedLogsOutputter:   config.TailedLogsOutputter,
		},
	}
}

func (commandFactory *AppRunnerCommandFactory) MakeStartAppCommand() cli.Command {

	var startFlags = []cli.Flag{
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
		cli.IntFlag{
			Name:  "port, p",
			Usage: "port that the running process will listen on",
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

	var startCommand = cli.Command{
		Name:      "start",
		ShortName: "s",
		Usage:     "ltc start APP_NAME DOCKER_IMAGE",
		Description: `Start a docker app on lattice
   
   APP_NAME is required and must be unique across the Lattice cluster
   DOCKER_IMAGE is required and must match the standard docker image format
   e.g.
   		1. "cloudfoundry/lattice-app"
   		2. "redis" - for official images; resolves to library/redis

   ltc will fetch the command associated with your Docker image.
   To provide a custom command:
   ltc start APP_NAME DOCKER_IMAGE <optional flags> -- START_COMMAND APP_ARG1 APP_ARG2 ...

   ltc will also fetch the working directory associated with your Docker image.
   If the image does not specify a working directory, ltc will default the working directory to "/"
   To provide a custom working directory:
   ltc start APP_NAME DOCKER_IMAGE --working-dir=/foo/app-folder -- START_COMMAND APP_ARG1 APP_ARG2 ...

   To specify environment variables:
   ltc start APP_NAME DOCKER_IMAGE -e FOO=BAR -e BAZ=WIBBLE`,
		Action: commandFactory.appRunnerCommand.startApp,
		Flags:  startFlags,
	}

	return startCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeScaleAppCommand() cli.Command {
	var scaleCommand = cli.Command{
		Name:        "scale",
		Description: "Scale a docker app on lattice",
		Usage:       "ltc scale APP_NAME NUM_INSTANCES",
		Action:      commandFactory.appRunnerCommand.scaleApp,
	}

	return scaleCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeStopAppCommand() cli.Command {

	var stopCommand = cli.Command{
		Name: "stop",
		Description: `Stop a docker app on lattice

   The application can be restarted with ltc scale`,
		Usage:  "ltc stop APP_NAME",
		Action: commandFactory.appRunnerCommand.stopApp,
	}

	return stopCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeRemoveAppCommand() cli.Command {
	var removeCommand = cli.Command{
		Name:        "remove",
		Description: "Stop and remove a docker app from lattice",
		Usage:       "ltc remove APP_NAME",
		Action:      commandFactory.appRunnerCommand.removeApp,
	}

	return removeCommand
}

type appRunnerCommand struct {
	appRunner             docker_app_runner.AppRunner
	dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	output                *output.Output
	timeout               time.Duration
	domain                string
	env                   []string
	clock                 clock.Clock
	tailedLogsOutputter   console_tailed_logs_outputter.TailedLogsOutputter
}

func (cmd *appRunnerCommand) startApp(context *cli.Context) {
	workingDir := context.String("working-dir")
	envVars := context.StringSlice("env")
	instances := context.Int("instances")
	memoryMB := context.Int("memory-mb")
	diskMB := context.Int("disk-mb")
	portFlag := context.Int("port")
	noMonitor := context.Bool("no-monitor")
	name := context.Args().Get(0)
	dockerImage := context.Args().Get(1)
	terminator := context.Args().Get(2)
	startCommand := context.Args().Get(3)

	var appArgs []string

	switch {
	case len(context.Args()) < 2:
		cmd.output.IncorrectUsage("APP_NAME and DOCKER_IMAGE are required")
		return
	case startCommand != "" && terminator != "--":
		cmd.output.IncorrectUsage("'--' Required before start command")
		return
	case len(context.Args()) > 4:
		appArgs = context.Args()[4:]
	}

	imageMetadata, err := cmd.dockerMetadataFetcher.FetchMetadata(dockerImage, "latest")
	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error fetching image metadata: %s", err))
		return
	}

	var portConfig docker_app_runner.PortConfig
	if portFlag == 0 && !imageMetadata.Ports.IsEmpty() {
		portStrs := make([]string, 0)
		for _, port := range imageMetadata.Ports.Exposed {
			portStrs = append(portStrs, strconv.Itoa(int(port)))
		}

		cmd.output.Say(fmt.Sprintf("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: %s\n", strings.Join(portStrs, ", ")))
		portConfig = imageMetadata.Ports
	} else if portFlag == 0 && imageMetadata.Ports.IsEmpty() && noMonitor {
		portConfig = docker_app_runner.PortConfig{
			Monitored: 0,
			Exposed:   []uint16{8080},
		}
	} else if portFlag == 0 && imageMetadata.Ports.IsEmpty() {
		cmd.output.Say(fmt.Sprintf("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))
		portConfig = docker_app_runner.PortConfig{
			Monitored: 8080,
			Exposed:   []uint16{8080},
		}
	} else {
		portConfig = docker_app_runner.PortConfig{
			Monitored: uint16(portFlag),
			Exposed:   []uint16{uint16(portFlag)},
		}
	}

	if workingDir == "" {
		cmd.output.Say("No working directory specified, using working directory from the image metadata...\n")
		if imageMetadata.WorkingDir != "" {
			workingDir = imageMetadata.WorkingDir
			cmd.output.Say("Working directory is:\n")
			cmd.output.Say(workingDir + "\n")
		} else {
			workingDir = "/"
		}
	}

	if !noMonitor {
		cmd.output.Say(fmt.Sprintf("Monitoring the app on port %d...\n", portConfig.Monitored))
	} else {
		cmd.output.Say("No ports will be monitored.\n")
	}

	if startCommand == "" {
		cmd.output.Say("No start command specified, using start command from the image metadata...\n")

		startCommand = imageMetadata.StartCommand[0]
		cmd.output.Say("Start command is:\n")
		cmd.output.Say(strings.Join(imageMetadata.StartCommand, " ") + "\n")

		appArgs = imageMetadata.StartCommand[1:]
	}

	err = cmd.appRunner.StartDockerApp(docker_app_runner.StartDockerAppParams{
		Name:                 name,
		DockerImagePath:      dockerImage,
		StartCommand:         startCommand,
		AppArgs:              appArgs,
		EnvironmentVariables: cmd.buildEnvironment(envVars),
		Privileged:           context.Bool("run-as-root"),
		Monitor:              !noMonitor,
		Instances:            instances,
		MemoryMB:             memoryMB,
		DiskMB:               diskMB,
		Ports:                portConfig,
		WorkingDir:           workingDir,
	})

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Starting App: %s", err))
		return
	}

	cmd.output.Say("Starting App: " + name + "\n")

	go cmd.tailedLogsOutputter.OutputTailedLogs(name)

	ok := cmd.pollUntilSuccess(func() bool {
		numberOfRunningInstances, _ := cmd.appRunner.NumOfRunningAppInstances(name)
		return numberOfRunningInstances == instances
	}, false)

	cmd.tailedLogsOutputter.StopOutputting()
	if ok {
		cmd.output.Say(colors.Green(name + " is now running.\n"))
		cmd.output.Say(colors.Green(cmd.urlForApp(name)))
	} else {
		cmd.output.Say(colors.Red(name + " took too long to start."))
	}

}

func (cmd *appRunnerCommand) scaleApp(c *cli.Context) {
	appName := c.Args().First()
	instancesArg := c.Args().Get(1)

	switch {
	case appName == "":
		cmd.output.IncorrectUsage("App Name required")
		return
	case instancesArg == "":
		cmd.output.IncorrectUsage("Number of Instances Required")
		return
	}

	instances, err := strconv.Atoi(instancesArg)
	if err != nil {
		cmd.output.IncorrectUsage("Number of Instances must be an integer")
		return
	}

	cmd.setAppInstances(appName, instances)
}

func (cmd *appRunnerCommand) stopApp(c *cli.Context) {
	appName := c.Args().First()

	if appName == "" {
		cmd.output.IncorrectUsage("App Name required")
		return
	}

	cmd.setAppInstances(appName, 0)
}

func (cmd *appRunnerCommand) setAppInstances(appName string, instances int) {
	err := cmd.appRunner.ScaleApp(appName, instances)

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Scaling App to %d instances: %s", instances, err))
		return
	}

	cmd.output.Say(fmt.Sprintf("Scaling %s to %d instances", appName, instances))

	ok := cmd.pollUntilSuccess(func() bool {
		numRunning, _ := cmd.appRunner.NumOfRunningAppInstances(appName)
		return numRunning == instances
	}, true)

	if ok {
		cmd.output.Say(colors.Green("App Scaled Successfully"))
	} else {
		cmd.output.Say(colors.Red(appName + " took too long to scale."))
	}
}

func (cmd *appRunnerCommand) removeApp(c *cli.Context) {
	appName := c.Args().First()
	if appName == "" {
		cmd.output.IncorrectUsage("App Name required")
		return
	}

	err := cmd.appRunner.RemoveApp(appName)
	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Stopping App: %s", err))
		return
	}

	cmd.output.Say(fmt.Sprintf("Removing %s", appName))
	ok := cmd.pollUntilSuccess(func() bool {
		appExists, err := cmd.appRunner.AppExists(appName)
		return err == nil && !appExists
	}, true)

	if ok {
		cmd.output.Say(colors.Green("Successfully Removed " + appName + "."))
	} else {
		cmd.output.Say(colors.Red(fmt.Sprintf("Failed to remove %s.", appName)))
	}
}

func (cmd *appRunnerCommand) pollUntilSuccess(pollingFunc func() bool, outputProgress bool) (ok bool) {
	startingTime := cmd.clock.Now()
	for startingTime.Add(cmd.timeout).After(cmd.clock.Now()) {
		if result := pollingFunc(); result {
			cmd.output.NewLine()
			return true
		} else if outputProgress {
			cmd.output.Say(".")
		}

		cmd.clock.Sleep(1 * time.Second)
	}
	cmd.output.NewLine()
	return false
}

func (cmd *appRunnerCommand) urlForApp(name string) string {
	return fmt.Sprintf("http://%s.%s", name, cmd.domain)
}

func (cmd *appRunnerCommand) buildEnvironment(envVars []string) map[string]string {
	environment := make(map[string]string)

	for _, envVarPair := range envVars {
		name, value := parseEnvVarPair(envVarPair)

		if value == "" {
			value = cmd.grabVarFromEnv(name)
		}

		environment[name] = value
	}
	return environment
}

func (cmd *appRunnerCommand) grabVarFromEnv(name string) string {
	for _, envVarPair := range cmd.env {
		if strings.HasPrefix(envVarPair, name) {
			_, value := parseEnvVarPair(envVarPair)
			return value
		}
	}
	return ""
}

func parseEnvVarPair(envVarPair string) (name, value string) {
	s := strings.Split(envVarPair, "=")
	if len(s) > 1 {
		return s[0], s[1]
	} else {
		return s[0], ""
	}
}
