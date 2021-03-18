package internal

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type Config struct {
	Version  string                `json:"version"`
	Services map[string]TaskConfig `json:"services"`
	Tasks    map[string]TaskConfig `json:"tasks"`
}

type TaskConfig struct {
	Image        string   `json:"image"`
	Command      []string `json:"command"`
	Dependencies []string `json:"depends_on"`
	Healthcheck  interface{}
}

func LoadConfig(reader io.Reader) (*Config, error) {
	config := Config{}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := toJSON(bytes)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
