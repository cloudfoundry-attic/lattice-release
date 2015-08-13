package instrumented_handler_test

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/instrumented_handler"
	"github.com/cloudfoundry/sonde-go/events"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstrumentedHandler", func() {
	var fakeEmitter *fake.FakeEventEmitter
	var h http.Handler
	var req *http.Request

	var origin = "testHandler/41"

	BeforeEach(func() {
		fakeEmitter = fake.NewFakeEventEmitter(origin)

		var err error
		fh := fakeHandler{}
		h = instrumented_handler.InstrumentedHandler(fh, fakeEmitter)
		req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
		Expect(err).ToNot(HaveOccurred())
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("User-Agent", "our-testing-client")
	})

	AfterEach(func() {
		instrumented_handler.GenerateUuid = uuid.NewV4
	})

	Describe("request ID", func() {
		It("should add it to the request", func() {
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
		})

		It("should not add it to the request if it's already there", func() {
			id, _ := uuid.NewV4()
			req.Header.Set("X-CF-RequestID", id.String())
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).To(Equal(id.String()))
		})

		It("should create a valid one if it's given an invalid one", func() {
			req.Header.Set("X-CF-RequestID", "invalid")
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(Equal("invalid"))
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
		})

		It("should add it to the response", func() {
			id, _ := uuid.NewV4()
			req.Header.Set("X-CF-RequestID", id.String())
			response := httptest.NewRecorder()
			h.ServeHTTP(response, req)
			Expect(response.Header().Get("X-CF-RequestID")).To(Equal(id.String()))
		})

		It("should use an empty request ID if generating a new one fails", func() {
			instrumented_handler.GenerateUuid = func() (u *uuid.UUID, err error) {
				return nil, errors.New("test error")
			}
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).To(Equal("00000000-0000-0000-0000-000000000000"))
		})
	})

	Describe("event emission", func() {
		var requestId *uuid.UUID

		BeforeEach(func() {
			requestId, _ = uuid.NewV4()
			req.Header.Set("X-CF-RequestID", requestId.String())
		})

		Context("without an application ID or instanceIndex", func() {
			BeforeEach(func() {
				h.ServeHTTP(httptest.NewRecorder(), req)
			})

			It("should emit a start event with the right origin", func() {
				Expect(fakeEmitter.GetMessages()[0].Event).To(BeAssignableToTypeOf(new(events.HttpStart)))
				Expect(fakeEmitter.GetMessages()[0].Origin).To(Equal("testHandler/41"))
			})

			It("should emit a stop event", func() {
				Expect(fakeEmitter.GetMessages()[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))
				stopEvent := fakeEmitter.GetMessages()[1].Event.(*events.HttpStop)
				Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 123))
				Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 12))
			})
		})
	})

	Describe("satisfaction of interfaces", func() {

		var (
			rwChan chan http.ResponseWriter
			fh     storageHelperHandler
			h      http.Handler
		)

		BeforeEach(func() {
			rwChan = make(chan http.ResponseWriter, 1)
			fh = storageHelperHandler{rwChan: rwChan}
			h = instrumented_handler.InstrumentedHandler(fh, fakeEmitter)
		})

		Describe("http.Flusher", func() {
			It("panics if the underlying writer is not an http.Flusher", func() {
				h.ServeHTTP(basicResponseWriter{}, req)

				flusher := (<-rwChan).(http.Flusher)
				Expect(flusher.Flush).To(Panic())
			})

			It("delegates Flush to the underlying writer if it is an http.Flusher", func() {
				rw := completeResponseWriter{isFlushed: make(chan struct{})}
				h.ServeHTTP(rw, req)

				flusher := (<-rwChan).(http.Flusher)
				flusher.Flush()
				Expect(rw.isFlushed).To(BeClosed())
			})
		})

		Describe("http.Hijacker", func() {
			It("panics if the underlying writer is not an http.Hijacker", func() {
				h.ServeHTTP(basicResponseWriter{}, req)

				hijacker := (<-rwChan).(http.Hijacker)
				Expect(func() { hijacker.Hijack() }).To(Panic())
			})

			It("delegates Hijack to the underlying writer if it is an http.Hijacker", func() {
				rw := completeResponseWriter{isHijacked: make(chan struct{})}
				h.ServeHTTP(rw, req)

				hijacker := (<-rwChan).(http.Hijacker)
				hijacker.Hijack()
				Expect(rw.isHijacked).To(BeClosed())
			})
		})

		Describe("http.CloseNotifier", func() {
			It("panics if the underlying writer is not an http.CloseNotifier", func() {
				h.ServeHTTP(basicResponseWriter{}, req)

				closeNotifier := (<-rwChan).(http.CloseNotifier)
				Expect(func() { closeNotifier.CloseNotify() }).To(Panic())
			})

			It("delegates CloseNotify to the underlying writer if it is an http.CloseNotifier", func() {
				rw := completeResponseWriter{isCloseNotified: make(chan struct{})}
				h.ServeHTTP(rw, req)

				closeNotifier := (<-rwChan).(http.CloseNotifier)
				closeNotifier.CloseNotify()
				Expect(rw.isCloseNotified).To(BeClosed())
			})
		})
	})
})

// fakeHandler is a basic http.Handler that responds with some valid data
type fakeHandler struct{}

func (fh fakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Hello World!"))
	rw.WriteHeader(123)
}

// storageHelperHandler stores the ResponseWriter it is given during ServeHTTP
type storageHelperHandler struct {
	rwChan chan http.ResponseWriter
}

func (shh storageHelperHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	shh.rwChan <- rw
}

// basicResponseWriter is a minimal http.ResponseWriter
type basicResponseWriter struct {
}

func (basicResponseWriter) Header() http.Header {
	return http.Header{}
}

func (basicResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (basicResponseWriter) WriteHeader(int) {
}

// completeResponseWriter implements http.ResponseWriter, http.Flusher, http.Hijacker and http.CloseNotifier
type completeResponseWriter struct {
	isFlushed       chan struct{}
	isHijacked      chan struct{}
	isCloseNotified chan struct{}
}

func (completeResponseWriter) Header() http.Header {
	return http.Header{}
}

func (completeResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (completeResponseWriter) WriteHeader(int) {
}

func (rw completeResponseWriter) Flush() {
	close(rw.isFlushed)
}

func (rw completeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	close(rw.isHijacked)
	return fakeConn{}, nil, nil
}

func (rw completeResponseWriter) CloseNotify() <-chan bool {
	close(rw.isCloseNotified)
	return make(chan bool)
}

// fakeConn is a minimal net.Conn
type fakeConn struct{}

func (fakeConn) Read([]byte) (int, error) {
	return 0, nil
}

func (fakeConn) Write([]byte) (int, error) {
	return 0, nil
}

func (fakeConn) Close() error {
	return nil
}

func (fakeConn) LocalAddr() net.Addr {
	return fakeAddr{}
}

func (fakeConn) RemoteAddr() net.Addr {
	return fakeAddr{}
}

func (fakeConn) SetDeadline(time.Time) error {
	return nil
}

func (fakeConn) SetReadDeadline(time.Time) error {
	return nil
}

func (fakeConn) SetWriteDeadline(time.Time) error {
	return nil
}

// fakeAddr is a minimal net.Addr
type fakeAddr struct{}

func (fakeAddr) Network() string {
	return ""
}

func (fakeAddr) String() string {
	return ""
}
