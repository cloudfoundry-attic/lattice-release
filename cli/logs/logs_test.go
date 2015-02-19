package logs_test

import (
	"errors"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/noaa/events"
	"github.com/pivotal-cf-experimental/lattice-cli/logs"
)

func NewFakeConsumer() *fakeConsumer {
	return &fakeConsumer{
		inboundLogStream:   make(chan *events.LogMessage),
		inboundErrorStream: make(chan error),
	}
}

type fakeConsumer struct {
	inboundLogStream   chan *events.LogMessage
	inboundErrorStream chan error
}

func (consumer *fakeConsumer) TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			defer close(errorChan)
			return
		case err := <-consumer.inboundErrorStream:
			errorChan <- err
		case logMessage := <-consumer.inboundLogStream:
			outputChan <- logMessage
		}
	}
}

func (consumer *fakeConsumer) sendToInboundLogStream(logMessage *events.LogMessage) {
	consumer.inboundLogStream <- logMessage
}

func (consumer *fakeConsumer) sendToInboundErrorStream(err error) {
	consumer.inboundErrorStream <- err
}

type MessageReceiver struct {
	sync.RWMutex
	receivedMessages []*events.LogMessage
}

func (mr *MessageReceiver) AppendMessage(logMessage *events.LogMessage) {
	defer mr.Unlock()
	mr.Lock()
	mr.receivedMessages = append(mr.receivedMessages, logMessage)
}

func (mr *MessageReceiver) GetMessages() []*events.LogMessage {
	defer mr.RUnlock()
	mr.RLock()
	return mr.receivedMessages
}

type ErrorReceiver struct {
	sync.RWMutex
	receivedErrors []error
}

func (e *ErrorReceiver) AppendError(err error) {
	defer e.Unlock()
	e.Lock()
	e.receivedErrors = append(e.receivedErrors, err)
}

func (e *ErrorReceiver) GetErrors() []error {
	defer e.RUnlock()
	e.RLock()
	return e.receivedErrors
}

var _ = Describe("logs", func() {
	Describe("TailLogs", func() {
		var (
			consumer  *fakeConsumer
			logReader logs.LogReader
			stopChan  chan struct{}
		)
		BeforeEach(func() {
			consumer = NewFakeConsumer()
			logReader = logs.NewLogReader(consumer)
			stopChan = make(chan struct{})

		})

		It("provides the logCallback with logs until StopTailing is called", func() {
			messageReceiver := &MessageReceiver{}

			responseFunc := func(logMessage *events.LogMessage) {
				messageReceiver.AppendMessage(logMessage)
			}

			go logReader.TailLogs("app-guid", responseFunc, func(error) {})

			logMessageOne := &events.LogMessage{Message: []byte("Message 1")}
			go consumer.sendToInboundLogStream(logMessageOne)

			Eventually(messageReceiver.GetMessages, 3).Should(Equal([]*events.LogMessage{logMessageOne}))

			logMessageTwo := &events.LogMessage{Message: []byte("Message 2")}
			go consumer.sendToInboundLogStream(logMessageTwo)

			Eventually(messageReceiver.GetMessages).Should(Equal([]*events.LogMessage{logMessageOne, logMessageTwo}))

			logReader.StopTailing()

			logMessageThree := &events.LogMessage{Message: []byte("Message 3")}
			go consumer.sendToInboundLogStream(logMessageThree)

			Consistently(messageReceiver.GetMessages).ShouldNot(ContainElement(logMessageThree))

		})

		It("provides the errorCallback with the pending errors until StopTailing is called.", func() {

			errorReceiver := &ErrorReceiver{}

			errorFunc := func(err error) {
				errorReceiver.AppendError(err)
			}

			go logReader.TailLogs("app-guid", func(*events.LogMessage) {}, errorFunc)

			go consumer.sendToInboundErrorStream(errors.New("error 1"))
			Eventually(errorReceiver.GetErrors).Should(Equal([]error{errors.New("error 1")}))

			go consumer.sendToInboundErrorStream(errors.New("error 2"))

			Eventually(errorReceiver.GetErrors).Should(Equal([]error{errors.New("error 1"), errors.New("error 2")}))

			logReader.StopTailing()

			errorThree := errors.New("error 3")
			go consumer.sendToInboundErrorStream(errorThree)
			Consistently(errorReceiver.GetErrors).ShouldNot(ContainElement(errorThree))
		})
	})

})
