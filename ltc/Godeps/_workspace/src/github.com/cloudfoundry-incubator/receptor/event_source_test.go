package receptor_test

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/go-sse/sse"
)

var _ = Describe("EventSource", func() {
	var eventSource receptor.EventSource
	var fakeRawEventSource *fake_receptor.FakeRawEventSource

	BeforeEach(func() {
		fakeRawEventSource = new(fake_receptor.FakeRawEventSource)
		eventSource = receptor.NewEventSource(fakeRawEventSource)
	})

	Describe("Next", func() {
		Describe("Desired LRP events", func() {
			var desiredLRPResponse receptor.DesiredLRPResponse

			BeforeEach(func() {
				desiredLRPResponse = serialization.DesiredLRPToResponse(
					models.DesiredLRP{
						ProcessGuid: "some-guid",
						Domain:      "some-domain",
						RootFS:      "some-rootfs",
						Action: &models.RunAction{
							Path: "true",
						},
					},
				)
			})

			Context("when receiving a DesiredLRPCreatedEvent", func() {
				var expectedEvent receptor.DesiredLRPCreatedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewDesiredLRPCreatedEvent(desiredLRPResponse)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "hi",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					desiredLRPCreateEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPCreateEvent).Should(Equal(expectedEvent))
				})
			})

			Context("when receiving a DesiredLRPChangedEvent", func() {
				var expectedEvent receptor.DesiredLRPChangedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewDesiredLRPChangedEvent(
						desiredLRPResponse,
						desiredLRPResponse,
					)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "hi",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					desiredLRPChangeEvent, ok := event.(receptor.DesiredLRPChangedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPChangeEvent).Should(Equal(expectedEvent))
				})
			})

			Context("when receiving a DesiredLRPRemovedEvent", func() {
				var expectedEvent receptor.DesiredLRPRemovedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewDesiredLRPRemovedEvent(desiredLRPResponse)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "sup",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
					Ω(ok).Should(BeTrue())
					Ω(desiredLRPRemovedEvent).Should(Equal(expectedEvent))
				})
			})
		})

		Describe("Actual LRP Events", func() {
			var actualLRPResponse receptor.ActualLRPResponse

			BeforeEach(func() {
				actualLRPResponse = serialization.ActualLRPToResponse(
					models.ActualLRP{
						ActualLRPKey: models.NewActualLRPKey("some-guid", 0, "some-domain"),
						State:        models.ActualLRPStateUnclaimed,
						Since:        1,
					},
					false,
				)
			})

			Context("when receiving a ActualLRPCreatedEvent", func() {
				var expectedEvent receptor.ActualLRPCreatedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewActualLRPCreatedEvent(actualLRPResponse)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "sup",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					actualLRPCreatedEvent, ok := event.(receptor.ActualLRPCreatedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPCreatedEvent).Should(Equal(expectedEvent))
				})
			})

			Context("when receiving a ActualLRPChangedEvent", func() {
				var expectedEvent receptor.ActualLRPChangedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewActualLRPChangedEvent(
						actualLRPResponse,
						actualLRPResponse,
					)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "sup",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					actualLRPChangedEvent, ok := event.(receptor.ActualLRPChangedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPChangedEvent).Should(Equal(expectedEvent))
				})
			})

			Context("when receiving a ActualLRPRemovedEvent", func() {
				var expectedEvent receptor.ActualLRPRemovedEvent

				BeforeEach(func() {
					expectedEvent = receptor.NewActualLRPRemovedEvent(actualLRPResponse)
					payload, err := json.Marshal(expectedEvent)
					Ω(err).ShouldNot(HaveOccurred())

					fakeRawEventSource.NextReturns(
						sse.Event{
							ID:   "sup",
							Name: string(expectedEvent.EventType()),
							Data: payload,
						},
						nil,
					)
				})

				It("returns the event", func() {
					event, err := eventSource.Next()
					Ω(err).ShouldNot(HaveOccurred())

					actualLRPRemovedEvent, ok := event.(receptor.ActualLRPRemovedEvent)
					Ω(ok).Should(BeTrue())
					Ω(actualLRPRemovedEvent).Should(Equal(expectedEvent))
				})
			})
		})

		Context("when receiving an unrecognized event", func() {
			BeforeEach(func() {
				fakeRawEventSource.NextReturns(
					sse.Event{
						ID:   "sup",
						Name: "unrecognized-event-type",
						Data: []byte("{\"key\":\"value\"}"),
					},
					nil,
				)
			})

			It("returns an unrecognized event error", func() {
				_, err := eventSource.Next()
				Ω(err).Should(Equal(receptor.ErrUnrecognizedEventType))
			})
		})

		Context("when receiving a bad payload", func() {
			BeforeEach(func() {
				fakeRawEventSource.NextReturns(
					sse.Event{
						ID:   "sup",
						Name: string(receptor.EventTypeDesiredLRPCreated),
						Data: []byte("{\"desired_lrp\":\"not a desired lrp\"}"),
					},
					nil,
				)
			})

			It("returns a json error", func() {
				_, err := eventSource.Next()
				Ω(err).Should(BeAssignableToTypeOf(receptor.NewInvalidPayloadError(errors.New("whatever"))))
			})
		})

		Context("when the raw event source returns an error", func() {
			var rawError error

			BeforeEach(func() {
				rawError = errors.New("raw-error")
				fakeRawEventSource.NextReturns(sse.Event{}, rawError)
			})

			It("propagates the error", func() {
				_, err := eventSource.Next()
				Ω(err).Should(Equal(receptor.NewRawEventSourceError(rawError)))
			})
		})

		Context("when the raw event source returns io.EOF", func() {
			BeforeEach(func() {
				fakeRawEventSource.NextReturns(sse.Event{}, io.EOF)
			})

			It("returns io.EOF", func() {
				_, err := eventSource.Next()
				Ω(err).Should(Equal(io.EOF))
			})
		})

		Context("when the raw event source returns sse.ErrSourceClosed", func() {
			BeforeEach(func() {
				fakeRawEventSource.NextReturns(sse.Event{}, sse.ErrSourceClosed)
			})

			It("returns receptor.ErrSourceClosed", func() {
				_, err := eventSource.Next()
				Ω(err).Should(Equal(receptor.ErrSourceClosed))
			})
		})
	})

	Describe("Close", func() {
		Context("when the raw source closes normally", func() {
			It("closes the raw event source", func() {
				eventSource.Close()
				Ω(fakeRawEventSource.CloseCallCount()).Should(Equal(1))
			})

			It("does not error", func() {
				err := eventSource.Close()
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the raw source closes with error", func() {
			var rawError error

			BeforeEach(func() {
				rawError = errors.New("ka-boom")
				fakeRawEventSource.CloseReturns(rawError)
			})

			It("closes the raw event source", func() {
				eventSource.Close()
				Ω(fakeRawEventSource.CloseCallCount()).Should(Equal(1))
			})

			It("propagates the error", func() {
				err := eventSource.Close()
				Ω(err).Should(Equal(receptor.NewCloseError(rawError)))
			})
		})
	})
})
