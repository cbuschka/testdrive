package internal

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"os"
)

type Docker struct {
	client  *client.Client
	context context.Context
}

func (docker *Docker) AddEventListener(ctx context.Context, listener func(event ContainerEvent)) {
	msgChannel, errChannel := docker.client.Events(ctx, types.EventsOptions{})
	for {
		select {
		case msg := <-msgChannel:
			listener(ContainerEvent{eventType: fmt.Sprintf("%s.%s", msg.Type, msg.Action), id: msg.ID})
		case err := <-errChannel:
			listener(ContainerEvent{eventType: "error", message: err.Error()})
			return
		case <-ctx.Done():
			return
		}
	}
}

func NewDocker() (*Docker, error) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	return &Docker{client: dockerClient, context: ctx}, nil
}

func (docker *Docker) CreateContainer(containerName string, image string) (string, error) {

	reader, err := docker.client.ImagePull(docker.context, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return "", err
	}

	containerConfig := container.Config{}
	containerConfig.Image = image
	response, err := docker.client.ContainerCreate(docker.context, &containerConfig,
		nil, nil, nil,
		containerName)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func (docker *Docker) StartContainer(containerId string) error {

	err := docker.client.ContainerStart(docker.context, containerId, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	return nil
}
