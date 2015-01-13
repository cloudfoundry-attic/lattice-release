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
	"log"
	"os"
)

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
}

func NewAppRunnerCommandFactory(appRunner app_runner.AppRunner, dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher, output *output.Output, timeout time.Duration, domain string, env []string, timeProvider timeprovider.TimeProvider) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{&appRunnerCommand{appRunner, dockerMetadataFetcher, output, timeout, domain, env, timeProvider}}
}

func (commandFactory *AppRunnerCommandFactory) MakeStartAppCommand() cli.Command {

	var startFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-image, i",
			Usage: "docker image to run",
		},
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "directory where lattice will run the start command",
			Value: "/",
		},
		cli.BoolFlag{
			Name:  "run-as-root, r",
			Usage: "run app as privileged user",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "environment variables to set, NAME[=VALUE]",
			Value: &cli.StringSlice{},
		},
		cli.IntFlag{
			Name:  "memory-mb, m",
			Usage: "amount of memory in MB to provide for the docker app",
			Value: 128,
		},
		cli.IntFlag{
			Name:  "disk-mb, d",
			Usage: "amount of disk memory in MB to provide for the docker app",
			Value: 1024,
		},
		cli.IntFlag{
			Name:  "port, p",
			Usage: "port that the docker app listens on",
			Value: 8080,
		},
		cli.IntFlag{
			Name:  "instances",
			Usage: "number of app instances",
			Value: 1,
		},
	}

	var startCommand = cli.Command{
		Name:        "start",
		ShortName:   "s",
		Description: "Start a docker app on lattice",
		Usage:       "ltc start APP_NAME -i DOCKER_IMAGE -e NAME[=VALUE] -- START_COMMAND [APP_ARG1 APP_ARG2...]",
		Action:      commandFactory.appRunnerCommand.startApp,
		Flags:       startFlags,
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
		Usage:       "ltc scale APP_NAME --instances NUM_INSTANCES ",
		Action:      commandFactory.appRunnerCommand.scaleApp,
		Flags:       scaleFlags,
	}

	return scaleCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeStopAppCommand() cli.Command {

	var stopCommand = cli.Command{
		Name:        "stop",
		Description: "Stop a docker app on lattice",
		Usage:       "ltc stop APP_NAME ",
		Action:      commandFactory.appRunnerCommand.stopApp,
	}

	return stopCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeRemoveAppCommand() cli.Command {

	var removeCommand = cli.Command{
		Name:        "remove",
		Description: "Remove a docker app from lattice",
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

	debugLogger := log.New(os.Stderr,
		"\nDebug:", log.Ltime|log.Lshortfile)

	debugLogger.Printf("startApp context.args: %#v\n", context.Args())

	var appArgs []string

	switch {
	case name == "":
		cmd.output.IncorrectUsage("App Name required")
		return
	case dockerImage == "":
		cmd.output.IncorrectUsage("Docker Image required")
		return
	case !strings.HasPrefix(dockerImage, "docker:///"):
		cmd.output.IncorrectUsage("Docker Image should begin with: docker:///")
		return
	case startCommand != "" && terminator != "--":
		cmd.output.IncorrectUsage("'--' Required before start command")
		return
	case len(context.Args()) > 3:
		appArgs = context.Args()[3:]
	case startCommand == "":
		cmd.output.Say("No start command specified, fetching metadata from the Dockerimage...\n")

		repoName := strings.Split(dockerImage, "docker:///")[1]
		imageMetadata, err := cmd.dockerMetadataFetcher.FetchMetadata(repoName, "latest")
		if err != nil {
			cmd.output.Say(fmt.Sprintf("Error Fetching metadata: %s", err))
			return
		}

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

	err := cmd.appRunner.StartDockerApp(app_runner.StartDockerAppParams{
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
