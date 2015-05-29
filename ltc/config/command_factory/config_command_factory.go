package command_factory

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

const TargetCommandName = "target"

type ConfigCommandFactory struct {
	config         *config.Config
	ui             terminal.UI
	targetVerifier target_verifier.TargetVerifier
	exitHandler    exit_handler.ExitHandler
}

func NewConfigCommandFactory(config *config.Config, ui terminal.UI, targetVerifier target_verifier.TargetVerifier, exitHandler exit_handler.ExitHandler) *ConfigCommandFactory {
	return &ConfigCommandFactory{config, ui, targetVerifier, exitHandler}
}

func (factory *ConfigCommandFactory) MakeTargetCommand() cli.Command {
	var startCommand = cli.Command{
		Name:        TargetCommandName,
		Aliases:     []string{"ta"},
		Usage:       "Targets a lattice cluster",
		Description: "ltc target TARGET (e.g., 192.168.11.11.xip.io)",
		Action:      factory.target,
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
		factory.ui.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if authorized {
		factory.save()
		return
	}

	username := factory.ui.Prompt("Username: ")
	password := factory.ui.PromptForPassword("Password: ")

	factory.config.SetLogin(username, password)
	if _, authorized, err := factory.targetVerifier.VerifyTarget(factory.config.Receptor()); err != nil {
		factory.ui.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	} else if !authorized {
		factory.ui.Say("Could not authorize target.")
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	factory.save()
}

func (factory *ConfigCommandFactory) save() {
	err := factory.config.Save()
	if err != nil {
		factory.ui.Say(err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	factory.ui.Say("Api Location Set")
}

func (factory *ConfigCommandFactory) printTarget() {
	if factory.config.Target() == "" {
		factory.ui.Say("Target not set.")
		return
	}
	factory.ui.Say(fmt.Sprintf("Target:\t\t%s", factory.config.Target()))

	if factory.config.Username() != "" {
		factory.ui.Say(fmt.Sprintf("\nUsername:\t%s", factory.config.Username()))
	}
}
