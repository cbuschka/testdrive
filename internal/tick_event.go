package internal

type TickEvent struct {
}

func (event *TickEvent) Type() string {
	return "tick"
}

func (event *TickEvent) String() string {
	return "tick"
}

func (event *TickEvent) Id() string {
	return ""
}
