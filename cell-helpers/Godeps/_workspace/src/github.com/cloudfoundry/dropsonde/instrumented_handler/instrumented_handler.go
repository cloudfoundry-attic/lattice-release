package instrumented_handler

import (
	"bufio"
	"net"
	"net/http"

	"log"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	uuid "github.com/nu7hatch/gouuid"
)

type instrumentedHandler struct {
	handler http.Handler
	emitter emitter.EventEmitter
}

/*
Helper for creating an Instrumented Handler which will delegate to the given http.Handler.
*/
func InstrumentedHandler(handler http.Handler, emitter emitter.EventEmitter) http.Handler {
	return &instrumentedHandler{handler, emitter}
}

/*
Wraps the given http.Handler ServerHTTP function
Will provide accounting metrics for the http.Request / http.Response life-cycle
*/
func (ih *instrumentedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestId, err := uuid.ParseHex(req.Header.Get("X-CF-RequestID"))
	if err != nil {
		requestId, err = GenerateUuid()
		if err != nil {
			log.Printf("failed to generated request ID: %v\n", err)
			requestId = &uuid.UUID{}
		}
		req.Header.Set("X-CF-RequestID", requestId.String())
	}
	rw.Header().Set("X-CF-RequestID", requestId.String())

	startEvent := factories.NewHttpStart(req, events.PeerType_Server, requestId)

	err = ih.emitter.Emit(startEvent)
	if err != nil {
		log.Printf("failed to emit start event: %v\n", err)
	}

	instrumentedWriter := &instrumentedResponseWriter{writer: rw, statusCode: 200}
	ih.handler.ServeHTTP(instrumentedWriter, req)

	stopEvent := factories.NewHttpStop(req, instrumentedWriter.statusCode, instrumentedWriter.contentLength, events.PeerType_Server, requestId)

	err = ih.emitter.Emit(stopEvent)
	if err != nil {
		log.Printf("failed to emit stop event: %v\n", err)
	}
}

type instrumentedResponseWriter struct {
	writer        http.ResponseWriter
	contentLength int64
	statusCode    int
}

func (irw *instrumentedResponseWriter) Header() http.Header {
	return irw.writer.Header()
}

func (irw *instrumentedResponseWriter) Write(data []byte) (int, error) {
	writeCount, err := irw.writer.Write(data)
	irw.contentLength += int64(writeCount)
	return writeCount, err
}

func (irw *instrumentedResponseWriter) WriteHeader(statusCode int) {
	irw.statusCode = statusCode
	irw.writer.WriteHeader(statusCode)
}

func (irw *instrumentedResponseWriter) Flush() {
	flusher, ok := irw.writer.(http.Flusher)

	if !ok {
		panic("Called Flush on an InstrumentedResponseWriter that wraps a non-Flushable writer.")
	}

	flusher.Flush()
}

func (irw *instrumentedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := irw.writer.(http.Hijacker)

	if !ok {
		panic("Called Hijack on an InstrumentedResponseWriter that wraps a non-Hijackable writer")
	}

	return hijacker.Hijack()
}

func (irw *instrumentedResponseWriter) CloseNotify() <-chan bool {
	notifier, ok := irw.writer.(http.CloseNotifier)

	if !ok {
		panic("Called CloseNotify on an InstrumentedResponseWriter that wraps a non-CloseNotifiable writer")
	}

	return notifier.CloseNotify()
}

var GenerateUuid = uuid.NewV4
