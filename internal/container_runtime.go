package internal

import "context"

type ContainerRuntime interface {
	CreateContainer(containerName string, imageName string, cmd []string) (string, error)
	AddEventListener(ctx context.Context, listener func(event ContainerEvent))
	StartContainer(containerId string) error
	StopContainer(containerId string) error
	DestroyContainer(containerId string) error
	ReadContainerLogs(containerId string, ctx context.Context, listener func(line string)) error
	ListContainers() (map[string]string, error)
}
