package emitter_test

import (
	"net"
	"sync"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/control"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UdpEmitter", func() {
	var testData = []byte("hello")

	Describe("Close()", func() {
		It("closes the UDP connection", func() {

			udpEmitter, _ := emitter.NewUdpEmitter("localhost:42420")

			udpEmitter.Close()

			err := udpEmitter.Emit(testData)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("use of closed network connection"))
		})
	})

	Describe("Emit()", func() {
		var udpEmitter emitter.ByteEmitter

		Context("when the agent is listening", func() {

			var agentListener net.PacketConn

			BeforeEach(func() {
				var err error
				agentListener, err = net.ListenPacket("udp4", "")
				Expect(err).ToNot(HaveOccurred())

				udpEmitter, err = emitter.NewUdpEmitter(agentListener.LocalAddr().String())
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				agentListener.Close()
			})

			It("should send the data", func(done Done) {
				err := udpEmitter.Emit(testData)
				Expect(err).ToNot(HaveOccurred())

				buffer := make([]byte, 4096)
				readCount, _, err := agentListener.ReadFrom(buffer)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer[:readCount]).To(Equal(testData))

				close(done)
			})
		})

		Context("when the agent is not listening", func() {
			BeforeEach(func() {
				udpEmitter, _ = emitter.NewUdpEmitter("localhost:12345")
			})

			It("should attempt to send the data", func() {
				err := udpEmitter.Emit(testData)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("then the agent starts Listening", func() {
				It("should eventually send data", func(done Done) {
					err := udpEmitter.Emit(testData)
					Expect(err).ToNot(HaveOccurred())

					agentListener, err := net.ListenPacket("udp4", ":12345")
					Expect(err).ToNot(HaveOccurred())

					err = udpEmitter.Emit(testData)
					Expect(err).ToNot(HaveOccurred())

					buffer := make([]byte, 4096)
					readCount, _, err := agentListener.ReadFrom(buffer)
					Expect(err).ToNot(HaveOccurred())
					Expect(buffer[:readCount]).To(Equal(testData))

					close(done)
				})
			})
		})
	})

	Describe("NewUdpEmitter()", func() {
		Context("when ResolveUDPAddr fails", func() {
			It("returns an error", func() {
				emitter, err := emitter.NewUdpEmitter("invalid-address:")
				Expect(emitter).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all is good", func() {
			It("creates an emitter", func() {
				emitter, err := emitter.NewUdpEmitter("localhost:123")
				Expect(emitter).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("ListenForHeartbeatRequest", func() {
		var timesCalled int
		var lastControlMessage *control.ControlMessage
		var lock sync.Mutex

		var fakeResponder = func(controlMessage *control.ControlMessage) {
			lock.Lock()
			defer lock.Unlock()
			timesCalled++
			lastControlMessage = controlMessage
		}

		var getTimesCalled = func() int {
			lock.Lock()
			defer lock.Unlock()
			return timesCalled
		}

		var getReceivedHeartbeatRequest = func() *control.ControlMessage {
			lock.Lock()
			defer lock.Unlock()

			return lastControlMessage
		}

		BeforeEach(func() {
			lock = sync.Mutex{}
			timesCalled = 0
		})

		It("calls responder with the correct heartbeat request when when heartbeat is requested", func() {
			emitter, _ := emitter.NewUdpEmitter("localhost:123")
			go emitter.ListenForHeartbeatRequest(fakeResponder)

			Expect(timesCalled).To(BeZero())

			heartbeatRequest := newHeartbeatRequest()

			sendHeartbeatRequest(emitter.Address(), heartbeatRequest)
			Eventually(getTimesCalled).Should(Equal(1))
			Expect(getReceivedHeartbeatRequest()).To(Equal(heartbeatRequest))
		})

		It("responds to multiple heartbeat requests", func() {
			emitter, _ := emitter.NewUdpEmitter("localhost:123")
			go emitter.ListenForHeartbeatRequest(fakeResponder)
			sendHeartbeatRequest(emitter.Address(), newHeartbeatRequest())
			sendHeartbeatRequest(emitter.Address(), newHeartbeatRequest())

			Eventually(getTimesCalled).Should(Equal(2))
		})

		It("returns an error if listening on the UDP port fails", func() {
			emitter, _ := emitter.NewUdpEmitter("localhost:123")
			emitter.Close()
			err := emitter.ListenForHeartbeatRequest(fakeResponder)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("use of closed network connection")))
		})
	})
})

func sendHeartbeatRequest(addr net.Addr, message *control.ControlMessage) {
	encodedMessage, _ := proto.Marshal(message)
	conn, _ := net.ListenPacket("udp4", "")
	conn.WriteTo(encodedMessage, addr)
}

func newHeartbeatRequest() *control.ControlMessage {
	id, _ := uuid.NewV4()

	return &control.ControlMessage{
		Origin:      proto.String("test"),
		Identifier:  factories.NewControlUUID(id),
		Timestamp:   proto.Int64(0),
		ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
	}
}
