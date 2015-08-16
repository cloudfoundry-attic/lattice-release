package emitter_test

import (
	"errors"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getHeartbeatEvent(ie emitter.InstrumentedEmitter) *events.Heartbeat {
	return ie.GetHeartbeatEvent().(*events.Heartbeat)
}

var _ = Describe("InstrumentedEmitter", func() {
	var fakeEmitter *fake.FakeByteEmitter
	var instrumentedEmitter emitter.InstrumentedEmitter

	BeforeEach(func() {
		fakeEmitter = fake.NewFakeByteEmitter()
		instrumentedEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
	})

	Describe("Delegators", func() {

		It("delegates Close() to the concreteEmitter", func() {
			instrumentedEmitter.Close()
			Eventually(fakeEmitter.IsClosed).Should(BeTrue())
		})
	})

	Describe("Emit()", func() {
		var testData = []byte("hello")

		It("calls the concrete emitter", func() {
			Expect(fakeEmitter.Messages).To(HaveLen(0))

			err := instrumentedEmitter.Emit(testData)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeEmitter.Messages).To(HaveLen(1))
			Expect(fakeEmitter.Messages[0]).To(Equal(testData))
		})
		It("increments the ReceivedMetricsCounter", func() {
			Expect(getHeartbeatEvent(instrumentedEmitter).GetReceivedCount()).To(BeNumerically("==", 0))

			err := instrumentedEmitter.Emit(testData)
			Expect(err).ToNot(HaveOccurred())

			Expect(getHeartbeatEvent(instrumentedEmitter).GetReceivedCount()).To(BeNumerically("==", 1))
		})
		Context("when the concrete ByteEmitter returns no error on Emit()", func() {
			It("increments the SentMetricsCounter", func() {
				Expect(getHeartbeatEvent(instrumentedEmitter).GetSentCount()).To(BeNumerically("==", 0))

				err := instrumentedEmitter.Emit(testData)
				Expect(err).ToNot(HaveOccurred())

				Expect(getHeartbeatEvent(instrumentedEmitter).GetSentCount()).To(BeNumerically("==", 1))
			})
		})
		Context("when the concrete ByteEmitter returns an error on Emit()", func() {
			BeforeEach(func() {
				fakeEmitter.ReturnError = errors.New("fake error")
			})
			It("increments the ErrorCounter", func() {
				Expect(getHeartbeatEvent(instrumentedEmitter).GetErrorCount()).To(BeNumerically("==", 0))
				Expect(getHeartbeatEvent(instrumentedEmitter).GetReceivedCount()).To(BeNumerically("==", 0))
				Expect(getHeartbeatEvent(instrumentedEmitter).GetSentCount()).To(BeNumerically("==", 0))

				err := instrumentedEmitter.Emit(testData)
				Expect(err).To(HaveOccurred())

				Expect(getHeartbeatEvent(instrumentedEmitter).GetErrorCount()).To(BeNumerically("==", 1))
				Expect(getHeartbeatEvent(instrumentedEmitter).GetReceivedCount()).To(BeNumerically("==", 1))
				Expect(getHeartbeatEvent(instrumentedEmitter).GetSentCount()).To(BeNumerically("==", 0))
			})
		})
	})

	Describe("NewInstrumentedEmitter", func() {
		Context("when the concrete ByteEmitter is nil", func() {
			It("returns a nil instrumented emitter", func() {
				emitter, _ := emitter.NewInstrumentedEmitter(nil)
				Expect(emitter).To(BeNil())
			})
			It("returns a helpful error", func() {
				_, err := emitter.NewInstrumentedEmitter(nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
