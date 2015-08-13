package emitter_test

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EventEmitter", func() {
	Describe("Emit", func() {
		Context("without an origin", func() {
			It("returns an error", func() {
				innerEmitter := fake.NewFakeByteEmitter()
				eventEmitter := emitter.NewEventEmitter(innerEmitter, "")

				testEvent := factories.NewHeartbeat(1, 2, 3)
				err := eventEmitter.Emit(testEvent)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Wrap: "))
			})
		})

		It("marshals events and delegates to the inner emitter", func() {
			innerEmitter := fake.NewFakeByteEmitter()
			origin := "fake-origin"
			eventEmitter := emitter.NewEventEmitter(innerEmitter, origin)

			testEvent := factories.NewHeartbeat(1, 2, 3)
			err := eventEmitter.Emit(testEvent)
			Expect(err).ToNot(HaveOccurred())

			Expect(innerEmitter.GetMessages()).To(HaveLen(1))
			msg := innerEmitter.GetMessages()[0]

			var envelope events.Envelope
			err = proto.Unmarshal(msg, &envelope)
			Expect(err).ToNot(HaveOccurred())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_Heartbeat))
		})
	})

	Describe("Close", func() {
		It("closes the inner emitter", func() {
			innerEmitter := fake.NewFakeByteEmitter()
			eventEmitter := emitter.NewEventEmitter(innerEmitter, "")

			eventEmitter.Close()
			Expect(innerEmitter.IsClosed()).To(BeTrue())
		})
	})
})
