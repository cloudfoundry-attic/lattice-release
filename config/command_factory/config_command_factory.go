package command_factory

import (
	"bufio"
	"io"
	"strings"

	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

type commandFactory struct {
	cmd *configCommand
}

func NewConfigCommandFactory(config *config.Config, targetVerifier target_verifier.TargetVerifier, input io.Reader, output *output.Output) *commandFactory {
	return &commandFactory{&configCommand{config, input, output, targetVerifier}}
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
	config         *config.Config
	input          io.Reader
	output         *output.Output
	targetVerifier target_verifier.TargetVerifier
}

func (cmd *configCommand) setTarget(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		cmd.output.IncorrectUsage("Target required.")
		return
	}

	cmd.config.SetTarget(target)
	cmd.config.SetLogin("", "")

	if ok, err := cmd.targetVerifier.ValidateAuthorization(cmd.config.Receptor()); err != nil {
		cmd.output.Say("Error verifying target: " + err.Error())
		return
	} else if ok {
		cmd.save()
		return
	}

	username := cmd.prompt("Username: ")
	password := cmd.prompt("Password: ")

	cmd.config.SetLogin(username, password)
	if ok, err := cmd.targetVerifier.ValidateAuthorization(cmd.config.Receptor()); err != nil {
		cmd.output.Say("Error verifying target: " + err.Error())
		return
	} else if !ok {
		cmd.output.Say("Could not authorize target.")
		return
	}

	cmd.save()
}

func (cmd *configCommand) save() {
	err := cmd.config.Save()
	if err != nil {
		cmd.output.Say(err.Error())
		return
	}

	cmd.output.Say("Api Location Set")
}

func (cmd *configCommand) prompt(promptText string) string {
	reader := bufio.NewReader(cmd.input)
	cmd.output.Say(promptText)

	result, _ := reader.ReadString('\n')

	return strings.TrimSuffix(result, "\n")
}

func (cmd *configCommand) incorrectUsage(message string) {
	if len(message) > 0 {
		cmd.output.Say("Incorrect Usage: " + message)
	} else {
		cmd.output.Say("Incorrect Usage")
	}
}
