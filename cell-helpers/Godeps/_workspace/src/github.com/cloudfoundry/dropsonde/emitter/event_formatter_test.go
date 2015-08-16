package emitter_test

import (
	"github.com/cloudfoundry/dropsonde/emitter"

	"time"

	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type unknownEvent struct{}

func (*unknownEvent) ProtoMessage() {}

var _ = Describe("EventFormatter", func() {
	Describe("wrap", func() {
		var origin string

		BeforeEach(func() {
			origin = "testEventFormatter/42"
		})

		It("works with dropsonde status (Heartbeat) events", func() {
			statusEvent := &events.Heartbeat{SentCount: proto.Uint64(1), ErrorCount: proto.Uint64(0)}
			envelope, _ := emitter.Wrap(statusEvent, origin)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_Heartbeat))
			Expect(envelope.GetHeartbeat()).To(Equal(statusEvent))
		})

		It("works with HttpStart events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStart{RequestId: factories.NewUUID(id)}

			envelope, _ := emitter.Wrap(testEvent, origin)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStart))
			Expect(envelope.GetHttpStart()).To(Equal(testEvent))
		})

		It("works with HttpStop events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStop{RequestId: factories.NewUUID(id)}

			envelope, _ := emitter.Wrap(testEvent, origin)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStop))
			Expect(envelope.GetHttpStop()).To(Equal(testEvent))
		})

		It("works with ValueMetric events", func() {
			testEvent := &events.ValueMetric{Name: proto.String("test-name")}

			envelope, _ := emitter.Wrap(testEvent, origin)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_ValueMetric))
			Expect(envelope.GetValueMetric()).To(Equal(testEvent))
		})

		It("works with CounterEvent events", func() {
			testEvent := &events.CounterEvent{Name: proto.String("test-counter")}

			envelope, _ := emitter.Wrap(testEvent, origin)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_CounterEvent))
			Expect(envelope.GetCounterEvent()).To(Equal(testEvent))
		})

		It("errors with unknown events", func() {
			envelope, err := emitter.Wrap(new(unknownEvent), origin)
			Expect(envelope).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("checks that origin is non-empty", func() {
			id, _ := uuid.NewV4()
			malformedOrigin := ""
			testEvent := &events.HttpStart{RequestId: factories.NewUUID(id)}
			envelope, err := emitter.Wrap(testEvent, malformedOrigin)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Event not emitted due to missing origin information"))
			Expect(envelope).To(BeNil())
		})

		Context("with a known event type", func() {
			var testEvent events.Event

			BeforeEach(func() {
				id, _ := uuid.NewV4()
				testEvent = &events.HttpStop{RequestId: factories.NewUUID(id)}
			})

			It("contains the origin", func() {
				envelope, _ := emitter.Wrap(testEvent, origin)
				Expect(envelope.GetOrigin()).To(Equal("testEventFormatter/42"))
			})

			Context("when the origin is empty", func() {
				It("errors with a helpful message", func() {
					envelope, err := emitter.Wrap(testEvent, "")
					Expect(envelope).To(BeNil())
					Expect(err.Error()).To(Equal("Event not emitted due to missing origin information"))
				})
			})

			It("sets the timestamp to now", func() {
				envelope, _ := emitter.Wrap(testEvent, origin)
				Expect(time.Unix(0, envelope.GetTimestamp())).To(BeTemporally("~", time.Now(), 100*time.Millisecond))
			})
		})
	})
})
