package internal

import "fmt"

type ContainerEvent struct {
	eventType string
	id        string
	message   string
}

func (event *ContainerEvent) Type() string {
	return event.eventType
}

func (event *ContainerEvent) String() string {
	return fmt.Sprintf("Event{eventType:%s, id:%s, message:%s}", event.eventType, event.id, event.message)
}

func (event *ContainerEvent) Id() string {
	return event.id
}
