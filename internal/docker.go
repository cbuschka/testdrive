package internal

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"time"
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

func (docker *Docker) CreateContainer(containerName string, image string, cmd []string) (string, error) {

	reader, err := docker.client.ImagePull(docker.context, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	lineReader := bufio.NewReader(reader)
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			if io.EOF == err {
				break
			} else {
				return "", err
			}
		}
		log.Infof("%s: %s", containerName, line)
	}

	containerConfig := container.Config{}
	containerConfig.Image = image
	containerConfig.Cmd = cmd
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

func (docker *Docker) StopContainer(containerId string, timeout time.Duration) error {

	err := docker.client.ContainerStop(docker.context, containerId, &timeout)
	if err != nil {
		return err
	}

	return nil
}

func (docker *Docker) DestroyContainer(containerId string) error {

	err := docker.client.ContainerRemove(docker.context, containerId, types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: true, Force: true})
	if err != nil {
		return err
	}

	return nil
}

func (docker *Docker) ReadContainerLogs(containerId string, ctx context.Context, listener func(line string)) error {
	reader, err := docker.client.ContainerLogs(context.Background(), containerId, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return nil
	}

	lineReader := bufio.NewReader(reader)
	defer reader.Close()

	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			return err
		} else {
			listener(string(line))
		}
	}
}

type DockerLogEvent struct {
	Status         string `json:"status"`
	ProgressDetail string `json:"progressDetail"`
	Id             string `json:"id"`
}

func (docker *Docker) ListContainers() (map[string]string, error) {

	stateByContainerId := make(map[string]string)
	containers, err := docker.client.ContainerList(docker.context, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, currContainer := range containers {
		stateByContainerId[currContainer.ID] = currContainer.State
	}

	return stateByContainerId, nil
}
