package command_factory

import (
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type commandFactory struct {
	cmd *configCommand
}

func NewConfigCommandFactory(config *config.Config, output *output.Output) *commandFactory {
	return &commandFactory{&configCommand{config, output}}
}

func (c *commandFactory) MakeSetTargetCommand() cli.Command {
	var targetFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "username, u",
			Usage: "lattice username",
		},
		cli.StringFlag{
			Name:  "password, pw, p",
			Usage: "lattice password",
		},
	}

	var startCommand = cli.Command{
		Name:        "target",
		ShortName:   "t",
		Description: "set a target lattice location",
		Usage:       "ltc target DIEGO_DOMAIN [--username USERNAME --password PASSWORD]",
		Action:      c.cmd.setTarget,
		Flags:       targetFlags,
	}

	return startCommand
}

type configCommand struct {
	config *config.Config
	output *output.Output
}

func (cmd *configCommand) setTarget(context *cli.Context) {
	target := context.Args().First()
	username := context.String("username")
	password := context.String("password")

	if target == "" {
		cmd.output.IncorrectUsage("Target required.")
		return
	} else if username != "" && password == "" {
		cmd.output.IncorrectUsage("Password required with Username.")
		return
	} else if username == "" && password != "" {
		cmd.output.IncorrectUsage("Username required with Password.")
		return
	}

	err := cmd.config.SetTarget(target)
	if err != nil {
		cmd.say(err.Error())
		return
	}

	err = cmd.config.SetLogin(username, password)
	if err != nil {
		cmd.say(err.Error())
		return
	}

	cmd.say("Api Location Set")
}

func (cmd *configCommand) say(output string) {
	cmd.output.Write([]byte(output))
}

func (cmd *configCommand) incorrectUsage(message string) {
	if len(message) > 0 {
		cmd.say("Incorrect Usage: " + message)
	} else {
		cmd.say("Incorrect Usage")
	}
}
