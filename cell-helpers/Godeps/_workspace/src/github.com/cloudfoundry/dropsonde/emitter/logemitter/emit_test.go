package logemitter_test

import (
	"log"
	"os"
	"strings"

	. "github.com/cloudfoundry/dropsonde/emitter/logemitter"
	"github.com/cloudfoundry/dropsonde/emitter/logemitter/testhelpers"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gogo/protobuf/proto"

	"io/ioutil"

	"github.com/cloudfoundry/loggregatorlib/loggregatorclient/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	log.SetOutput(ioutil.Discard)
})

var _ = Describe("Testing with Ginkgo", func() {
	var (
		received chan *[]byte
		emitter  *LoggregatorEmitter
	)

	BeforeEach(func() {
		var err error
		received = make(chan *[]byte, 10)
		os.Setenv("LOGGREGATOR_SHARED_SECRET", "secret")
		emitter, err = NewEmitter("localhost:3456", "ROUTER", "42", false)
		Expect(err).ToNot(HaveOccurred())

		emitter.LoggregatorClient = &fake.FakeLoggregatorClient{Received: received}

	})

	It("returns an error if LOGGREGATOR_SHARED_SECRET is not set", func() {
		os.Setenv("LOGGREGATOR_SHARED_SECRET", "")
		_, err := NewEmitter("localhost:3456", "ROUTER", "42", false)
		Expect(err).To(Equal(ERR_SHARED_SECRET_NOT_SET))
	})

	It("should emit stdout", func() {
		emitter.Emit("appid", "foo")
		receivedMessage := extractLogMessage(<-received)

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedMessage.GetAppId()).To(Equal("appid"))
		Expect(receivedMessage.GetSourceId()).To(Equal("42"))
		Expect(receivedMessage.GetMessageType()).To(Equal(logmessage.LogMessage_OUT))
	})

	It("should emit stderr", func() {
		emitter.EmitError("appid", "foo")
		receivedMessage := extractLogMessage(<-received)

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedMessage.GetAppId()).To(Equal("appid"))
		Expect(receivedMessage.GetSourceId()).To(Equal("42"))
		Expect(receivedMessage.GetMessageType()).To(Equal(logmessage.LogMessage_ERR))
	})

	It("should emit fully formed log messages", func() {
		logMessage := testhelpers.NewLogMessage("test_msg", "test_app_id")
		logMessage.SourceInstance = proto.String("src_id")

		emitter.EmitLogMessage(logMessage)
		receivedMessage := extractLogMessage(<-received)

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("test_msg")))
		Expect(receivedMessage.GetAppId()).To(Equal("test_app_id"))
		Expect(receivedMessage.GetSourceId()).To(Equal("src_id"))
	})

	It("should truncate long messages", func() {
		longMessage := strings.Repeat("7", MAX_MESSAGE_BYTE_SIZE*2)
		logMessage := testhelpers.NewLogMessage(longMessage, "test_app_id")

		emitter.EmitLogMessage(logMessage)

		receivedMessage := extractLogMessage(<-received)
		receivedMessageText := receivedMessage.GetMessage()

		truncatedOffset := len(receivedMessageText) - len(TRUNCATED_BYTES)
		expectedBytes := append([]byte(receivedMessageText)[:truncatedOffset], TRUNCATED_BYTES...)

		Expect(receivedMessageText).To(Equal(expectedBytes))
		Expect(receivedMessageText).To(HaveLen(MAX_MESSAGE_BYTE_SIZE))
	})

	It("should split messages on new lines", func() {
		message := "message1\n\rmessage2\nmessage3\r\nmessage4\r"
		logMessage := testhelpers.NewLogMessage(message, "test_app_id")

		emitter.EmitLogMessage(logMessage)
		Expect(received).To(HaveLen(4))

		for _, expectedMessage := range []string{"message1", "message2", "message3", "message4"} {
			receivedMessage := extractLogMessage(<-received)
			Expect(receivedMessage.GetMessage()).To(Equal([]byte(expectedMessage)))
		}
	})

	It("should build the log envelope correctly", func() {
		emitter.Emit("appid", "foo")
		receivedEnvelope := extractLogEnvelope(<-received)

		Expect(receivedEnvelope.GetLogMessage().GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedEnvelope.GetLogMessage().GetAppId()).To(Equal("appid"))
		Expect(receivedEnvelope.GetRoutingKey()).To(Equal("appid"))
		Expect(receivedEnvelope.GetLogMessage().GetSourceId()).To(Equal("42"))
	})

	It("should sign the log message correctly", func() {
		emitter.Emit("appid", "foo")
		receivedEnvelope := extractLogEnvelope(<-received)
		Expect(receivedEnvelope.VerifySignature("secret")).To(BeTrue(), "Expected envelope to be signed with the correct secret key")
	})

	It("source name is set if mapping is unknown", func() {
		os.Setenv("LOGGREGATOR_SHARED_SECRET", "secret")
		emitter, err := NewEmitter("localhost:3456", "XYZ", "42", false)
		Expect(err).ToNot(HaveOccurred())
		emitter.LoggregatorClient = &fake.FakeLoggregatorClient{Received: received}

		emitter.Emit("test_app_id", "test_msg")
		receivedMessage := extractLogMessage(<-received)

		Expect(receivedMessage.GetSourceName()).To(Equal("XYZ"))
	})

	Context("when missing an app id", func() {
		It("should not emit", func() {
			emitter.Emit("", "foo")
			Expect(received).ToNot(Receive(), "Message without app id should not have been emitted")

			emitter.Emit("    ", "foo")
			Expect(received).ToNot(Receive(), "Message with an empty app id should not have been emitted")
		})
	})
})

func extractLogEnvelope(data *[]byte) *logmessage.LogEnvelope {
	receivedEnvelope := &logmessage.LogEnvelope{}

	err := proto.Unmarshal(*data, receivedEnvelope)
	Expect(err).ToNot(HaveOccurred())

	return receivedEnvelope
}

func extractLogMessage(data *[]byte) *logmessage.LogMessage {
	envelope := extractLogEnvelope(data)

	return envelope.GetLogMessage()
}
