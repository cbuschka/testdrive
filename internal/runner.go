package internal

import (
	"fmt"
	"time"
	)

func tick(eventQueue chan Event) {
	for true {
		eventQueue <- &TickEvent{}
		time.Sleep(10 * time.Millisecond)
	}
}

func (session *Session) Run() (int, error) {

	go tick(session.eventQueue)

	for event := range session.eventQueue {
		if event == nil {
			break
		}
		fmt.Printf("%s\n", event.Type())
	}

	return 0, nil
}
