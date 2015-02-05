package logs

import (
	"github.com/cloudfoundry/noaa/events"
)

type LogReader interface {
	TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error), stopChan chan struct{})
}

type logConsumer interface {
	TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error, stopChan chan struct{})
}

type logReader struct {
	consumer logConsumer
}

func NewLogReader(consumer logConsumer) LogReader {
	return &logReader{consumer: consumer}
}

func (l *logReader) TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error), stopChan chan struct{}) {
	outputChan := make(chan *events.LogMessage, 10)
	errorChan := make(chan error, 10)

	go l.consumer.TailingLogs(appGuid, "", outputChan, errorChan, stopChan)

	go func() {
		readErrorChannel(errorChan, errorCallback)
		close(outputChan)
	}()

	readOutputChannel(outputChan, logCallback)
}

func readOutputChannel(outputChan <-chan *events.LogMessage, callback func(*events.LogMessage)) {
	for logMessage := range outputChan {
		callback(logMessage)
	}
}

func readErrorChannel(errorChan <-chan error, callback func(error)) {
	for err := range errorChan {
		callback(err)
	}
}
