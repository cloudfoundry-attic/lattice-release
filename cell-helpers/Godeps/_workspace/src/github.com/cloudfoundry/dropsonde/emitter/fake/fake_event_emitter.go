package fake

import (
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type envelope struct {
	Event  events.Event
	Origin string
}

type FakeEventEmitter struct {
	ReturnError error
	messages    []envelope
	Origin      string
	isClosed    bool
	sync.RWMutex
}

func NewFakeEventEmitter(origin string) *FakeEventEmitter {
	return &FakeEventEmitter{Origin: origin}
}

func (f *FakeEventEmitter) Emit(e events.Event) error {

	f.Lock()
	defer f.Unlock()

	if f.ReturnError != nil {
		err := f.ReturnError
		f.ReturnError = nil
		return err
	}

	f.messages = append(f.messages, envelope{e, f.Origin})
	return nil
}

func (f *FakeEventEmitter) GetMessages() (messages []envelope) {
	f.Lock()
	defer f.Unlock()

	messages = make([]envelope, len(f.messages))
	copy(messages, f.messages)
	return
}

func (f *FakeEventEmitter) GetEvents() []events.Event {
	messages := f.GetMessages()
	events := []events.Event{}
	for _, msg := range messages {
		events = append(events, msg.Event)
	}
	return events
}

func (f *FakeEventEmitter) Close() {
	f.Lock()
	defer f.Unlock()
	f.isClosed = true
}

func (f *FakeEventEmitter) IsClosed() bool {
	f.RLock()
	defer f.RUnlock()
	return f.isClosed
}

func (f *FakeEventEmitter) Reset() {
	f.Lock()
	defer f.Unlock()

	f.isClosed = false
	f.messages = []envelope{}
	f.ReturnError = nil
}
