package logs

import (
	"github.com/cloudfoundry/noaa/events"
)

type LogReader interface {
	TailLogs(appGuid string, callback func(string))
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

func (l *logReader) TailLogs(appGuid string, callback func(string)) {
	outputChan := make(chan *events.LogMessage, 10)
	errorChan := make(chan error, 10)

	go l.consumer.TailingLogs(appGuid, "", outputChan, errorChan, nil)

	go func() {
		readErrorChannel(errorChan, callback)
		close(outputChan)
	}()

	readOutputChannel(outputChan, callback)
}

func readOutputChannel(outputChan <-chan *events.LogMessage, callback func(string)) {
	for logMessage := range outputChan {
		callback(string(logMessage.Message))
	}
}

func readErrorChannel(errorChan <-chan error, callback func(string)) {
	for err := range errorChan {
		callback(err.Error())
	}
}
