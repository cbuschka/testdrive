package container_runtime

import "fmt"

type ContainerEvent struct {
	eventType string
	id        string
	message   string
}

func NewContainerEvent(eventType string, id string, message string) *ContainerEvent {
	return &ContainerEvent{eventType: eventType, id: id, message: message}
}

func (event *ContainerEvent) String() string {
	return fmt.Sprintf("Event{eventType:%s, id:%s, message:%s}", event.eventType, event.id, event.message)
}

func (event *ContainerEvent) Id() string {
	return event.id
}

func (event *ContainerEvent) Type() string {
	return event.eventType
}

func (event *ContainerEvent) Message() string {
	return event.message
}
