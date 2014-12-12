package command_factory

import (
	"bufio"
	"io"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type commandFactory struct {
	cmd *configCommand
}

func NewConfigCommandFactory(config *config.Config, input io.Reader, output *output.Output) *commandFactory {
	return &commandFactory{&configCommand{config, input, output}}
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
		Usage:       "ltc target LATTICE_DOMAIN [--username USERNAME --password PASSWORD]",
		Action:      c.cmd.setTarget,
		Flags:       targetFlags,
	}

	return startCommand
}

type configCommand struct {
	config *config.Config
	input  io.Reader
	output *output.Output
}

func (cmd *configCommand) setTarget(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		cmd.output.IncorrectUsage("Target required.")
		return
	}

	cmd.say("Username: ")
	reader := bufio.NewReader(cmd.input)
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1]

	cmd.say("Password: ")
	password, _ := reader.ReadString('\n')
	password = password[:len(password)-1]

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
