package command_factory

import (
	"io"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
)

type commandFactory struct {
	config *config.Config
	output io.Writer
}

func NewConfigCommandFactory(config *config.Config, output io.Writer) *commandFactory {
	return &commandFactory{config, output}
}

func (c *commandFactory) MakeSetTargetCommand() cli.Command {
	var startCommand = cli.Command{
		Name:        "target",
		ShortName:   "t",
		Description: "Set a target api location",
		Usage:       "diego-edge-cli target API_PATH",
		Action:      c.setTarget,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (c *commandFactory) MakeSetTargetLoggregatorCommand() cli.Command {
	var startCommand = cli.Command{
		Name:        "target-loggregator",
		Description: "Set a target loggregator api location",
		Usage:       "diego-edge-cli target-loggregator LOGGREGATOR_PATH",
		Action:      c.setLoggregatorTarget,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (c *commandFactory) setTarget(context *cli.Context) {
	api := context.Args().First()

	if api == "" {
		c.say("Incorrect Usage\n")
		return
	}

	err := c.config.SetApi(api)
	if err != nil {
		c.say(err.Error())
		return
	}

	c.say("Api Location Set\n")
}

func (c *commandFactory) setLoggregatorTarget(context *cli.Context) {
	loggregator := context.Args().First()

	if loggregator == "" {
		c.say("Incorrect Usage\n")
		return
	}

	err := c.config.SetLoggregator(loggregator)
	if err != nil {
		c.say(err.Error())
		return
	}

	c.say("Loggregator Api Location Set\n")
}

func (c *commandFactory) say(output string) {
	c.output.Write([]byte(output))
}
