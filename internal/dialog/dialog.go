package dialog

import "github.com/cbuschka/testdrive/internal/log"

func ContainerOutput(containerName *string, line *string) {
	log.Infof("%s: %s", *containerName, *line)
}
