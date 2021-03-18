package internal

import "context"

type ContainerRuntime interface {
	CreateContainer(containerName string, imageName string) (string, error)
	AddEventListener(ctx context.Context, listener func(event ContainerEvent))
	StartContainer(containerId string) error
}
