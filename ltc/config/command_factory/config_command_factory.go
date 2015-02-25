package command_factory

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/codegangsta/cli"
)

const TargetCommandName = "target"

type commandFactory struct {
	cmd *configCommand
}

func NewConfigCommandFactory(config *config.Config, targetVerifier target_verifier.TargetVerifier, input io.Reader, output *output.Output, exitHandler exit_handler.ExitHandler) *commandFactory {
	return &commandFactory{&configCommand{config, input, output, targetVerifier, exitHandler}}
}

func (c *commandFactory) MakeTargetCommand() cli.Command {
	var startCommand = cli.Command{
		Name:      TargetCommandName,
		ShortName: "t",
		Description: `Set a target lattice location.

   For a Vagrant deployed Lattice:
   ltc target 192.168.11.11.xip.io`,
		Usage:  "ltc target LATTICE_TARGET",
		Action: c.cmd.target,
	}

	return startCommand
}

type configCommand struct {
	config         *config.Config
	input          io.Reader
	output         *output.Output
	targetVerifier target_verifier.TargetVerifier
	exitHandler    exit_handler.ExitHandler
}

func (cmd *configCommand) target(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		cmd.printTarget()
		return
	}

	cmd.config.SetTarget(target)
	cmd.config.SetLogin("", "")

	if _, authorized, err := cmd.targetVerifier.VerifyTarget(cmd.config.Receptor()); err != nil {
		cmd.output.Say("Error verifying target: " + err.Error())
		cmd.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if authorized {
		cmd.save()
		return
	}

	username := cmd.prompt("Username: ")
	password := cmd.prompt("Password: ")

	cmd.config.SetLogin(username, password)
	if _, authorized, err := cmd.targetVerifier.VerifyTarget(cmd.config.Receptor()); err != nil {
		cmd.output.Say("Error verifying target: " + err.Error())
		cmd.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if !authorized {
		cmd.output.Say("Could not authorize target.")
		cmd.exitHandler.Exit(exit_codes.BadTarget)
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

func (cmd *configCommand) printTarget() {
	if cmd.config.Target() == "" {
		cmd.output.Say("Target not set.")
		return
	}
	cmd.output.Say(fmt.Sprintf("Target:\t\t%s", cmd.config.Target()))

	if cmd.config.Username() != "" {
		cmd.output.Say(fmt.Sprintf("\nUsername:\t%s", cmd.config.Username()))
	}
}
