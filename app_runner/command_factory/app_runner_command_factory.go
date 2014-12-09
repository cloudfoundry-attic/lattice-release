package command_factory

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/colors"
)

type appRunner interface {
	StartDockerApp(name, startCommand, dockerImagePath string, appArgs []string, environmentVariables map[string]string, privileged bool, memoryMB, diskMB, port int) error
	ScaleDockerApp(name string, instances int) error
	StopDockerApp(name string) error
	IsDockerAppUp(name string) (bool, error)
}

type AppRunnerCommandFactory struct {
	appRunnerCommand *appRunnerCommand
}

func NewAppRunnerCommandFactory(appRunner appRunner, output io.Writer, timeout time.Duration, domain string, env []string) *AppRunnerCommandFactory {
	return &AppRunnerCommandFactory{&appRunnerCommand{appRunner, output, timeout, domain, env}}
}

func (commandFactory *AppRunnerCommandFactory) MakeStartDiegoAppCommand() cli.Command {

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
		Description: "Start a docker app on diego",
		Usage:       "diego-edge-cli start APP_NAME -i DOCKER_IMAGE -e NAME[=VALUE] -- START_COMMAND [APP_ARG1 APP_ARG2...]",
		Action:      commandFactory.appRunnerCommand.startDiegoApp,
		Flags:       startFlags,
	}

	return startCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeScaleDiegoAppCommand() cli.Command {

	var scaleFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "instances, i",
			Usage: "the number of instances to scale to",
		},
	}

	var scaleCommand = cli.Command{
		Name:        "scale",
		Description: "Scale a docker app on diego",
		Usage:       "diego-edge-cli scale APP_NAME --instances NUM_INSTANCES ",
		Action:      commandFactory.appRunnerCommand.scaleDiegoApp,
		Flags:       scaleFlags,
	}

	return scaleCommand
}

func (commandFactory *AppRunnerCommandFactory) MakeStopDiegoAppCommand() cli.Command {

	var stopCommand = cli.Command{
		Name:        "stop",
		Description: "Stop a docker app on diego",
		Usage:       "diego-edge-cli stop APP_NAME",
		Action:      commandFactory.appRunnerCommand.stopDiegoApp,
	}

	return stopCommand
}

type appRunnerCommand struct {
	appRunner appRunner
	output    io.Writer
	timeout   time.Duration
	domain    string
	env       []string
}

func (cmd *appRunnerCommand) startDiegoApp(c *cli.Context) {
	dockerImage := c.String("docker-image")
	envVars := c.StringSlice("env")
	privileged := c.Bool("run-as-root")
	memoryMB := c.Int("memory-mb")
	diskMB := c.Int("disk-mb")
	port := c.Int("port")
	name := c.Args().First()

	switch {
	case name == "":
		cmd.incorrectUsage("App Name required")
		return
	case dockerImage == "":
		cmd.incorrectUsage("Docker Image required")
		return
	case !strings.HasPrefix(dockerImage, "docker:///"):
		cmd.incorrectUsage("Docker Image should begin with: docker:///")
		return
	case len(c.Args()) < 3:
		cmd.incorrectUsage("Start Command required")
		return
	case c.Args().Get(1) != "--":
		cmd.incorrectUsage("'--' Required before start command")
		return
	}
	startCommand := c.Args().Get(2)
	appArgs := c.Args()[3:]

	err := cmd.appRunner.StartDockerApp(name, dockerImage, startCommand, appArgs, cmd.buildEnvironment(envVars), privileged, memoryMB, diskMB, port)

	if err != nil {
		cmd.say(fmt.Sprintf("Error Starting App: %s", err))
		return
	}

	cmd.pollAppUntilUp(name)
}

func (cmd *appRunnerCommand) scaleDiegoApp(c *cli.Context) {
	instances := c.Int("instances")
	appName := c.Args().First()
	if appName == "" {
		cmd.incorrectUsage("App Name required")
		return
	} else if instances == 0 {
		cmd.say(fmt.Sprintf("Error Scaling to 0 instances - Please stop with: diego-edge-cli stop cool-web-app"))
		return
	}

	err := cmd.appRunner.ScaleDockerApp(appName, instances)

	if err != nil {
		cmd.say(fmt.Sprintf("Error Scaling App: %s", err))
		return
	}

	cmd.say("App Scaled Successfully")
}

func (cmd *appRunnerCommand) stopDiegoApp(c *cli.Context) {
	appName := c.Args().First()
	if appName == "" {
		cmd.incorrectUsage("App Name required")
		return
	}

	err := cmd.appRunner.StopDockerApp(appName)

	if err != nil {
		cmd.say(fmt.Sprintf("Error Stopping App: %s", err))
		return
	}

	cmd.say("App Stopped Successfully")
}

func (cmd *appRunnerCommand) pollAppUntilUp(name string) {
	cmd.say("Starting App: " + name)
	startingTime := time.Now()
	for startingTime.Add(cmd.timeout).After(time.Now()) {
		if status, _ := cmd.appRunner.IsDockerAppUp(name); status {
			cmd.newLine()
			cmd.say(colors.Green(name + " is now running."))
			cmd.newLine()
			cmd.say(colors.Green(cmd.urlForApp(name)))
			return
		} else {
			cmd.dot()
		}
		time.Sleep(time.Second)
	}
	cmd.newLine()
	cmd.say(colors.Red(name + " took too long to start."))
}

func (cmd *appRunnerCommand) urlForApp(name string) string {
	return fmt.Sprintf("http://%s.%s", name, cmd.domain)
}

func (cmd *appRunnerCommand) say(output string) {
	cmd.output.Write([]byte(output))
}

func (cmd *appRunnerCommand) incorrectUsage(message string) {
	if len(message) > 0 {
		cmd.say("Incorrect Usage: " + message + "\n")
	} else {
		cmd.say("Incorrect Usage\n")
	}
}

func (cmd *appRunnerCommand) dot() {
	cmd.output.Write([]byte("."))
}

func (cmd *appRunnerCommand) newLine() {
	cmd.output.Write([]byte("\n"))
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
