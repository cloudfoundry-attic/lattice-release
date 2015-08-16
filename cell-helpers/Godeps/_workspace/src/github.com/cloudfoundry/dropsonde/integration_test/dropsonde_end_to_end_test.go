package integration_test

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/sonde-go/control"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// these tests need to be invoked individually from an external script,
// since environment variables need to be set/unset before starting the tests
var _ = Describe("Autowire End-to-End", func() {
	Context("with standard initialization", func() {
		origin := []string{"test-origin"}

		BeforeEach(func() {
			dropsonde.Initialize("localhost:3457", origin...)
			sender := metric_sender.NewMetricSender(dropsonde.AutowiredEmitter())
			batcher := metricbatcher.New(sender, 100*time.Millisecond)
			metrics.Initialize(sender, batcher)
		})

		It("emits HTTP client/server events and heartbeats", func() {
			udpListener, err := net.ListenPacket("udp4", ":3457")
			Expect(err).ToNot(HaveOccurred())
			defer udpListener.Close()
			udpDataChan := make(chan []byte, 16)

			var receivedEvents []eventTracker

			lock := sync.RWMutex{}
			heartbeatRequest := newHeartbeatRequest()
			marshalledHeartbeatRequest, _ := proto.Marshal(heartbeatRequest)

			requestedHeartbeat := false
			go func() {
				defer close(udpDataChan)
				for {
					buffer := make([]byte, 1024)
					n, addr, err := udpListener.ReadFrom(buffer)
					if err != nil {
						return
					}

					if !requestedHeartbeat {

						udpListener.WriteTo(marshalledHeartbeatRequest, addr)
						requestedHeartbeat = true
					}

					if n == 0 {
						panic("Received empty packet")
					}
					envelope := new(events.Envelope)
					err = proto.Unmarshal(buffer[0:n], envelope)
					if err != nil {
						panic(err)
					}

					var eventId = envelope.GetEventType().String()

					tracker := eventTracker{eventType: eventId}

					switch envelope.GetEventType() {
					case events.Envelope_HttpStart:
						tracker.name = envelope.GetHttpStart().GetPeerType().String()
					case events.Envelope_HttpStop:
						tracker.name = envelope.GetHttpStop().GetPeerType().String()
					case events.Envelope_Heartbeat:
						tracker.name = envelope.GetHeartbeat().GetControlMessageIdentifier().String()
					case events.Envelope_ValueMetric:
						tracker.name = envelope.GetValueMetric().GetName()
					case events.Envelope_CounterEvent:
						tracker.name = envelope.GetCounterEvent().GetName()
					default:
						panic("Unexpected message type")

					}

					if envelope.GetOrigin() != strings.Join(origin, "/") {
						panic("origin not as expected")
					}

					func() {
						lock.Lock()
						defer lock.Unlock()
						receivedEvents = append(receivedEvents, tracker)
					}()
				}
			}()

			httpListener, err := net.Listen("tcp", "localhost:0")
			Expect(err).ToNot(HaveOccurred())
			defer httpListener.Close()
			httpHandler := dropsonde.InstrumentedHandler(FakeHandler{})
			go http.Serve(httpListener, httpHandler)

			_, err = http.Get("http://" + httpListener.Addr().String())
			Expect(err).ToNot(HaveOccurred())

			metrics.SendValue("TestMetric", 0, "")
			metrics.IncrementCounter("TestIncrementCounter")
			metrics.BatchIncrementCounter("TestBatchedCounter")

			expectedEvents := []eventTracker{
				{eventType: "HttpStart", name: "Client"},
				{eventType: "HttpStart", name: "Server"},
				{eventType: "HttpStop", name: "Client"},
				{eventType: "HttpStop", name: "Server"},
				{eventType: "ValueMetric", name: "numCPUS"},
				{eventType: "ValueMetric", name: "TestMetric"},
				{eventType: "CounterEvent", name: "TestIncrementCounter"},
				{eventType: "CounterEvent", name: "TestBatchedCounter"},
				{eventType: "Heartbeat", name: heartbeatRequest.GetIdentifier().String()},
			}

			for _, tracker := range expectedEvents {
				Eventually(func() []eventTracker { lock.Lock(); defer lock.Unlock(); return receivedEvents }).Should(ContainElement(tracker))
			}
		})
	})
})

type eventTracker struct {
	eventType string
	name      string
}

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Hello")
}

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func newHeartbeatRequest() *control.ControlMessage {
	id, _ := uuid.NewV4()

	return &control.ControlMessage{
		Origin:      proto.String("MET"),
		Identifier:  factories.NewControlUUID(id),
		Timestamp:   proto.Int64(time.Now().UnixNano()),
		ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
	}
}
