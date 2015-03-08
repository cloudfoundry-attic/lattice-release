package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeEvent struct {
	Token string `json:"token"`
}

func (fakeEvent) EventType() receptor.EventType {
	return "fake"
}

func (fakeEvent) Key() string {
	return "fake"
}

type unmarshalableEvent struct {
	Fn func() `json:"fn"`
}

func (unmarshalableEvent) EventType() receptor.EventType {
	return "unmarshalable"
}

func (unmarshalableEvent) Key() string {
	return "unmarshalable"
}

var _ = Describe("Event Stream Handlers", func() {
	var (
		logger lager.Logger
		hub    event.Hub

		handler *handlers.EventStreamHandler

		server *httptest.Server
	)

	BeforeEach(func() {
		hub = event.NewHub()

		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		handler = handlers.NewEventStreamHandler(hub, logger)
	})

	AfterEach(func() {
		hub.Close()

		if server != nil {
			server.Close()
		}
	})

	Describe("EventStream", func() {
		var (
			response        *http.Response
			eventStreamDone chan struct{}
		)

		BeforeEach(func() {
			eventStreamDone = make(chan struct{})
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.EventStream(w, r)
				close(eventStreamDone)
			}))
		})

		JustBeforeEach(func() {
			var err error
			response, err = http.Get(server.URL)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when failing to subscribe to the event hub", func() {
			BeforeEach(func() {
				hub.Close()
			})

			It("returns an internal server error", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusInternalServerError))
			})
		})

		Context("when successfully subscribing to the event hub", func() {
			It("emits events from the hub to the connection", func() {
				reader := sse.NewReadCloser(response.Body)

				hub.Emit(fakeEvent{"A"})

				Ω(reader.Next()).Should(Equal(sse.Event{
					ID:   "0",
					Name: "fake",
					Data: []byte(`{"token":"A"}`),
				}))

				hub.Emit(fakeEvent{"B"})

				Ω(reader.Next()).Should(Equal(sse.Event{
					ID:   "1",
					Name: "fake",
					Data: []byte(`{"token":"B"}`),
				}))
			})

			It("returns Content-Type as text/event-stream", func() {
				Ω(response.Header.Get("Content-Type")).Should(Equal("text/event-stream; charset=utf-8"))
				Ω(response.Header.Get("Cache-Control")).Should(Equal("no-cache, no-store, must-revalidate"))
				Ω(response.Header.Get("Connection")).Should(Equal("keep-alive"))
			})

			Context("when the source provides an unmarshalable event", func() {
				It("closes the event stream to the client", func() {
					hub.Emit(unmarshalableEvent{Fn: func() {}})

					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Ω(err).Should(Equal(io.EOF))
				})
			})

			Context("when the event source returns an error", func() {
				BeforeEach(func() {
					hub.Close()
				})

				It("closes the client event stream", func() {
					reader := sse.NewReadCloser(response.Body)
					_, err := reader.Next()
					Ω(err).Should(Equal(io.EOF))
				})
			})

			Context("when the client closes the response body", func() {
				It("returns early", func() {
					reader := sse.NewReadCloser(response.Body)
					hub.Emit(fakeEvent{"A"})
					err := reader.Close()
					Ω(err).ShouldNot(HaveOccurred())
					Eventually(eventStreamDone, 10).Should(BeClosed())
				})
			})
		})
	})
})
