package command_factory

import (
	"io"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/logs"
)

type logsCommandFactory struct {
	logReader  logs.LogReader
	output     io.Writer
	outputChan chan string
}

func NewLogsCommandFactory(logReader logs.LogReader, output io.Writer) *logsCommandFactory {
	outputChan := make(chan string, 10)
	return &logsCommandFactory{logReader, output, outputChan}
}

func (factory *logsCommandFactory) MakeLogsCommand() cli.Command {
	var logsCommand = cli.Command{
		Name:        "logs",
		ShortName:   "l",
		Description: "",
		Usage:       "",
		Action:      factory.tailLogs,
		Flags:       []cli.Flag{},
	}

	return logsCommand
}

func (factory *logsCommandFactory) tailLogs(context *cli.Context) {
	appGuid := context.Args().First()

	if appGuid == "" {
		factory.say("Invalid Usage\n")
		return
	}

	go factory.logReader.TailLogs(appGuid, factory.logCallback)

	for log := range factory.outputChan {
		factory.say(log + "\n")
	}
}

func (factory *logsCommandFactory) logCallback(log string) {
	factory.outputChan <- log
}

func (factory *logsCommandFactory) say(output string) {
	factory.output.Write([]byte(output))
}
