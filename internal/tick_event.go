package internal

type TickEvent struct {
}

func (event* TickEvent) Type() string {
	return "tick"
}
