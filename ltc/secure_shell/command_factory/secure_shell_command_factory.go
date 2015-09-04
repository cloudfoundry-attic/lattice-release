package command_factory

import (
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

	secureShell SecureShell
}

//go:generate counterfeiter -o fake_secure_shell/fake_secure_shell.go . SecureShell
type SecureShell interface {
	ConnectToShell(appName string, instanceIndex int, config *config_package.Config) error
}

func NewSSHCommandFactory(config *config_package.Config, ui terminal.UI, exitHandler exit_handler.ExitHandler, secureShell SecureShell) *SSHCommandFactory {
	return &SSHCommandFactory{config, ui, exitHandler, secureShell}
}

func (factory *SSHCommandFactory) MakeSSHCommand() cli.Command {
	return cli.Command{
		Name:        "ssh",
		Usage:       "Connects to a running app",
		Description: "ltc ssh APP_NAME",
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

	factory.ui.SayLine("Connecting to %s/%d at %s", appName, instanceIndex, factory.config.Target())

	factory.secureShell.ConnectToShell(appName, instanceIndex, factory.config)
}
