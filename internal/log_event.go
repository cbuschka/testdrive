package internal

type LogEvent struct {
	line string
}

func (event *LogEvent) Type() string {
	return "log"
}

func (event *LogEvent) String() string {
	return event.line
}

func (event *LogEvent) Id() string {
	return ""
}
