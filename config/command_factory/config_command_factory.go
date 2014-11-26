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
		Description: "Set a target diego location",
		Usage:       "diego-edge-cli target DIEGO_DOMAIN",
		Action:      c.setTarget,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (c *commandFactory) setTarget(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		c.say("Incorrect Usage\n")
		return
	}

	err := c.config.SetTarget(target)
	if err != nil {
		c.say(err.Error())
		return
	}

	c.say("Api Location Set\n")
}

func (c *commandFactory) say(output string) {
	c.output.Write([]byte(output))
}
