package command_factory

import (
	"fmt"
	"io"
	"time"

	"github.com/cloudfoundry/noaa/events"
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/colors"
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
		Description: "stream logs from an app",
		Usage:       "diego-edge-cli logs APP_NAME",
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
		cmd.say("Incorrect Usage")
		return
	}

	go cmd.logReader.TailLogs(appGuid, cmd.logCallback, cmd.errorCallback)

	for log := range cmd.outputChan {
		cmd.say(log + "\n")
	}
}

func (cmd *logsCommand) logCallback(log *events.LogMessage) {
	timeString := time.Unix(0, log.GetTimestamp()).Format("02 Jan 15:04")
	logOutput := fmt.Sprintf("%s [%s|%s] %s", colors.Cyan(timeString), colors.Yellow(log.GetSourceType()), colors.Yellow(log.GetSourceInstance()), log.GetMessage())
	cmd.outputChan <- logOutput
}

func (cmd *logsCommand) errorCallback(err error) {
	cmd.outputChan <- err.Error()
}

func (cmd *logsCommand) say(output string) {
	cmd.output.Write([]byte(output))
}
