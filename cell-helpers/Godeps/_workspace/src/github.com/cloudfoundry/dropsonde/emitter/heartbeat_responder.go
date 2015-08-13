package emitter

import (
	"log"
	"runtime"
	"sync"

	"github.com/cloudfoundry/sonde-go/control"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type heartbeatResponder struct {
	instrumentedEmitter InstrumentedEmitter
	innerEmitter        ByteEmitter
	origin              string
	sync.Mutex
	closed bool
}

func NewHeartbeatResponder(byteEmitter ByteEmitter, origin string) (RespondingByteEmitter, error) {
	instrumentedEmitter, err := NewInstrumentedEmitter(byteEmitter)
	if err != nil {
		return nil, err
	}

	hbEmitter := &heartbeatResponder{
		instrumentedEmitter: instrumentedEmitter,
		innerEmitter:        byteEmitter,
		origin:              origin,
	}

	runtime.SetFinalizer(hbEmitter, (*heartbeatResponder).Close)

	return hbEmitter, nil
}

func (e *heartbeatResponder) Emit(data []byte) error {
	return e.instrumentedEmitter.Emit(data)
}

func (e *heartbeatResponder) Close() {
	e.Lock()
	defer e.Unlock()

	if e.closed {
		return
	}

	e.instrumentedEmitter.Close()
	e.closed = true
}

func (e *heartbeatResponder) Respond(controlMessage *control.ControlMessage) {
	hbEvent := e.instrumentedEmitter.GetHeartbeatEvent().(*events.Heartbeat)
	hbEvent.ControlMessageIdentifier = convertToEventUUID(controlMessage.GetIdentifier())
	hbEnvelope, err := Wrap(hbEvent, e.origin)
	if err != nil {
		log.Printf("Failed to wrap heartbeat event: %v\n", err)
		return
	}

	hbData, err := proto.Marshal(hbEnvelope)
	if err != nil {
		log.Printf("Failed to marshal heartbeat event: %v\n", err)
		return
	}

	err = e.innerEmitter.Emit(hbData)
	if err != nil {
		log.Printf("Problem while emitting heartbeat data: %v\n", err)
	}
}

func convertToEventUUID(uuid *control.UUID) *events.UUID {
	return &events.UUID{Low: uuid.Low, High: uuid.High}
}
