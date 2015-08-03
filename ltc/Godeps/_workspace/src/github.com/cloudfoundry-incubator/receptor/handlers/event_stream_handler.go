package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"
)

type EventStreamHandler struct {
	bbs    bbs.Client
	logger lager.Logger
}

func NewEventStreamHandler(bbs bbs.Client, logger lager.Logger) *EventStreamHandler {
	return &EventStreamHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *EventStreamHandler) EventStream(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("event-stream-handler")

	closeNotifier := w.(http.CloseNotifier).CloseNotify()
	sourceChan := make(chan events.EventSource)

	flusher := w.(http.Flusher)

	go func() {
		source, err := h.bbs.SubscribeToEvents()
		if err != nil {
			logger.Error("failed-to-subscribe-to-events", err)
			close(sourceChan)
			return
		}
		sourceChan <- source
	}()

	var source events.EventSource

	select {
	case source = <-sourceChan:
		if source == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case <-closeNotifier:
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
		bbsEvent, err := source.Next()
		if err != nil {
			logger.Error("failed-to-get-next-event", err)
			return
		}

		event, err := NewEventFromBBS(bbsEvent)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
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

func NewEventFromBBS(bbsEvent models.Event) (receptor.Event, error) {
	switch bbsEvent := bbsEvent.(type) {
	case *models.ActualLRPCreatedEvent:
		actualLRP, evacuating := bbsEvent.ActualLrpGroup.Resolve()
		return receptor.NewActualLRPCreatedEvent(serialization.ActualLRPProtoToResponse(actualLRP, evacuating)), nil
	case *models.ActualLRPChangedEvent:
		before, evacuating := bbsEvent.Before.Resolve()
		after, evacuating := bbsEvent.After.Resolve()
		return receptor.NewActualLRPChangedEvent(
			serialization.ActualLRPProtoToResponse(before, evacuating),
			serialization.ActualLRPProtoToResponse(after, evacuating),
		), nil
	case *models.ActualLRPRemovedEvent:
		actualLRP, evacuating := bbsEvent.ActualLrpGroup.Resolve()
		return receptor.NewActualLRPRemovedEvent(serialization.ActualLRPProtoToResponse(actualLRP, evacuating)), nil
	case *models.DesiredLRPCreatedEvent:
		return receptor.NewDesiredLRPCreatedEvent(serialization.DesiredLRPProtoToResponse(bbsEvent.DesiredLrp)), nil
	case *models.DesiredLRPChangedEvent:
		return receptor.NewDesiredLRPChangedEvent(
			serialization.DesiredLRPProtoToResponse(bbsEvent.Before),
			serialization.DesiredLRPProtoToResponse(bbsEvent.After),
		), nil
	case *models.DesiredLRPRemovedEvent:
		return receptor.NewDesiredLRPRemovedEvent(serialization.DesiredLRPProtoToResponse(bbsEvent.DesiredLrp)), nil
	}
	return nil, fmt.Errorf("unknown event type: %#v", bbsEvent)
}
