package command_factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type appRunner interface {
	StartDockerApp(name, startCommand, dockerImagePath string, appArgs []string, environmentVariables map[string]string, privileged bool, memoryMB, diskMB, port int) error
	ScaleDockerApp(name string, instances int) error
	RemoveDockerApp(name string) error
	IsDockerAppUp(name string) (bool, error)
}

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
}

func NewAppRunnerCommandFactory(appRunner appRunner, output *output.Output, timeout time.Duration, domain string, env []string) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{&appRunnerCommand{appRunner, output, timeout, domain, env}}
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
	appRunner appRunner
	output    *output.Output
	timeout   time.Duration
	domain    string
	env       []string
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

	cmd.pollAppUntilUp(name)
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

	err := cmd.appRunner.ScaleDockerApp(appName, instances)

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

	err := cmd.appRunner.RemoveDockerApp(appName)

	if err != nil {
		cmd.output.Say(fmt.Sprintf("Error Stopping App: %s", err))
		return
	}

	cmd.output.Say("App Removed Successfully")
}

func (cmd *appRunnerCommand) pollAppUntilUp(name string) {
	cmd.output.Say("Starting App: " + name)
	startingTime := time.Now()
	for startingTime.Add(cmd.timeout).After(time.Now()) {
		if status, _ := cmd.appRunner.IsDockerAppUp(name); status {
			cmd.output.NewLine()
			cmd.output.Say(colors.Green(name + " is now running."))
			cmd.output.NewLine()
			cmd.output.Say(colors.Green(cmd.urlForApp(name)))
			return
		} else {
			cmd.output.Dot()
		}
		time.Sleep(time.Second)
	}
	cmd.output.NewLine()
	cmd.output.Say(colors.Red(name + " took too long to start."))
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
