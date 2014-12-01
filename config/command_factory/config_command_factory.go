package command_factory

import (
	"io"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
)

type commandFactory struct {
	cmd *configCommand
}

func NewConfigCommandFactory(config *config.Config, output io.Writer) *commandFactory {
	return &commandFactory{&configCommand{config, output}}
}

func (c *commandFactory) MakeSetTargetCommand() cli.Command {
	var startCommand = cli.Command{
		Name:        "target",
		ShortName:   "t",
		Description: "Set a target diego location",
		Usage:       "diego-edge-cli target DIEGO_DOMAIN",
		Action:      c.cmd.setTarget,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

type configCommand struct {
	config *config.Config
	output io.Writer
}

func (cmd *configCommand) setTarget(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		cmd.say("Incorrect Usage\n")
		return
	}

	err := cmd.config.SetTarget(target)
	if err != nil {
		cmd.say(err.Error())
		return
	}

	cmd.say("Api Location Set\n")
}

func (cmd *configCommand) say(output string) {
	cmd.output.Write([]byte(output))
}
