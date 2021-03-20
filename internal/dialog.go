package internal

import "github.com/cbuschka/testdrive/internal/log"

type Dialog struct{}

var dialog = Dialog{}

func (dialog *Dialog) ContainerOutput(containerName *string, line *string) {
	log.Infof("%s: %s", *containerName, *line)
}
