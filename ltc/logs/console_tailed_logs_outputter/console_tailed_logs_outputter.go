package console_tailed_logs_outputter

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/prettify"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry/sonde-go/events"
)

type TailedLogsOutputter interface {
	OutputDebugLogs(pretty bool)
	OutputTailedLogs(appGuid string)
	StopOutputting()
}

type ConsoleTailedLogsOutputter struct {
	outputChan chan string
	ui         terminal.UI
	logReader  logs.LogReader
}

func NewConsoleTailedLogsOutputter(ui terminal.UI, logReader logs.LogReader) *ConsoleTailedLogsOutputter {
	return &ConsoleTailedLogsOutputter{
		outputChan: make(chan string, 10),
		ui:         ui,
		logReader:  logReader,
	}

}

func (ctlo *ConsoleTailedLogsOutputter) OutputDebugLogs(pretty bool) {
	if pretty {
		go ctlo.logReader.TailLogs(reserved_app_ids.LatticeDebugLogStreamAppId, ctlo.prettyDebugLogCallback, ctlo.prettyDebugErrorCallback)
	} else {
		go ctlo.logReader.TailLogs(reserved_app_ids.LatticeDebugLogStreamAppId, ctlo.rawDebugLogCallback, ctlo.rawDebugErrorCallback)
	}
	for log := range ctlo.outputChan {
		ctlo.ui.SayLine(log)
	}
}

func (ctlo *ConsoleTailedLogsOutputter) OutputTailedLogs(appGuid string) {
	go ctlo.logReader.TailLogs(appGuid, ctlo.logCallback, ctlo.errorCallback)

	for log := range ctlo.outputChan {
		ctlo.ui.SayLine(log)
	}
}

func (ctlo *ConsoleTailedLogsOutputter) StopOutputting() {
	ctlo.logReader.StopTailing()
}

func (ctlo *ConsoleTailedLogsOutputter) logCallback(log *events.LogMessage) {
	timeString := time.Unix(0, log.GetTimestamp()).Format("01/02 15:04:05.00")
	logOutput := fmt.Sprintf("%s [%s|%s] %s", colors.Cyan(timeString), colors.Yellow(log.GetSourceType()), colors.Yellow(log.GetSourceInstance()), log.GetMessage())
	ctlo.outputChan <- logOutput
}

func (ctlo *ConsoleTailedLogsOutputter) errorCallback(err error) {
	ctlo.outputChan <- err.Error()
}

func (ctlo *ConsoleTailedLogsOutputter) prettyDebugLogCallback(log *events.LogMessage) {
	ctlo.outputChan <- prettify.Prettify(log)
}

func (ctlo *ConsoleTailedLogsOutputter) prettyDebugErrorCallback(err error) {
	ctlo.outputChan <- err.Error()
}

func (ctlo *ConsoleTailedLogsOutputter) rawDebugLogCallback(log *events.LogMessage) {
	timeString := time.Unix(0, log.GetTimestamp()).Format("01/02 15:04:05.00")
	logOutput := fmt.Sprintf("%s [%s|%s] %s", timeString, log.GetSourceType(), log.GetSourceInstance(), log.GetMessage())
	ctlo.outputChan <- logOutput
}

func (ctlo *ConsoleTailedLogsOutputter) rawDebugErrorCallback(err error) {
	ctlo.outputChan <- err.Error()
}
