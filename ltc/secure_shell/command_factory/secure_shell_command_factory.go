package command_factory

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

type SSHCommandFactory struct {
	config      *config_package.Config
	ui          terminal.UI
	exitHandler exit_handler.ExitHandler
	appExaminer app_examiner.AppExaminer

	secureShell SecureShell
}

//go:generate counterfeiter -o fake_secure_shell/fake_secure_shell.go . SecureShell
type SecureShell interface {
	ConnectToShell(appName string, instanceIndex int, command string, config *config_package.Config) error
}

func NewSSHCommandFactory(config *config_package.Config, ui terminal.UI, exitHandler exit_handler.ExitHandler, appExaminer app_examiner.AppExaminer, secureShell SecureShell) *SSHCommandFactory {
	return &SSHCommandFactory{config, ui, exitHandler, appExaminer, secureShell}
}

func (factory *SSHCommandFactory) MakeSSHCommand() cli.Command {
	return cli.Command{
		Name:        "ssh",
		Usage:       "Connects to a running app",
		Description: "ltc ssh APP_NAME [[--] optional command with args]\n\n   If a command is specified, no interactive shell will be provided.\n   A \"--\" token should be provided to avoid parsing of command flags.\n",
		Action:      factory.ssh,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "instance, i",
				Usage: "Connects to specified instance index",
				Value: 0,
			},
		},
	}
}

func (factory *SSHCommandFactory) ssh(context *cli.Context) {
	instanceIndex := context.Int("instance")
	appName := context.Args().First()

	if appName == "" {
		factory.ui.SayIncorrectUsage("")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	appInfo, err := factory.appExaminer.AppStatus(appName)
	if err != nil {
		factory.ui.SayLine("App " + appName + " not found.")
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}
	if instanceIndex < 0 || instanceIndex >= appInfo.ActualRunningInstances {
		factory.ui.SayLine(fmt.Sprintf("Instance %s/%d does not exist.", appName, instanceIndex))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	command := ""
	if len(context.Args()) > 1 {
		start := 1
		if context.Args().Get(1) == "--" {
			start = 2
		}
		command = strings.Join(context.Args()[start:len(context.Args())], " ")
	}

	if command == "" {
		factory.ui.SayLine("Connecting to %s/%d at %s", appName, instanceIndex, factory.config.Target())
	}

	if err := factory.secureShell.ConnectToShell(appName, instanceIndex, command, factory.config); err != nil {
		factory.ui.SayLine(fmt.Sprintf("Error connecting to %s/%d: %s", appName, instanceIndex, err.Error()))
		factory.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}
}
