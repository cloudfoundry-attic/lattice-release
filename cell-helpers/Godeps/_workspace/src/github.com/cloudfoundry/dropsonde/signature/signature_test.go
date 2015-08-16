package signature_test

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/cloudfoundry/dropsonde/signature"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verifier", func() {
	var (
		inputChan   chan []byte
		outputChan  chan []byte
		runComplete chan struct{}

		signatureVerifier *signature.Verifier
	)

	BeforeEach(func() {
		inputChan = make(chan []byte, 10)
		outputChan = make(chan []byte, 10)
		runComplete = make(chan struct{})

		signatureVerifier = signature.NewVerifier(loggertesthelper.Logger(), "valid-secret")

		go func() {
			signatureVerifier.Run(inputChan, outputChan)
			close(runComplete)
		}()
	})

	AfterEach(func() {
		close(inputChan)
		Eventually(runComplete).Should(BeClosed())
	})

	It("discards messages less than 32 bytes long", func() {
		loggertesthelper.TestLoggerSink.Clear()

		message := make([]byte, 1)
		inputChan <- message
		Consistently(outputChan).ShouldNot(Receive())

		Expect(loggertesthelper.TestLoggerSink.LogContents()).To(ContainSubstring("missing signature for message"))
	})

	It("discards messages when verification fails", func() {
		loggertesthelper.TestLoggerSink.Clear()

		message := make([]byte, 33)

		inputChan <- message
		Consistently(outputChan).ShouldNot(Receive())

		Expect(loggertesthelper.TestLoggerSink.LogContents()).To(ContainSubstring("invalid signature for message"))
	})

	It("passes through messages with valid signature", func() {
		loggertesthelper.TestLoggerSink.Clear()

		message := []byte{1, 2, 3}
		mac := hmac.New(sha256.New, []byte("valid-secret"))
		mac.Write(message)
		signature := mac.Sum(nil)

		signedMessage := append(signature, message...)

		inputChan <- signedMessage
		outputMessage := <-outputChan
		Expect(outputMessage).To(Equal(message))

		Expect(loggertesthelper.TestLoggerSink.LogContents()).To(BeEmpty())
	})

	Context("metrics", func() {

		BeforeEach(func() {
			fakeEventEmitter.Reset()
			metricBatcher.Reset()
		})

		It("emits a missing signature error counter", func() {
			inputChan <- []byte{1, 2, 3}
			Eventually(fakeEventEmitter.GetMessages).Should(HaveLen(1))
			Expect(fakeEventEmitter.GetMessages()[0].Event.(*events.CounterEvent)).To(Equal(&events.CounterEvent{
				Name:  proto.String("signatureVerifier.missingSignatureErrors"),
				Delta: proto.Uint64(1),
			}))
		})

		It("emits an invalid signature error counter", func() {
			inputChan <- make([]byte, 32)

			Eventually(fakeEventEmitter.GetMessages).Should(HaveLen(1))
			Expect(fakeEventEmitter.GetMessages()[0].Event.(*events.CounterEvent)).To(Equal(&events.CounterEvent{
				Name:  proto.String("signatureVerifier.invalidSignatureErrors"),
				Delta: proto.Uint64(1),
			}))
		})

		It("emits an valid signature counter", func() {
			message := []byte{1, 2, 3}
			mac := hmac.New(sha256.New, []byte("valid-secret"))
			mac.Write(message)
			signature := mac.Sum(nil)

			signedMessage := append(signature, message...)
			inputChan <- signedMessage

			Eventually(fakeEventEmitter.GetMessages).Should(HaveLen(1))
			Expect(fakeEventEmitter.GetMessages()[0].Event.(*events.CounterEvent)).To(Equal(&events.CounterEvent{
				Name:  proto.String("signatureVerifier.validSignatures"),
				Delta: proto.Uint64(1),
			}))
		})
	})
})
