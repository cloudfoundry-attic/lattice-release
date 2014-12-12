package command_factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
}

func NewAppRunnerCommandFactory(appRunner app_runner.AppRunner, output *output.Output, timeout time.Duration, domain string, env []string, timeProvider timeprovider.TimeProvider) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{&appRunnerCommand{appRunner, output, timeout, domain, env, timeProvider}}
}

func (commandFactory *AppRunnerCommandFactory) MakeStartAppCommand() cli.Command {

	var startFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-image, i",
			Usage: "docker image to run",
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

func (commandFactory *AppRunnerCommandFactory) MakeRemoveAppCommand() cli.Command {

	var removeCommand = cli.Command{
		Name:        "remove",
		Description: "Remove a docker app from lattice",
		Usage:       "ltc stop APP_NAME",
		Action:      commandFactory.appRunnerCommand.removeApp,
	}

	return removeCommand
}

type appRunnerCommand struct {
	appRunner    app_runner.AppRunner
	output       *output.Output
	timeout      time.Duration
	domain       string
	env          []string
	timeProvider timeprovider.TimeProvider
}

func (cmd *appRunnerCommand) startApp(context *cli.Context) {
	dockerImage := context.String("docker-image")
	envVars := context.StringSlice("env")
	privileged := context.Bool("run-as-root")
	memoryMB := context.Int("memory-mb")
	diskMB := context.Int("disk-mb")
	port := context.Int("port")
	name := context.Args().First()

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
	case len(context.Args()) < 3:
		cmd.output.IncorrectUsage("Start Command required")
		return
	case context.Args().Get(1) != "--":
		cmd.output.IncorrectUsage("'--' Required before start command")
		return
	}
	startCommand := context.Args().Get(2)
	appArgs := context.Args()[3:]

	err := cmd.appRunner.StartDockerApp(name, dockerImage, startCommand, appArgs, cmd.buildEnvironment(envVars), privileged, memoryMB, diskMB, port)

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Starting App: %s", err))
		return
	}

	cmd.output.Say("Starting App: " + name)

	ok := cmd.pollUntilSuccess(func() bool {
		appUp, _ := cmd.appRunner.IsAppUp(name)
		return appUp
	})

	if ok {
		cmd.output.Say(colors.Green(name + " is now running.\n"))
		cmd.output.Say(colors.Green(cmd.urlForApp(name)))
	} else {
		cmd.output.Say(colors.Red(name + " took too long to start."))
	}

}

func (cmd *appRunnerCommand) scaleApp(c *cli.Context) {
	instances := c.Int("instances")
	appName := c.Args().First()
	if appName == "" {
		cmd.output.IncorrectUsage("App Name required")
		return
	} else if instances == 0 {
		cmd.output.Say(fmt.Sprintf("Error Scaling to 0 instances - Please stop with: lattice-cli stop cool-web-app"))
		return
	}

	err := cmd.appRunner.ScaleApp(appName, instances)

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Scaling App: %s", err))
		return
	}

	cmd.output.Say("App Scaled Successfully")
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
