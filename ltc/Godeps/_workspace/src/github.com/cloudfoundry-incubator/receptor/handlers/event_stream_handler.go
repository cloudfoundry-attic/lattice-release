package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"
)

type EventStreamHandler struct {
	hub    event.Hub
	logger lager.Logger
}

func NewEventStreamHandler(hub event.Hub, logger lager.Logger) *EventStreamHandler {
	return &EventStreamHandler{
		hub:    hub,
		logger: logger,
	}
}

func (h *EventStreamHandler) EventStream(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("event-stream-handler")

	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	flusher := w.(http.Flusher)

	source, err := h.hub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer source.Close()

	go func() {
		<-closeNotifier
		source.Close()
	}()

	w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)

	flusher.Flush()

	eventID := 0
	for {
		event, err := source.Next()
		if err != nil {
			logger.Error("failed-to-get-next-event", err)
			return
		}

		payload, err := json.Marshal(event)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
			return
		}

		err = sse.Event{
			ID:   strconv.Itoa(eventID),
			Name: string(event.EventType()),
			Data: payload,
		}.Write(w)
		if err != nil {
			break
		}

		flusher.Flush()

		eventID++
	}
}
