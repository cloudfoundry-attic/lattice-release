package models

import (
	"github.com/gogo/protobuf/proto"
)

type Event interface {
	EventType() string
	Key() string
	proto.Message
}

const (
	EventTypeInvalid = ""

	EventTypeDesiredLRPCreated = "desired_lrp_created"
	EventTypeDesiredLRPChanged = "desired_lrp_changed"
	EventTypeDesiredLRPRemoved = "desired_lrp_removed"

	EventTypeActualLRPCreated = "actual_lrp_created"
	EventTypeActualLRPChanged = "actual_lrp_changed"
	EventTypeActualLRPRemoved = "actual_lrp_removed"
)

func NewDesiredLRPCreatedEvent(desiredLRP *DesiredLRP) *DesiredLRPCreatedEvent {
	return &DesiredLRPCreatedEvent{
		DesiredLrp: desiredLRP,
	}
}

func (event *DesiredLRPCreatedEvent) EventType() string {
	return EventTypeDesiredLRPCreated
}

func (event *DesiredLRPCreatedEvent) Key() string {
	return event.DesiredLrp.GetProcessGuid()
}

func NewDesiredLRPChangedEvent(before, after *DesiredLRP) *DesiredLRPChangedEvent {
	return &DesiredLRPChangedEvent{
		Before: before,
		After:  after,
	}
}

func (event *DesiredLRPChangedEvent) EventType() string {
	return EventTypeDesiredLRPChanged
}

func (event *DesiredLRPChangedEvent) Key() string {
	return event.Before.GetProcessGuid()
}

func NewDesiredLRPRemovedEvent(desiredLRP *DesiredLRP) *DesiredLRPRemovedEvent {
	return &DesiredLRPRemovedEvent{
		DesiredLrp: desiredLRP,
	}
}

func (event *DesiredLRPRemovedEvent) EventType() string {
	return EventTypeDesiredLRPRemoved
}

func (event DesiredLRPRemovedEvent) Key() string {
	return event.DesiredLrp.GetProcessGuid()
}

func NewActualLRPChangedEvent(before, after *ActualLRPGroup) *ActualLRPChangedEvent {
	return &ActualLRPChangedEvent{
		Before: before,
		After:  after,
	}
}

func (event *ActualLRPChangedEvent) EventType() string {
	return EventTypeActualLRPChanged
}

func (event *ActualLRPChangedEvent) Key() string {
	actualLRP, _ := event.Before.Resolve()
	return actualLRP.GetInstanceGuid()
}

func NewActualLRPRemovedEvent(actualLRPGroup *ActualLRPGroup) *ActualLRPRemovedEvent {
	return &ActualLRPRemovedEvent{
		ActualLrpGroup: actualLRPGroup,
	}
}

func (event *ActualLRPRemovedEvent) EventType() string {
	return EventTypeActualLRPRemoved
}

func (event *ActualLRPRemovedEvent) Key() string {
	actualLRP, _ := event.ActualLrpGroup.Resolve()
	return actualLRP.GetInstanceGuid()
}

func NewActualLRPCreatedEvent(actualLRPGroup *ActualLRPGroup) *ActualLRPCreatedEvent {
	return &ActualLRPCreatedEvent{
		ActualLrpGroup: actualLRPGroup,
	}
}

func (event *ActualLRPCreatedEvent) EventType() string {
	return EventTypeActualLRPCreated
}

func (event *ActualLRPCreatedEvent) Key() string {
	actualLRP, _ := event.ActualLrpGroup.Resolve()
	return actualLRP.GetInstanceGuid()
}
