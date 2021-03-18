package internal

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
)

type Config struct {
	Version  string                 `json:"version"`
	Services map[string]*TaskConfig `json:"services"`
	Tasks    map[string]*TaskConfig `json:"tasks"`
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

	cleanConfig(&config)

	return &config, nil
}

func cleanConfig(c *Config) {
	for name, task := range c.Tasks {
		cleanTaskConfig(name, task)
	}
}

func cleanTaskConfig(name string, task *TaskConfig) {

	cleanedDependencies := make([]string, 0)
	for _, dependency := range task.Dependencies {
		if dependency != name {
			cleanedDependencies = append(cleanedDependencies, dependency)
		} else {
			log.Printf("Removed self dependency in %s.\n", name)
		}
	}
	task.Dependencies = cleanedDependencies
}
