package emitter

import (
	"errors"
	"sync"

	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
)

type InstrumentedEmitter interface {
	ByteEmitter
	GetHeartbeatEvent() events.Event
}

type instrumentedEmitter struct {
	wrappedEmitter         ByteEmitter
	mutex                  *sync.RWMutex
	ReceivedMetricsCounter uint64
	SentMetricsCounter     uint64
	ErrorCounter           uint64
}

func (emitter *instrumentedEmitter) Emit(data []byte) error {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	emitter.ReceivedMetricsCounter++

	err := emitter.wrappedEmitter.Emit(data)
	if err != nil {
		emitter.ErrorCounter++
	} else {
		emitter.SentMetricsCounter++
	}

	return err
}

func NewInstrumentedEmitter(wrappedEmitter ByteEmitter) (InstrumentedEmitter, error) {
	if wrappedEmitter == nil {
		return nil, errors.New("wrappedEmitter is nil")
	}

	emitter := &instrumentedEmitter{wrappedEmitter: wrappedEmitter, mutex: &sync.RWMutex{}}
	return emitter, nil
}

func (emitter *instrumentedEmitter) Close() {
	emitter.wrappedEmitter.Close()
}

func (emitter *instrumentedEmitter) GetHeartbeatEvent() events.Event {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	return factories.NewHeartbeat(emitter.SentMetricsCounter, emitter.ReceivedMetricsCounter, emitter.ErrorCounter)
}
