package logs

import "github.com/cloudfoundry/sonde-go/events"

type LogReader interface {
	TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error))
	StopTailing()
}

type logConsumer interface {
	TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error)
	Close() error
}

type logReader struct {
	consumer logConsumer
}

func NewLogReader(consumer logConsumer) LogReader {
	return &logReader{
		consumer: consumer,
	}
}

func (l *logReader) TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error)) {
	outputChan := make(chan *events.LogMessage, 10)
	errorChan := make(chan error, 10)

	go l.consumer.TailingLogs(appGuid, "", outputChan, errorChan)

	l.readChannels(outputChan, errorChan, logCallback, errorCallback)
}

func (l *logReader) StopTailing() {
	l.consumer.Close()
}

func (l *logReader) readChannels(outputChan <-chan *events.LogMessage, errorChan <-chan error, logCallback func(*events.LogMessage), errorCallback func(error)) {
	for {
		select {
		case err := <-errorChan:
			errorCallback(err)
		case logMessage, ok := (<-outputChan):
			if !ok {
				return
			}
			logCallback(logMessage)
		}
	}
}
