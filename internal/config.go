package internal

import (
	"io"
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version  string                      `yaml:"version"`
	Services map[string]*ContainerConfig `yaml:"services"`
	Tasks    map[string]*ContainerConfig `yaml:"tasks"`
}

type ContainerConfig struct {
	Image        string   `yaml:"image"`
	Command      []string `yaml:"command"`
	Dependencies []string `yaml:"depends_on"`
	Healthcheck  interface{}
}

func LoadConfig(reader io.Reader) (*Config, error) {
	config := Config{}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	cleanConfig(&config)

	return &config, nil
}

func cleanConfig(c *Config) {
	for name, task := range c.Tasks {
		cleanContainerConfig(name, task)
	}
}

func cleanContainerConfig(name string, task *ContainerConfig) {

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
