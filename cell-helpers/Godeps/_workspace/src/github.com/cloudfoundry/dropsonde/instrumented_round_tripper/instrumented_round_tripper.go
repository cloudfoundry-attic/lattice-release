package instrumented_round_tripper

import (
	"log"
	"net/http"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	uuid "github.com/nu7hatch/gouuid"
)

type instrumentedRoundTripper struct {
	roundTripper http.RoundTripper
	emitter      emitter.EventEmitter
}

type instrumentedCancelableRoundTripper struct {
	instrumentedRoundTripper *instrumentedRoundTripper
}

/*
Helper for creating an InstrumentedRoundTripper which will delegate to the given RoundTripper
*/
func InstrumentedRoundTripper(roundTripper http.RoundTripper, emitter emitter.EventEmitter) http.RoundTripper {
	irt := &instrumentedRoundTripper{roundTripper, emitter}

	_, ok := roundTripper.(canceler)
	if ok {
		return &instrumentedCancelableRoundTripper{
			instrumentedRoundTripper: irt,
		}
	}

	return irt
}

/*
Wraps the RoundTrip function of the given RoundTripper.
Will provide accounting metrics for the http.Request / http.Response life-cycle
Callers of RoundTrip are responsible for setting the ‘X-CF-RequestID’ field in the request header if they have one.
Callers are also responsible for setting the ‘X-CF-ApplicationID’ and ‘X-CF-InstanceIndex’ fields in the request header if they are known.
*/
func (irt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	requestId, err := GenerateUuid()
	if err != nil {
		log.Printf("failed to generated request ID: %v\n", err)
		requestId = &uuid.UUID{}
	}

	httpStart := factories.NewHttpStart(req, events.PeerType_Client, requestId)

	parentRequestId, err := uuid.ParseHex(req.Header.Get("X-CF-RequestID"))
	if err == nil {
		httpStart.ParentRequestId = factories.NewUUID(parentRequestId)
	}

	req.Header.Set("X-CF-RequestID", requestId.String())

	err = irt.emitter.Emit(httpStart)
	if err != nil {
		log.Printf("failed to emit start event: %v\n", err)
	}

	resp, roundTripErr := irt.roundTripper.RoundTrip(req)

	var httpStop *events.HttpStop
	if roundTripErr != nil {
		httpStop = factories.NewHttpStop(req, 0, 0, events.PeerType_Client, requestId)
	} else {
		httpStop = factories.NewHttpStop(req, resp.StatusCode, resp.ContentLength, events.PeerType_Client, requestId)
	}

	err = irt.emitter.Emit(httpStop)
	if err != nil {
		log.Printf("failed to emit stop event: %v\n", err)
	}

	return resp, roundTripErr
}

func (icrt *instrumentedCancelableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return icrt.instrumentedRoundTripper.RoundTrip(req)
}

func (icrt *instrumentedCancelableRoundTripper) CancelRequest(req *http.Request) {
	cancelableTransport := icrt.instrumentedRoundTripper.roundTripper.(canceler)
	cancelableTransport.CancelRequest(req)
}

var GenerateUuid = uuid.NewV4

type canceler interface {
	CancelRequest(*http.Request)
}
