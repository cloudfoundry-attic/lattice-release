package sse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

type ReadCloser struct {
	lastID string

	buf         *bufio.Reader
	closeSource func() error
	closed      bool
}

func NewReadCloser(source io.ReadCloser) *ReadCloser {
	return &ReadCloser{
		closeSource: func() error { return source.Close() },
		buf:         bufio.NewReader(source),
	}
}

var alreadyClosedError = errors.New("ReadCloser already closed")

func (rc *ReadCloser) Close() error {
	if rc.closed {
		return alreadyClosedError
	}

	rc.closed = true

	return rc.closeSource()
}

func (rc *ReadCloser) Next() (Event, error) {
	var event Event

	// event ID defaults to last ID per the spec
	event.ID = rc.lastID

	// if an empty id is explicitly given, it sets the value and resets the last
	// id; track its presence with a bool to distinguish between zero-value
	idPresent := false

	prefix := []byte{}
	for {
		line, isPrefix, err := rc.buf.ReadLine()
		if err != nil {
			return Event{}, err
		}

		line = append(prefix, line...)

		if isPrefix {
			prefix = line
			continue
		} else {
			prefix = []byte{}
		}

		// empty line; dispatch event
		if len(line) == 0 {
			if len(event.Data) == 0 {
				// event had no data; skip it per the spec
				continue
			}

			if idPresent {
				// record last ID
				rc.lastID = event.ID
			}

			// trim terminating linebreak
			event.Data = event.Data[0 : len(event.Data)-1]

			// dispatch event
			return event, nil
		}

		if line[0] == ':' {
			// comment; skip
			continue
		}

		var field, value string

		segments := bytes.SplitN(line, []byte(":"), 2)
		if len(segments) == 1 {
			// line with no colon is just the field, with empty value
			field = string(segments[0])
		} else {
			field = string(segments[0])
			value = string(segments[1])
		}

		if len(value) > 0 {
			// trim only a single leading space
			if value[0] == ' ' {
				value = value[1:]
			}
		}

		switch field {
		case "id":
			idPresent = true
			event.ID = value
		case "event":
			event.Name = value
		case "data":
			event.Data = append(event.Data, []byte(value+"\n")...)
		case "retry":
			retryInMS, err := strconv.Atoi(value)
			if err == nil {
				event.Retry = time.Duration(retryInMS) * time.Millisecond
			}
		}
	}
}
