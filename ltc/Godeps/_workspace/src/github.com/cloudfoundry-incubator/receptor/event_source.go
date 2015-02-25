package receptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/vito/go-sse/sse"
)

var ErrUnrecognizedEventType = errors.New("unrecognized event type")

var ErrSourceClosed = errors.New("source closed")

type invalidPayloadError struct {
	jsonErr error
}

func NewInvalidPayloadError(jsonErr error) error {
	return invalidPayloadError{jsonErr: jsonErr}
}

func (e invalidPayloadError) Error() string {
	return fmt.Sprintf("invalid json payload: %s", e.jsonErr.Error())
}

type rawEventSourceError struct {
	rawError error
}

func NewRawEventSourceError(rawError error) error {
	return rawEventSourceError{rawError: rawError}
}

func (e rawEventSourceError) Error() string {
	return fmt.Sprintf("raw event source error: %s", e.rawError.Error())
}

type closeError struct {
	err error
}

func NewCloseError(err error) error {
	return closeError{err: err}
}

func (e closeError) Error() string {
	return fmt.Sprintf("error closing raw source: %s", e.err.Error())
}

//go:generate counterfeiter -o fake_receptor/fake_event_source.go . EventSource

// EventSource provides sequential access to a stream of events.
type EventSource interface {
	// Next reads the next event from the source. If the connection is lost, it
	// automatically reconnects.
	//
	// If the end of the stream is reached cleanly (which should actually never
	// happen), io.EOF is returned. If called after or during Close,
	// ErrSourceClosed is returned.
	Next() (Event, error)

	// Close releases the underlying response, interrupts any in-flight Next, and
	// prevents further calls to Next.
	Close() error
}

//go:generate counterfeiter -o fake_receptor/fake_raw_event_source.go . RawEventSource

type RawEventSource interface {
	Next() (sse.Event, error)
	Close() error
}

type eventSource struct {
	rawEventSource RawEventSource
}

func NewEventSource(raw RawEventSource) EventSource {
	return &eventSource{
		rawEventSource: raw,
	}
}

func (e *eventSource) Next() (Event, error) {
	rawEvent, err := e.rawEventSource.Next()
	if err != nil {
		switch err {
		case io.EOF:
			return nil, err

		case sse.ErrSourceClosed:
			return nil, ErrSourceClosed

		default:
			return nil, NewRawEventSourceError(err)
		}
	}

	return parseRawEvent(rawEvent)
}

func (e *eventSource) Close() error {
	err := e.rawEventSource.Close()
	if err != nil {
		return NewCloseError(err)
	}

	return nil
}

func parseRawEvent(rawEvent sse.Event) (Event, error) {
	switch EventType(rawEvent.Name) {
	case EventTypeDesiredLRPCreated:
		var event DesiredLRPCreatedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil

	case EventTypeDesiredLRPChanged:
		var event DesiredLRPChangedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil

	case EventTypeDesiredLRPRemoved:
		var event DesiredLRPRemovedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil

	case EventTypeActualLRPCreated:
		var event ActualLRPCreatedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil

	case EventTypeActualLRPChanged:
		var event ActualLRPChangedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil

	case EventTypeActualLRPRemoved:
		var event ActualLRPRemovedEvent
		err := json.Unmarshal(rawEvent.Data, &event)
		if err != nil {
			return nil, NewInvalidPayloadError(err)
		}

		return event, nil
	}

	return nil, ErrUnrecognizedEventType
}
