package logs_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/noaa/events"
	"github.com/pivotal-cf-experimental/diego-edge-cli/logs"
)

type fakeConsumer struct {
	pendingLogs   []*events.LogMessage
	pendingErrors []error
}

func (consumer *fakeConsumer) TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error, stopChan chan struct{}) {
	for _, logMessage := range consumer.pendingLogs {
		outputChan <- logMessage
	}

	for _, err := range consumer.pendingErrors {
		errorChan <- err
	}
	close(errorChan)
}

func (consumer *fakeConsumer) addToPendingLogs(logMessage *events.LogMessage) {
	consumer.pendingLogs = append(consumer.pendingLogs, logMessage)
}

func (consumer *fakeConsumer) addToPendingErrors(err error) {
	consumer.pendingErrors = append(consumer.pendingErrors, err)
}

var _ = Describe("logs", func() {
	Describe("tailing logs", func() {
		It("uses the logMessage callback", func() {
			consumer := &fakeConsumer{pendingLogs: []*events.LogMessage{}, pendingErrors: []error{}}
			logReader := logs.NewLogReader(consumer)

			logMessageOne := &events.LogMessage{Message: []byte("Message 1")}
			consumer.addToPendingLogs(logMessageOne)

			logMessageTwo := &events.LogMessage{Message: []byte("Message 2")}
			consumer.addToPendingLogs(logMessageTwo)

			receivedLogs := []*events.LogMessage{}
			responseFunc := func(log *events.LogMessage) {
				receivedLogs = append(receivedLogs, log)
			}

			logReader.TailLogs("app-guid", responseFunc, nil)

			Expect(receivedLogs).To(Equal([]*events.LogMessage{logMessageOne, logMessageTwo}))
		})

		It("uses the logMessage callback", func() {
			consumer := &fakeConsumer{pendingErrors: []error{}}
			logReader := logs.NewLogReader(consumer)

			consumer.addToPendingErrors(errors.New("error 1"))
			consumer.addToPendingErrors(errors.New("error 2"))

			errorsFromCallback := []error{}
			errorFunc := func(err error) {
				errorsFromCallback = append(errorsFromCallback, err)
			}

			logReader.TailLogs("app-guid", nil, errorFunc)

			Expect(errorsFromCallback).To(Equal([]error{errors.New("error 1"), errors.New("error 2")}))
		})
	})

})
