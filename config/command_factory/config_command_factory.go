package command_factory

import (
	"io"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
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
		Description: "set a target diego location",
		Usage:       "ltc target DIEGO_DOMAIN",
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
		cmd.say("Incorrect Usage")
		return
	}

	err := cmd.config.SetTarget(target)
	if err != nil {
		cmd.say(err.Error())
		return
	}

	cmd.say("Api Location Set")
}

func (cmd *configCommand) say(output string) {
	cmd.output.Write([]byte(output))
}
