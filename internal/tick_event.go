package internal

import "fmt"

type TickEvent struct {
}

func (event* TickEvent) Type() string {
	return "tick"
}

func (event* TickEvent) String() string {
	return fmt.Sprintf("%v", *event)
}

func (event* TickEvent) Id() string {
	return ""
}

