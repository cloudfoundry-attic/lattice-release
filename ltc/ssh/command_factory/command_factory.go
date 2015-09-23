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

	secureShell SSH
}

//go:generate counterfeiter -o mocks/fake_ssh.go . SSH
type SSH interface {
	ConnectAndForward(appName string, instanceIndex int, localAddress, remoteAddress string, config *config_package.Config) error
	ConnectToShell(appName string, instanceIndex int, command string, config *config_package.Config) error
}

func NewSSHCommandFactory(config *config_package.Config, ui terminal.UI, exitHandler exit_handler.ExitHandler, appExaminer app_examiner.AppExaminer, secureShell SSH) *SSHCommandFactory {
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
			cli.StringFlag{
				Name:  "L",
				Usage: "Listens on specified local address/port and forwards connections to specified remote address/port\n     \te.g. ltc ssh APP_NAME -L [localhost:]1234:remotehost:5678",
			},
		},
	}
}

func (factory *SSHCommandFactory) ssh(context *cli.Context) {
	instanceIndex := context.Int("instance")
	localForward := context.String("L")

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

	if localForward != "" {
		var localHost, localPort, remoteHost, remotePort string

		parts := strings.Split(localForward, ":")

		switch len(parts) {
		case 3:
			localHost, localPort, remoteHost, remotePort = "localhost", parts[0], parts[1], parts[2]
		case 4:
			localHost, localPort, remoteHost, remotePort = parts[0], parts[1], parts[2], parts[3]
		default:
			factory.ui.SayIncorrectUsage("-L expects [localhost:]localport:remotehost:remoteport")
			factory.exitHandler.Exit(exit_codes.InvalidSyntax)
			return
		}

		localAddr := fmt.Sprintf("%s:%s", localHost, localPort)
		remoteAddr := fmt.Sprintf("%s:%s", remoteHost, remotePort)

		factory.ui.SayLine("Forwarding %s to %s via %s/%d at %s", localAddr, remoteAddr, appName, instanceIndex, factory.config.Target())

		if err := factory.secureShell.ConnectAndForward(appName, instanceIndex, localAddr, remoteAddr, factory.config); err != nil {
			factory.ui.SayLine(fmt.Sprintf("Error connecting to %s/%d: %s", appName, instanceIndex, err.Error()))
			factory.exitHandler.Exit(exit_codes.CommandFailed)
			return
		}
	} else {
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
}
