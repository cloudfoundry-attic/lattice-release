package command_factory

import (
	"io"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/logs"
)

type logsCommandFactory struct {
	cmd *logsCommand
}

func NewLogsCommandFactory(logReader logs.LogReader, output io.Writer) *logsCommandFactory {
	outputChan := make(chan string, 10)
	return &logsCommandFactory{&logsCommand{logReader, output, outputChan}}
}

func (factory *logsCommandFactory) MakeLogsCommand() cli.Command {
	var logsCommand = cli.Command{
		Name:        "logs",
		ShortName:   "l",
		Description: "",
		Usage:       "",
		Action:      factory.cmd.tailLogs,
		Flags:       []cli.Flag{},
	}

	return logsCommand
}

type logsCommand struct {
	logReader  logs.LogReader
	output     io.Writer
	outputChan chan string
}

func (cmd *logsCommand) tailLogs(context *cli.Context) {
	appGuid := context.Args().First()

	if appGuid == "" {
		cmd.say("Invalid Usage\n")
		return
	}

	go cmd.logReader.TailLogs(appGuid, cmd.logCallback)

	for log := range cmd.outputChan {
		cmd.say(log + "\n")
	}
}

func (cmd *logsCommand) logCallback(log string) {
	cmd.outputChan <- log
}

func (cmd *logsCommand) say(output string) {
	cmd.output.Write([]byte(output))
}
