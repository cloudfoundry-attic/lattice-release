package command_factory

import (
	"fmt"
	"io"

	"github.com/codegangsta/cli"
)

type appRunner interface {
	StartDockerApp(name, startCommand, dockerImagePath string) error
}

type StartAppCommandFactory struct {
	appRunner appRunner
	output    io.Writer
}

func NewStartAppCommandFactory(appRunner appRunner, output io.Writer) *StartAppCommandFactory {
	return &StartAppCommandFactory{appRunner, output}
}

func (commandFactory *StartAppCommandFactory) MakeCommand() cli.Command {

	var startFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-image, i",
			Usage: "the docker image to run",
		},
		cli.StringFlag{
			Name:  "start-command, c",
			Usage: "the command to run in the context of the docker image (ie the start command for the app)",
		},
	}

	var startCommand = cli.Command{
		Name:      "start",
		ShortName: "s",
		Description: "Start a docker app on diego",
		Usage:     "diego-edge-cli start APP_NAME [-i DOCKER_IMAGE] [-C START_COMMAND]",
		Action:    commandFactory.startDiegoApp,
		Flags:     startFlags,
	}

	return startCommand
}

func (commandFactory *StartAppCommandFactory) startDiegoApp(c *cli.Context) {
	startCommand := c.String("start-command")
	dockerImage := c.String("docker-image")
	name := c.Args().First()

	if name == "" || dockerImage == "" || startCommand == "" {
		commandFactory.output.Write([]byte("Incorrect Usage\n"))
		return
	}

	err := commandFactory.appRunner.StartDockerApp(name, startCommand, dockerImage)

	if err != nil {
		commandFactory.output.Write([]byte(fmt.Sprintf("Error Starting App: %s", err.Error())))
		return
	}

	commandFactory.output.Write([]byte("App Staged Successfully"))
}
