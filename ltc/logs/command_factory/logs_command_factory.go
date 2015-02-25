package command_factory

import (
	"github.com/cloudfoundry-incubator/lattice/cli/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/cli/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/cli/output"
	"github.com/codegangsta/cli"
)

type logsCommandFactory struct {
	cmd *logsCommand
}

func NewLogsCommandFactory(output *output.Output, tailedLogsOutputter console_tailed_logs_outputter.TailedLogsOutputter, exitHandler exit_handler.ExitHandler) *logsCommandFactory {
	return &logsCommandFactory{
		&logsCommand{
			output:              output,
			tailedLogsOutputter: tailedLogsOutputter,
			exitHandler:         exitHandler,
		},
	}
}

func (factory *logsCommandFactory) MakeLogsCommand() cli.Command {
	var logsCommand = cli.Command{
		Name:        "logs",
		ShortName:   "l",
		Description: "Stream logs from the specified application",
		Usage:       "ltc logs APP_NAME",
		Action:      factory.cmd.tailLogs,
		Flags:       []cli.Flag{},
	}

	return logsCommand
}

type logsCommand struct {
	output              *output.Output
	tailedLogsOutputter console_tailed_logs_outputter.TailedLogsOutputter
	exitHandler         exit_handler.ExitHandler
}

func (cmd *logsCommand) tailLogs(context *cli.Context) {
	appGuid := context.Args().First()

	if appGuid == "" {
		cmd.output.IncorrectUsage("")
		return
	}

	cmd.tailedLogsOutputter.OutputTailedLogs(appGuid)
}
