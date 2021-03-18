package internal

import "os"

type Session struct {
	id               string
	config           *Config
	model            *Model
	eventQueue       chan Event
	phase            Phase
	containerRuntime ContainerRuntime
}

func NewSession() (*Session, error) {
	model := NewModel()
	docker, err := NewDocker()
	if err != nil {
		return nil, err
	}
	session := Session{id: "1", config: nil, model: model, eventQueue: make(chan Event, 100),
		phase: Running, containerRuntime: ContainerRuntime(docker)}
	return &session, nil
}

func (session *Session) LoadConfig(file string) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	config, err := LoadConfig(reader)
	if err != nil {
		return err
	}

	session.config = config

	return nil
}
