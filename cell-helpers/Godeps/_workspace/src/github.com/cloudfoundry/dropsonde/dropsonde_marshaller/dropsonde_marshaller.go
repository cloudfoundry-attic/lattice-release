// Package dropsonde_marshaller provides a tool for marshalling Envelopes
// to Protocol Buffer messages.
//
// Use
//
// Instantiate a Marshaller and run it:
//
//		marshaller := dropsonde_marshaller.NewDropsondeMarshaller(logger)
//		inputChan := make(chan *events.Envelope) // or use a channel provided by some other source
//		outputChan := make(chan []byte)
//		go marshaller.Run(inputChan, outputChan)
//
// The marshaller self-instruments, counting the number of messages
// processed and the number of errors. These can be accessed through the Emit
// function on the marshaller.
package dropsonde_marshaller

import (
	"sync/atomic"
	"unicode"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
)

// A DropsondeMarshaller is an self-instrumenting tool for converting dropsonde
// Envelopes to binary (Protocol Buffer) messages.
type DropsondeMarshaller interface {
	instrumentation.Instrumentable
	Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte)
}

// NewDropsondeMarshaller instantiates a DropsondeMarshaller and logs to the
// provided logger.
func NewDropsondeMarshaller(logger *gosteno.Logger) DropsondeMarshaller {
	messageCounts := make(map[events.Envelope_EventType]*uint64)
	for key := range events.Envelope_EventType_name {
		var count uint64
		messageCounts[events.Envelope_EventType(key)] = &count
	}
	return &dropsondeMarshaller{
		logger:        logger,
		messageCounts: messageCounts,
	}
}

type dropsondeMarshaller struct {
	logger            *gosteno.Logger
	messageCounts     map[events.Envelope_EventType]*uint64
	marshalErrorCount uint64
}

// Run reads Envelopes from inputChan, marshals them to Protocol Buffer format,
// and emits the binary messages onto outputChan. It operates one message at a
// time, and will block if outputChan is not read.
func (u *dropsondeMarshaller) Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte) {
	for message := range inputChan {

		messageBytes, err := proto.Marshal(message)
		if err != nil {
			u.logger.Errorf("dropsondeMarshaller: marshal error %v for message %v", err, message)
			incrementCount(&u.marshalErrorCount)
			continue
		}

		u.logger.Debugf("dropsondeMarshaller: marshalled message %v", spew.Sprintf("%v", message))

		u.incrementMessageCount(message.GetEventType())
		outputChan <- messageBytes
	}
}

func (u *dropsondeMarshaller) incrementMessageCount(eventType events.Envelope_EventType) {
	incrementCount(u.messageCounts[eventType])
}

func incrementCount(count *uint64) {
	atomic.AddUint64(count, 1)
}

func (m *dropsondeMarshaller) metrics() []instrumentation.Metric {
	var metrics []instrumentation.Metric

	for eventType, eventName := range events.Envelope_EventType_name {
		modifiedEventName := []rune(eventName)
		modifiedEventName[0] = unicode.ToLower(modifiedEventName[0])
		metricName := string(modifiedEventName) + "Marshalled"

		metricValue := atomic.LoadUint64(m.messageCounts[events.Envelope_EventType(eventType)])
		metrics = append(metrics, instrumentation.Metric{Name: metricName, Value: metricValue})
	}

	metrics = append(metrics, instrumentation.Metric{
		Name:  "marshalErrors",
		Value: atomic.LoadUint64(&m.marshalErrorCount),
	})

	return metrics
}

// Emit returns the current metrics the DropsondeMarshaller keeps about itself.
func (m *dropsondeMarshaller) Emit() instrumentation.Context {
	return instrumentation.Context{
		Name:    "dropsondeMarshaller",
		Metrics: m.metrics(),
	}
}
