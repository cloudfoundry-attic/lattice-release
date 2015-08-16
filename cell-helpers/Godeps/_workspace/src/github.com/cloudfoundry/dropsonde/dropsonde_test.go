package dropsonde_test

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/control"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Autowire", func() {

	Describe("Initialize", func() {
		It("resets the HTTP default transport to be instrumented", func() {
			dropsonde.InitializeWithEmitter(&dropsonde.NullEventEmitter{})
			Expect(reflect.TypeOf(http.DefaultTransport).Elem().Name()).To(Equal("instrumentedCancelableRoundTripper"))
		})
	})

	Describe("CreateDefaultEmitter", func() {
		Context("with origin set", func() {
			It("responds to heartbeat requests with heartbeats", func() {
				err := dropsonde.Initialize("localhost:1235", "cf", "metron")
				Expect(err).ToNot(HaveOccurred())

				messages := make(chan []byte, 100)
				readyChan := make(chan struct{})

				go respondWithHeartbeatRequest(1235, messages, readyChan)
				<-readyChan

				emitter := dropsonde.AutowiredEmitter()

				err = emitter.Emit(&events.CounterEvent{Name: proto.String("name"), Delta: proto.Uint64(1)})
				Expect(err).NotTo(HaveOccurred())

				bytes := make([]byte, 65000)
				Eventually(messages, 5).Should(Receive(&bytes))
				message := &events.Envelope{}
				proto.Unmarshal(bytes, message)
				Expect(message.GetOrigin()).To(Equal("cf/metron"))
			})
		})

		Context("with origin missing", func() {
			It("returns a NullEventEmitter", func() {
				err := dropsonde.Initialize("localhost:2343", "")
				Expect(err).To(HaveOccurred())

				emitter := dropsonde.AutowiredEmitter()
				Expect(emitter).ToNot(BeNil())
				nullEmitter := &dropsonde.NullEventEmitter{}
				Expect(emitter).To(BeAssignableToTypeOf(nullEmitter))
			})
		})
	})
})

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {}

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func respondWithHeartbeatRequest(port int, messages chan []byte, readyChan chan struct{}) {
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	close(readyChan)
	n, addr, _ := conn.ReadFrom(buf)

	conn.WriteTo(newMarshalledHeartbeatRequest(), addr)
	n, addr, _ = conn.ReadFrom(buf)

	messages <- buf[:n]
	conn.Close()
}

func newMarshalledHeartbeatRequest() []byte {
	id, _ := uuid.NewV4()

	heartbeatRequest := &control.ControlMessage{
		Origin:      proto.String("test"),
		Identifier:  factories.NewControlUUID(id),
		Timestamp:   proto.Int64(time.Now().UnixNano()),
		ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
	}

	bytes, err := proto.Marshal(heartbeatRequest)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}
