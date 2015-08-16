package dropsonde_unmarshaller_test

import (
	"github.com/cloudfoundry/dropsonde/dropsonde_unmarshaller"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"

	"fmt"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
	"sync"
)

var _ = Describe("DropsondeUnmarshallerCollection", func() {
	var (
		inputChan  chan []byte
		outputChan chan *events.Envelope
		collection dropsonde_unmarshaller.DropsondeUnmarshallerCollection
		waitGroup  *sync.WaitGroup
	)
	BeforeEach(func() {
		inputChan = make(chan []byte, 10)
		outputChan = make(chan *events.Envelope, 10)
		collection = dropsonde_unmarshaller.NewDropsondeUnmarshallerCollection(loggertesthelper.Logger(), 5)
		waitGroup = &sync.WaitGroup{}
	})

	Context("DropsondeUnmarshallerCollection", func() {
		It("creates the right number of unmarshallers", func() {
			Expect(collection.Size()).To(Equal(5))
		})

	})

	Context("Run", func() {
		It("runs its collection of unmarshallers in separate go routines", func() {
			startingCountGoroutines := runtime.NumGoroutine()
			collection.Run(inputChan, outputChan, waitGroup)
			Expect(startingCountGoroutines + 5).To(Equal(runtime.NumGoroutine()))
		})
	})

	Context("metrics", func() {
		It("emits a total log messages concatenated from the different unmarshallers", func() {
			for n := 0; n < 5; n++ {
				envelope := &events.Envelope{
					Origin:     proto.String("fake-origin-3"),
					EventType:  events.Envelope_LogMessage.Enum(),
					LogMessage: factories.NewLogMessage(events.LogMessage_OUT, "test log message "+string(n), "fake-app-id-1", "DEA"),
				}
				message, _ := proto.Marshal(envelope)

				inputChan <- message
			}

			collection.Run(inputChan, outputChan, waitGroup)

			for n := 0; n < 5; n++ {
				<-outputChan
			}

			metrics := collection.Emit().Metrics

			Expect(metrics).NotTo(BeNil())

			metricsNameMap := make(map[string]int)
			for _, m := range metrics {
				metricsNameMap[m.Name]++
			}

			Expect(metricsNameMap["logMessageTotal"]).To(Equal(1))
			for _, metric := range metrics {
				if metric.Name == "logMessageTotal" {
					Expect(metric.Value.(uint64)).To(Equal(uint64(5)))
				}
			}

			for name, count := range metricsNameMap {
				Expect(count).To(Equal(1), fmt.Sprintf("%v has %v metrics, expected only ONE", name, count))
			}
		})

		It("emits log messages metrics per app concatenated from the different unmarshallers", func() {
			collection.Run(inputChan, outputChan, waitGroup)

			for n := 0; n < 25; n++ {
				envelope := &events.Envelope{
					Origin:     proto.String("fake-origin-3"),
					EventType:  events.Envelope_LogMessage.Enum(),
					LogMessage: factories.NewLogMessage(events.LogMessage_OUT, "test log message "+string(n), "fake-app-id-"+string(n%5), "DEA"),
				}
				message, _ := proto.Marshal(envelope)

				inputChan <- message
			}

			for n := 0; n < 25; n++ {
				<-outputChan
			}

			metrics := collection.Emit().Metrics

			Expect(metrics).NotTo(BeNil())

			metricsNameMap := make(map[string]int)
			for _, m := range metrics {
				metricsNameMap[m.Name]++
			}
		})

		It("emits event type metrics concatenated from the different unmarshallers", func() {
			collection.Run(inputChan, outputChan, waitGroup)

			for n := 0; n < 7; n++ {
				envelope := &events.Envelope{
					Origin:    proto.String("fake-origin-1"),
					EventType: events.Envelope_Heartbeat.Enum(),
					Heartbeat: factories.NewHeartbeat(1, 2, 3),
				}
				message, _ := proto.Marshal(envelope)

				inputChan <- message
			}

			for n := 0; n < 7; n++ {
				<-outputChan
			}

			metrics := collection.Emit().Metrics

			Expect(metrics).NotTo(BeNil())

			metricsNameMap := make(map[string]int)
			for _, m := range metrics {
				metricsNameMap[m.Name]++
			}

			Expect(metricsNameMap["heartbeatReceived"]).To(Equal(1))
			for _, metric := range metrics {
				if metric.Name == "heartbeatReceived" {
					Expect(metric.Value.(uint64)).To(Equal(uint64(7)))
				}
			}
		})

		It("emits the correct metrics context", func() {
			Expect(collection.Emit().Name).To(Equal("dropsondeUnmarshaller"))
		})
	})
})
