package logs_test

import (
	"errors"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs"
	"github.com/cloudfoundry/sonde-go/events"
)

func NewFakeConsumer() *fakeConsumer {
	return &fakeConsumer{
		inboundLogStream:   make(chan *events.LogMessage),
		inboundErrorStream: make(chan error),
		stopChan:           make(chan struct{}),
	}
}

type fakeConsumer struct {
	inboundLogStream   chan *events.LogMessage
	inboundErrorStream chan error
	stopChan           chan struct{}
}

func (consumer *fakeConsumer) TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
	for {
		select {
		case <-consumer.stopChan:
			defer close(errorChan)
			defer close(outputChan)
			return
		case err := <-consumer.inboundErrorStream:
			errorChan <- err
		case logMessage := <-consumer.inboundLogStream:
			outputChan <- logMessage
		}
	}
}

func (consumer *fakeConsumer) Close() error {
	defer close(consumer.stopChan)
	return nil
}

func (consumer *fakeConsumer) sendToInboundLogStream(logMessage *events.LogMessage) {
	consumer.inboundLogStream <- logMessage
}

func (consumer *fakeConsumer) sendToInboundErrorStream(err error) {
	consumer.inboundErrorStream <- err
}

type messageReceiver struct {
	sync.RWMutex
	receivedMessages []*events.LogMessage
}

func (mr *messageReceiver) AppendMessage(logMessage *events.LogMessage) {
	defer mr.Unlock()
	mr.Lock()
	mr.receivedMessages = append(mr.receivedMessages, logMessage)
}

func (mr *messageReceiver) GetMessages() []*events.LogMessage {
	defer mr.RUnlock()
	mr.RLock()
	return mr.receivedMessages
}

type errorReceiver struct {
	sync.RWMutex
	receivedErrors []error
}

func (e *errorReceiver) AppendError(err error) {
	defer e.Unlock()
	e.Lock()
	e.receivedErrors = append(e.receivedErrors, err)
}

func (e *errorReceiver) GetErrors() []error {
	defer e.RUnlock()
	e.RLock()
	return e.receivedErrors
}

var _ = Describe("Logs", func() {
	var (
		consumer  *fakeConsumer
		logReader logs.LogReader
	)

	BeforeEach(func() {
		consumer = NewFakeConsumer()
		logReader = logs.NewLogReader(consumer)
	})

	Describe("TailLogs", func() {

		It("provides the logCallback with logs until StopTailing is called", func() {
			messageReceiver := &messageReceiver{}

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
			errorReceiver := &errorReceiver{}

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

	Describe("StopTailing", func() {
		It("stops tailing logs when requested", func() {
			doneChan := make(chan struct{})
			go func() {
				defer GinkgoRecover()

				logReader.TailLogs("app-guid", func(*events.LogMessage) {}, func(error) {})
				close(doneChan)
			}()

			logReader.StopTailing()

			Eventually(doneChan).Should(BeClosed())
		})
	})

})
