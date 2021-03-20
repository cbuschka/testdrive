package config

import (
	"github.com/cbuschka/testdrive/internal/log"
	"io"
	"io/ioutil"
	"gopkg.in/yaml.v3"
)

type TestdriveConfig struct {
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

func LoadConfig(reader io.Reader) (*TestdriveConfig, error) {
	config := TestdriveConfig{}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, err
	}

	err = node.Decode(&config)
	if err != nil {
		return nil, err
	}

	cleanConfig(&config)

	return &config, nil
}

func cleanConfig(c *TestdriveConfig) {
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
			log.Debugf("Removed self dependency in %s.", name)
		}
	}
	task.Dependencies = cleanedDependencies
}
