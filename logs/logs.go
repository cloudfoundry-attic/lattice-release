package logs

import (
	"github.com/cloudfoundry/noaa/events"
)

type logCallbackFunc func(string)

type logConsumer interface {
	TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error, stopChan chan struct{})
}

type logReader struct {
	consumer logConsumer
}

func NewLogReader(consumer logConsumer) *logReader {
	return &logReader{consumer: consumer}
}

func (l *logReader) TailLogs(appGuid string, callback logCallbackFunc) {
	outputChan := make(chan *events.LogMessage, 10)
	errorChan := make(chan error, 10)

	go l.consumer.TailingLogs(appGuid, "", outputChan, errorChan, nil)

	go func() {
		readErrorChannel(errorChan, callback)
		close(outputChan)
	}()

	readOutputChannel(outputChan, callback)
}

func readOutputChannel(outputChan <-chan *events.LogMessage, callback logCallbackFunc) {
	for logMessage := range outputChan {
		callback(string(logMessage.Message))
	}
}

func readErrorChannel(errorChan <-chan error, callback logCallbackFunc) {
	for err := range errorChan {
		callback(err.Error())
	}
}
