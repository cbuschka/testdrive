package container_runtime

import (
	"context"
	"time"
)

type ContainerRuntime interface {
	GetContainerExitCode(containerId string) (int, error)
	CreateContainer(containerName string, imageName string, cmd []string) (string, error)
	AddEventListener(ctx context.Context, listener func(event ContainerEvent))
	StartContainer(containerId string) error
	StopContainer(containerId string, timeout time.Duration) error
	DestroyContainer(containerId string) error
	ReadContainerLogs(containerId string, ctx context.Context, listener func(line string)) error
	ListContainers() (map[string]string, error)
}
