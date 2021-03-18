package internal

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type Config struct {
	Version  string          `json:"version"`
	Services map[string]Task `json:"services"`
	Tasks    map[string]Task `json:"tasks"`
}

type Task struct {
	Image        string   `json:"image"`
	Dependencies []string `json:"depends_on"`
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
