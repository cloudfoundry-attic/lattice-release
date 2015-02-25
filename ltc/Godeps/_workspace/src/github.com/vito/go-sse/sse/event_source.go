package sse

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type BadResponseError struct {
	Response *http.Response
}

func (err BadResponseError) Error() string {
	return fmt.Sprintf("bad response from event source: %s", err.Response.Status)
}

// EventSource behaves like the EventSource interface from the Server-Sent
// Events spec implemented in many browsers.  See
// http://www.w3.org/TR/eventsource/#the-eventsource-interface for details.
//
// To use, optionally call Connect(), and then call Next(). If Next() is called
// prior to Connect(), it will connect for you.
//
// Next() is often called asynchronously in a loop so that the event source can
// be closed. Next() will block on reading from the server.
//
// If Close() is called while reading an event, Next() will return early, and
// subsequent calls to Next() will return early. To read new events, Connect()
// must be called.
//
// If an EOF is received, Next() returns io.EOF, and subsequent calls to Next()
// will return early. To read new events, Connect() must be called.
type EventSource struct {
	client        *http.Client
	createRequest func() *http.Request

	currentReadCloser *ReadCloser
	lastEventID       string
	retryInterval     time.Duration
	lock              sync.Mutex

	closeOnce *sync.Once
	closed    chan struct{}
}

func NewEventSource(client *http.Client, defaultRetryInterval time.Duration, requestCreator func() *http.Request) *EventSource {
	return &EventSource{
		client:        client,
		createRequest: requestCreator,

		retryInterval: defaultRetryInterval,

		closeOnce: new(sync.Once),
		closed:    make(chan struct{}),
	}
}

func Connect(client *http.Client, defaultRetryInterval time.Duration, requestCreator func() *http.Request) (*EventSource, error) {
	source := NewEventSource(client, defaultRetryInterval, requestCreator)

	readCloser, err := source.establishConnection()
	if err != nil {
		return nil, err
	}

	source.currentReadCloser = readCloser

	return source, nil
}

func (source *EventSource) Next() (Event, error) {
	select {
	case <-source.closed:
		return Event{}, ErrSourceClosed
	default:
	}

	for {
		readCloser, err := source.ensureReadCloser()
		if err != nil {
			return Event{}, err
		}

		event, err := readCloser.Next()
		if err == nil {
			source.lastEventID = event.ID

			if event.Retry != 0 {
				source.retryInterval = event.Retry
			}

			return event, nil
		}

		if err == io.EOF {
			return Event{}, err
		}

		readCloser.Close()

		if err := source.waitForRetry(); err != nil {
			return Event{}, err
		}
	}

	panic("unreachable")
}

func (source *EventSource) Close() error {
	source.lock.Lock()
	defer source.lock.Unlock()

	source.closeOnce.Do(func() {
		close(source.closed)
	})

	if source.currentReadCloser != nil {
		err := source.currentReadCloser.Close()
		if err != nil {
			return err
		}

		source.currentReadCloser = nil
	}

	return nil
}

func (source *EventSource) ensureReadCloser() (*ReadCloser, error) {
	source.lock.Lock()

	if source.currentReadCloser == nil {
		source.lock.Unlock()

		newReadCloser, err := source.establishConnection()
		if err != nil {
			return nil, err
		}

		source.lock.Lock()

		select {
		case <-source.closed:
			source.lock.Unlock()
			newReadCloser.Close()
			return nil, ErrSourceClosed

		default:
			source.currentReadCloser = newReadCloser
		}
	}

	readCloser := source.currentReadCloser

	source.lock.Unlock()

	return readCloser, nil
}

func (source *EventSource) establishConnection() (*ReadCloser, error) {
	for {
		req := source.createRequest()

		req.Header.Set("Last-Event-ID", source.lastEventID)

		res, err := source.client.Do(req)
		if err != nil {
			err := source.waitForRetry()
			if err != nil {
				return nil, err
			}

			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			return NewReadCloser(res.Body), nil

		// reestablish the connection
		case http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			res.Body.Close()

			err := source.waitForRetry()
			if err != nil {
				return nil, err
			}

			continue

		// fail the connection
		default:
			res.Body.Close()

			return nil, BadResponseError{
				Response: res,
			}
		}
	}
}

func (source *EventSource) waitForRetry() error {
	source.lock.Lock()
	source.currentReadCloser = nil
	source.lock.Unlock()

	select {
	case <-time.After(source.retryInterval):
		return nil
	case <-source.closed:
		return ErrSourceClosed
	}
}
