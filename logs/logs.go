package logs

import (
	"github.com/cloudfoundry/noaa/events"
)

type LogReader interface {
	TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error))
	StopTailing()
}

type logConsumer interface {
	TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error, stopChan chan struct{})
}

type logReader struct {
	consumer logConsumer
	stopChan chan struct{}
}

func NewLogReader(consumer logConsumer) LogReader {
	return &logReader{
		consumer: consumer,
		stopChan: make(chan struct{}),
	}
}

func (l *logReader) TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error)) {
	outputChan := make(chan *events.LogMessage, 10)
	errorChan := make(chan error, 10)

	go l.consumer.TailingLogs(appGuid, "", outputChan, errorChan, l.stopChan)

	l.readChannels(outputChan, errorChan, logCallback, errorCallback)

	close(l.stopChan)
	close(errorChan)
	close(outputChan)
}

func (l *logReader) StopTailing() {
	l.stopChan <- struct{}{}
}

func (l *logReader) readChannels(outputChan <-chan *events.LogMessage, errorChan <-chan error, logCallback func(*events.LogMessage), errorCallback func(error)) {
	for {
		select {
		case <-l.stopChan:
			return
		case err := <-errorChan:
			errorCallback(err)
		case logMessage := <-outputChan:
			logCallback(logMessage)
		}
	}
}
