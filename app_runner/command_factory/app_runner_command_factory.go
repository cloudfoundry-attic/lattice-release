package command_factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-golang/lager"
)

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
}

type AppRunnerCommandFactoryConfig struct {
	AppRunner             app_runner.AppRunner
	DockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	Output                *output.Output
	Timeout               time.Duration
	Domain                string
	Env                   []string
	TimeProvider          timeprovider.TimeProvider
	Logger                lager.Logger
}

func NewAppRunnerCommandFactory(config AppRunnerCommandFactoryConfig) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{&appRunnerCommand{config.AppRunner, config.DockerMetadataFetcher, config.Output, config.Timeout, config.Domain, config.Env, config.TimeProvider}}
}

func (commandFactory *AppRunnerCommandFactory) MakeStartAppCommand() cli.Command {

	var startFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-image, i",
			Usage: "docker image to run",
		},
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "working directory to assign to the running process",
			Value: "/",
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
			Value: 8080,
		},
		cli.IntFlag{
			Name:  "instances",
			Usage: "number of container instances to launch",
			Value: 1,
		},
	}

	var startCommand = cli.Command{
		Name:      "start",
		ShortName: "s",
		Usage:     "ltc start APP_NAME -i DOCKER_IMAGE",
		Description: `Start a docker app on lattice
   
   APP_NAME is required and must be unique across the Lattice cluster
   DOCKER_IMAGE is required and must match the format (e.g.)
   "cloudfoundry/lattice-app" or "redis" (for official images)

   ltc will fetch the command associated with your Docker image.
   To provide a custom command:
   ltc start APP_NAME -i DOCKER_IMAGE -- START_COMMAND APP_ARG1 APP_ARG2 ...

   To specify environment variables:
   ltc start APP_NAME -i DOCKER_IMAGE -e FOO=BAR -e BAZ=WIBBLE`,
		Action: commandFactory.appRunnerCommand.startApp,
		Flags:  startFlags,
	}

	return startCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeScaleAppCommand() cli.Command {

	var scaleFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "instances, i",
			Usage: "the number of instances to scale to",
		},
	}

	var scaleCommand = cli.Command{
		Name:        "scale",
		Description: "Scale a docker app on lattice",
		Usage:       "ltc scale APP_NAME --instances NUM_INSTANCES",
		Action:      commandFactory.appRunnerCommand.scaleApp,
		Flags:       scaleFlags,
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
	appRunner             app_runner.AppRunner
	dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
	output                *output.Output
	timeout               time.Duration
	domain                string
	env                   []string
	timeProvider          timeprovider.TimeProvider
}

func (cmd *appRunnerCommand) startApp(context *cli.Context) {
	dockerImage := context.String("docker-image")
	workingDir := context.String("working-dir")
	envVars := context.StringSlice("env")
	privileged := context.Bool("run-as-root")
	instances := context.Int("instances")
	memoryMB := context.Int("memory-mb")
	diskMB := context.Int("disk-mb")
	port := context.Int("port")
	name := context.Args().Get(0)
	terminator := context.Args().Get(1)
	startCommand := context.Args().Get(2)

	var appArgs []string

	switch {
	case name == "":
		cmd.output.IncorrectUsage("App Name required")
		return
	case dockerImage == "":
		cmd.output.IncorrectUsage("Docker Image required")
		return
	case startCommand != "" && terminator != "--":
		cmd.output.IncorrectUsage("'--' Required before start command")
		return
	case len(context.Args()) > 3:
		appArgs = context.Args()[3:]
	}

	imageMetadata, err := cmd.dockerMetadataFetcher.FetchMetadata(dockerImage, "latest")
	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error fetching metadata: %s", err))
		return
	}

	if startCommand == "" {
		cmd.output.Say("No start command specified, fetching metadata from the Dockerimage...\n")

		startCommand = imageMetadata.StartCommand[0]
		cmd.output.Say("Start command is:\n")
		cmd.output.Say(strings.Join(imageMetadata.StartCommand, " ") + "\n")

		workingDir = imageMetadata.WorkingDir
		if workingDir != "" {
			cmd.output.Say("Working directory is:\n")
			cmd.output.Say(workingDir + "\n")
		}

		appArgs = imageMetadata.StartCommand[1:]
	}

	err = cmd.appRunner.StartDockerApp(app_runner.StartDockerAppParams{
		Name:                 name,
		DockerImagePath:      dockerImage,
		StartCommand:         startCommand,
		AppArgs:              appArgs,
		EnvironmentVariables: cmd.buildEnvironment(envVars),
		Privileged:           privileged,
		Instances:            instances,
		MemoryMB:             memoryMB,
		DiskMB:               diskMB,
		Port:                 port,
		WorkingDir:           workingDir,
	})

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Starting App: %s", err))
		return
	}

	cmd.output.Say("Starting App: " + name)

	ok := cmd.pollUntilSuccess(func() bool {
		numberOfRunningInstances, _ := cmd.appRunner.NumOfRunningAppInstances(name)
		return numberOfRunningInstances == instances
	})

	if ok {
		cmd.output.Say(colors.Green(name + " is now running.\n"))
		cmd.output.Say(colors.Green(cmd.urlForApp(name)))
	} else {
		cmd.output.Say(colors.Red(name + " took too long to start."))
	}

}

func (cmd *appRunnerCommand) scaleApp(c *cli.Context) {
	appName := c.Args().First()
	instances := c.Int("instances")

	if appName == "" {
		cmd.output.IncorrectUsage("App Name required")
		return
	} else if !c.IsSet("instances") {
		cmd.output.IncorrectUsage("Number of Instances Required")
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
	})

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
	})

	if ok {
		cmd.output.Say(colors.Green("Successfully Removed " + appName + "."))
	} else {
		cmd.output.Say(colors.Red(fmt.Sprintf("Failed to remove %s.", appName)))
	}
}

func (cmd *appRunnerCommand) pollUntilSuccess(pollingFunc func() bool) (ok bool) {
	startingTime := cmd.timeProvider.Now()
	for startingTime.Add(cmd.timeout).After(cmd.timeProvider.Now()) {
		if result := pollingFunc(); result {
			cmd.output.NewLine()
			return true
		} else {
			cmd.output.Dot()
		}
		cmd.timeProvider.Sleep(1 * time.Second)
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
