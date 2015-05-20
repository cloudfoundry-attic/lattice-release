package command_factory

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
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
		Aliases:     []string{"lg", "lo"},
		Usage:       "Streams logs from the specified application",
		Description: "ltc logs APP_NAME",
		Action:      factory.tailLogs,
		Flags:       []cli.Flag{},
	}

	return logsCommand
}

func (factory *logsCommandFactory) MakeDebugLogsCommand() cli.Command {
	var debugLogsFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "raw, r",
			Usage: "Removes pretty formatting",
		},
	}
	return cli.Command{
		Name:    "debug-logs",
		Aliases: []string{"dl"},
		Usage:   "Streams logs from the lattice cluster components",
		// Description: "ltc debug-logs",
		Description: `ltc debug-logs [--raw]

   Output format is:

    [source|instance] [loglevel] timestamp session message summary
                                                   (error message)
                                                   (message detail)`,
		Action: factory.tailDebugLogs,
		Flags:  debugLogsFlags,
	}
}

func (factory *logsCommandFactory) tailLogs(context *cli.Context) {
	appGuid := context.Args().First()

	if appGuid == "" {
		factory.ui.SayIncorrectUsage("APP_NAME required")
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
	rawFlag := context.Bool("raw")
	factory.tailedLogsOutputter.OutputDebugLogs(!rawFlag)
}
