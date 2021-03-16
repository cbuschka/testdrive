package internal

import "os"

type Session struct {
	id    string
	model *Model
}

func NewSession() *Session {
	model := NewModel()
	session := Session{id: "1", model: model}
	return &session
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

	err = ApplyConfig(session, config)
	if err != nil {
		return err
	}

	return nil
}
