package command_factory

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

type logsCommandFactory struct {
	appExaminer         app_examiner.AppExaminer
	ui                  terminal.UI
	tailedLogsOutputter console_tailed_logs_outputter.TailedLogsOutputter
	exitHandler         exit_handler.ExitHandler
}

func NewLogsCommandFactory(appExaminer app_examiner.AppExaminer, ui terminal.UI, tailedLogsOutputter console_tailed_logs_outputter.TailedLogsOutputter, exitHandler exit_handler.ExitHandler) *logsCommandFactory {
	return &logsCommandFactory{
		appExaminer:         appExaminer,
		ui:                  ui,
		tailedLogsOutputter: tailedLogsOutputter,
		exitHandler:         exitHandler,
	}
}

func (factory *logsCommandFactory) MakeLogsCommand() cli.Command {
	var logsCommand = cli.Command{
		Name:        "logs",
		Aliases:     []string{"lo"},
		Usage:       "Streams logs from the specified application",
		Description: "ltc logs APP_NAME",
		Action:      factory.tailLogs,
		Flags:       []cli.Flag{},
	}

	return logsCommand
}

func (factory *logsCommandFactory) MakeDebugLogsCommand() cli.Command {
	return cli.Command{
		Name:        "debug-logs",
		Aliases:     []string{"dl"},
		Usage:       "Streams logs from the lattice cluster components",
		Description: "ltc debug-logs",
		Action:      factory.tailDebugLogs,
	}
}

func (factory *logsCommandFactory) tailLogs(context *cli.Context) {
	appGuid := context.Args().First()

	if appGuid == "" {
		factory.ui.IncorrectUsage("")
		return
	}

	if appExists, err := factory.appExaminer.AppExists(appGuid); err != nil {
		factory.ui.SayLine(fmt.Sprintf("Error: %s", err.Error()))
		return
	} else if !appExists {
		factory.ui.SayLine(fmt.Sprintf("Application %s not found.", appGuid))
		factory.ui.SayLine(fmt.Sprintf("Tailing logs and waiting for %s to appear...", appGuid))
	}

	factory.tailedLogsOutputter.OutputTailedLogs(appGuid)
}

func (factory *logsCommandFactory) tailDebugLogs(context *cli.Context) {
	factory.tailedLogsOutputter.OutputTailedLogs(reserved_app_ids.LatticeDebugLogStreamAppId)
}
