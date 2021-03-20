package processing

type LogEvent struct {
	line string
	containerName string
}

func (event *LogEvent) ContainerName() string {
	return event.containerName
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
