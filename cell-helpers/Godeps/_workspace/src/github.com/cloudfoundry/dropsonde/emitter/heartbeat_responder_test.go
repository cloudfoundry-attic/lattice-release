package emitter_test

import (
	"bytes"
	"errors"
	"log"
	"time"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/control"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HeartbeatResponder", func() {
	var (
		wrappedEmitter *fake.FakeByteEmitter
		origin         = "testHeartbeatResponder/0"
	)

	BeforeEach(func() {
		wrappedEmitter = fake.NewFakeByteEmitter()
	})

	Describe("NewHeartbeatResponder", func() {
		It("requires non-nil args", func() {
			heartbeatResponder, err := emitter.NewHeartbeatResponder(nil, origin)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("wrappedEmitter is nil"))
			Expect(heartbeatResponder).To(BeNil())
		})
	})

	Describe("Emit", func() {
		var (
			heartbeatResponder emitter.RespondingByteEmitter
			testData           = []byte("hello")
		)

		BeforeEach(func() {
			heartbeatResponder, _ = emitter.NewHeartbeatResponder(wrappedEmitter, origin)
		})

		It("delegates to the wrapped emitter", func() {
			heartbeatResponder.Emit(testData)

			messages := wrappedEmitter.GetMessages()
			Expect(messages).To(HaveLen(1))
			Expect(messages[0]).To(Equal(testData))
		})

		It("increments the heartbeat counter", func() {
			id, _ := uuid.NewV4()

			heartbeatRequest := &control.ControlMessage{
				Origin:      proto.String("test"),
				Identifier:  factories.NewControlUUID(id),
				Timestamp:   proto.Int64(time.Now().UnixNano()),
				ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
			}
			heartbeatResponder.Emit(testData)
			heartbeatResponder.Respond(heartbeatRequest)

			Eventually(wrappedEmitter.GetMessages).Should(HaveLen(2))

			message := wrappedEmitter.GetMessages()[1]
			hbEnvelope := &events.Envelope{}
			err := proto.Unmarshal(message, hbEnvelope)
			Expect(err).NotTo(HaveOccurred())

			hbEvent := hbEnvelope.GetHeartbeat()

			Expect(hbEvent.GetReceivedCount()).To(Equal(uint64(1)))
		})
	})

	Describe("Close", func() {
		var heartbeatResponder emitter.ByteEmitter

		BeforeEach(func() {
			heartbeatResponder, _ = emitter.NewHeartbeatResponder(wrappedEmitter, origin)
		})

		It("eventually delegates to the inner heartbeat emitter", func() {
			heartbeatResponder.Close()
			Eventually(wrappedEmitter.IsClosed).Should(BeTrue())
		})

		It("can be called more than once", func() {
			heartbeatResponder.Close()
			Expect(heartbeatResponder.Close).ToNot(Panic())
		})
	})

	Describe("RespondToHeartbeat", func() {
		var heartbeatResponder emitter.RespondingByteEmitter

		BeforeEach(func() {
			heartbeatResponder, _ = emitter.NewHeartbeatResponder(wrappedEmitter, origin)
		})

		It("creates a Heartbeat message", func() {
			id, _ := uuid.NewV4()

			heartbeatRequest := &control.ControlMessage{
				Origin:      proto.String("tst"),
				Identifier:  factories.NewControlUUID(id),
				Timestamp:   proto.Int64(time.Now().UnixNano()),
				ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
			}

			heartbeatResponder.Respond(heartbeatRequest)
			Expect(wrappedEmitter.GetMessages()).To(HaveLen(1))
			hbBytes := wrappedEmitter.GetMessages()[0]

			var heartbeat events.Envelope
			err := proto.Unmarshal(hbBytes, &heartbeat)
			Expect(err).NotTo(HaveOccurred())

			heartbeatUuid := heartbeatRequest.GetIdentifier().String()
			Expect(heartbeat.GetHeartbeat().ControlMessageIdentifier.String()).To(Equal(heartbeatUuid))
		})

		It("logs an error when heartbeat emission fails", func() {
			wrappedEmitter.ReturnError = errors.New("fake error")

			logWriter := new(bytes.Buffer)
			log.SetOutput(logWriter)

			id, _ := uuid.NewV4()

			heartbeatRequest := &control.ControlMessage{
				Origin:      proto.String("tst"),
				Identifier:  factories.NewControlUUID(id),
				Timestamp:   proto.Int64(time.Now().UnixNano()),
				ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
			}

			heartbeatResponder.Respond(heartbeatRequest)

			loggedText := string(logWriter.Bytes())
			expectedText := "Problem while emitting heartbeat data: fake error"
			Expect(loggedText).To(ContainSubstring(expectedText))
			heartbeatResponder.Close()
		})
	})
})
