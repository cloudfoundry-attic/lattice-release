package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/events/eventfakes"
	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event Stream Handlers", func() {
	var (
		logger  lager.Logger
		fakeBBS *fake_bbs.FakeClient

		handler *handlers.EventStreamHandler

		server *httptest.Server
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		handler = handlers.NewEventStreamHandler(fakeBBS, logger)
	})

	AfterEach(func(done Done) {
		if server != nil {
			server.Close()
		}
		close(done)
	})

	Describe("EventStream", func() {
		var (
			request         *http.Request
			responseChan    chan *http.Response
			eventStreamDone chan struct{}
		)

		BeforeEach(func() {
			responseChan = make(chan *http.Response)
			eventStreamDone = make(chan struct{})
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.EventStream(w, r)
				close(eventStreamDone)
			}))
		})

		JustBeforeEach(func() {
			var err error
			request, err = http.NewRequest("GET", server.URL, nil)
			Expect(err).NotTo(HaveOccurred())
			go func() {
				defer GinkgoRecover()
				response, _ := http.DefaultClient.Do(request)
				responseChan <- response
			}()
		})

		Context("when failing to subscribe to the event stream", func() {
			BeforeEach(func() {
				fakeBBS.SubscribeToEventsReturns(nil, models.ErrUnknownError)
			})

			It("returns an internal server error", func() {
				response := &http.Response{}
				Eventually(responseChan).Should(Receive(&response))
				Expect(response.StatusCode).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when connection is closed before subscription is complete", func() {
			var subscribe chan bool

			BeforeEach(func() {
				subscribe = make(chan bool)
				fakeBBS.SubscribeToEventsStub = func() (events.EventSource, error) {
					<-subscribe
					return nil, errors.New("broken")
				}
			})

			It("returns an internal server error", func() {
				Eventually(fakeBBS.SubscribeToEventsCallCount).Should(Equal(1))
				http.DefaultTransport.(*http.Transport).CancelRequest(request)

				// should not timeout
				server.Close()
				Eventually(responseChan).Should(Receive())
			})
		})

		Context("when successfully subscribing to the event stream", func() {
			var eventSource *eventfakes.FakeEventSource
			var eventChannel chan models.Event

			BeforeEach(func() {
				eventChannel = make(chan models.Event, 2)
				eventSource = new(eventfakes.FakeEventSource)
				eventSource.NextStub = func() (models.Event, error) {
					select {
					case event := <-eventChannel:
						return event, nil
					case <-time.After(time.Second):
					}
					return nil, errors.New("timeout waiting for events")
				}
				fakeBBS.SubscribeToEventsReturns(eventSource, nil)
			})

			It("emits events from the stream to the connection", func(done Done) {
				response := &http.Response{}
				Eventually(responseChan).Should(Receive(&response))
				reader := sse.NewReadCloser(response.Body)

				desiredLRP := &models.DesiredLRP{
					ProcessGuid: "some-guid",
					Domain:      "some-domain",
					RootFs:      "some-rootfs",
					Action: models.WrapAction(&models.RunAction{
						Path: "true",
						User: "user",
					}),
				}
				eventChannel <- models.NewDesiredLRPCreatedEvent(desiredLRP)

				data, err := json.Marshal(receptor.NewDesiredLRPCreatedEvent(serialization.DesiredLRPProtoToResponse(desiredLRP)))
				Expect(err).NotTo(HaveOccurred())

				event, err := reader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(event.ID).To(Equal("0"))
				Expect(event.Name).To(Equal(string(receptor.EventTypeDesiredLRPCreated)))
				Expect(event.Data).To(MatchJSON(data))

				actualLRP := models.NewUnclaimedActualLRP(models.NewActualLRPKey("some-guid", 3, "some-domain"), 0)
				actualLRPGroup := &models.ActualLRPGroup{Instance: actualLRP}
				eventChannel <- models.NewActualLRPCreatedEvent(actualLRPGroup)

				data, err = json.Marshal(receptor.NewActualLRPCreatedEvent(serialization.ActualLRPProtoToResponse(actualLRP, false)))
				Expect(err).NotTo(HaveOccurred())

				event, err = reader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(event.ID).To(Equal("1"))
				Expect(event.Name).To(Equal(string(receptor.EventTypeActualLRPCreated)))
				Expect(event.Data).To(MatchJSON(data))
				eventChannel <- eventfakes.FakeEvent{"B"}

				close(done)
			})

			It("returns Content-Type as text/event-stream", func() {
				response := &http.Response{}
				Eventually(responseChan).Should(Receive(&response))
				Expect(response.Header.Get("Content-Type")).To(Equal("text/event-stream; charset=utf-8"))
				Expect(response.Header.Get("Cache-Control")).To(Equal("no-cache, no-store, must-revalidate"))
				Expect(response.Header.Get("Connection")).To(Equal("keep-alive"))
			})

			Context("when the source provides an unmarshalable event", func() {
				It("closes the event stream to the client", func(done Done) {
					response := &http.Response{}
					Eventually(responseChan).Should(Receive(&response))
					eventChannel <- eventfakes.UnmarshalableEvent{Fn: func() {}}

					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Expect(err).To(Equal(io.EOF))
					close(done)
				})
			})

			Context("when the event source returns an error", func() {
				BeforeEach(func() {
					eventSource.NextReturns(nil, models.ErrUnknownError)
				})

				It("closes the client event stream", func() {
					response := &http.Response{}
					Eventually(responseChan).Should(Receive(&response))
					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Expect(err).To(Equal(io.EOF))
				})
			})

			Context("when the client closes the response body", func() {
				It("returns early", func() {
					response := &http.Response{}
					Eventually(responseChan).Should(Receive(&response))
					reader := sse.NewReadCloser(response.Body)
					err := reader.Close()
					Expect(err).NotTo(HaveOccurred())
					Eventually(eventStreamDone, 10).Should(BeClosed())
				})
			})
		})
	})
})
