package internal

type Dialog struct{}

var dialog = Dialog{}

func (dialog *Dialog) ContainerOutput(containerName *string, line *string) {
	log.Infof("%s| %s", *containerName, *line)
}
