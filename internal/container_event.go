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
	return fmt.Sprintf("%v", *event)
}

func (event *ContainerEvent) Id() string {
	return event.id
}
