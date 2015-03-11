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

type ConfigCommandFactory struct {
	config         *config.Config
	input          io.Reader
	output         *output.Output
	targetVerifier target_verifier.TargetVerifier
	exitHandler    exit_handler.ExitHandler
}

func NewConfigCommandFactory(config *config.Config, targetVerifier target_verifier.TargetVerifier, input io.Reader, output *output.Output, exitHandler exit_handler.ExitHandler) *ConfigCommandFactory {
	return &ConfigCommandFactory{config, input, output, targetVerifier, exitHandler}
}

func (factory *ConfigCommandFactory) MakeTargetCommand() cli.Command {
	var startCommand = cli.Command{
		Name:      TargetCommandName,
		ShortName: "t",
		Description: `Set a target lattice location.

   For a Vagrant deployed Lattice:
   ltc target 192.168.11.11.xip.io`,
		Usage:  "ltc target LATTICE_TARGET",
		Action: factory.target,
	}

	return startCommand
}

func (factory *ConfigCommandFactory) target(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		factory.printTarget()
		return
	}

	factory.config.SetTarget(target)
	factory.config.SetLogin("", "")

	if _, authorized, err := factory.targetVerifier.VerifyTarget(factory.config.Receptor()); err != nil {
		factory.output.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if authorized {
		factory.save()
		return
	}

	username := factory.prompt("Username: ")
	password := factory.prompt("Password: ")

	factory.config.SetLogin(username, password)
	if _, authorized, err := factory.targetVerifier.VerifyTarget(factory.config.Receptor()); err != nil {
		factory.output.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if !authorized {
		factory.output.Say("Could not authorize target.")
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	factory.save()
}

func (factory *ConfigCommandFactory) save() {
	err := factory.config.Save()
	if err != nil {
		factory.output.Say(err.Error())
		return
	}

	factory.output.Say("Api Location Set")
}

func (factory *ConfigCommandFactory) prompt(promptText string) string {
	reader := bufio.NewReader(factory.input)
	factory.output.Say(promptText)

	result, _ := reader.ReadString('\n')

	return strings.TrimSuffix(result, "\n")
}

func (factory *ConfigCommandFactory) printTarget() {
	if factory.config.Target() == "" {
		factory.output.Say("Target not set.")
		return
	}
	factory.output.Say(fmt.Sprintf("Target:\t\t%s", factory.config.Target()))

	if factory.config.Username() != "" {
		factory.output.Say(fmt.Sprintf("\nUsername:\t%s", factory.config.Username()))
	}
}
