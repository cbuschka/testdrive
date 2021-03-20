package model

import (
	"github.com/cbuschka/testdrive/internal/log"
	"time"
)


type Model struct {
	Containers map[string]*Container
}

func NewModel() *Model {
	return &Model{Containers: make(map[string]*Container, 0)}
}

func (model *Model) GetContainerByContainerId(containerId string) *Container {
	for _, task := range model.Containers {
		if task.ContainerId == containerId {
			return task
		}
	}

	return nil
}

func (model *Model) AddContainer(container *Container) error {
	model.Containers[container.Name] = container
	return nil
}

func (model *Model) AllDependenciesReady(container *Container) bool {
	for _, dependency := range container.Config.Dependencies {
		if model.Containers[dependency].Status != Ready {
			log.Debugf("Dependency %s->%s not ready.", container.Name, dependency)
			return false
		}
	}

	return true
}

func (model *Model) AllServiceDependenciesReadyAndAllTaskDependenciesStopped(container *Container) bool {
	for _, dependency := range container.Config.Dependencies {
		if model.Containers[dependency].ContainerType == ContainerType_Service && model.Containers[dependency].Status != Ready {
			return false
		}
		if model.Containers[dependency].ContainerType == ContainerType_Task && model.Containers[dependency].Status != Stopped {
			return false
		}
	}

	return true
}

func (model *Model) GetCreatableContainers(containerType string) []*Container {
	createableContainers := make([]*Container, 0)
	for _, container := range model.Containers {
		if container.Status == New && container.ContainerType == containerType {
			createableContainers = append(createableContainers, container)
		}
	}

	return createableContainers
}

func (model *Model) GetStartableServiceContainers() []*Container {
	startableContainers := make([]*Container, 0)
	for _, container := range model.Containers {
		if container.Status == Created && container.ContainerType == ContainerType_Service && model.AllDependenciesReady(container) {
			startableContainers = append(startableContainers, container)
		}
	}

	return startableContainers
}

func (model *Model) GetStartableTaskContainers() []*Container {
	startableContainers := make([]*Container, 0)
	for _, container := range model.Containers {
		if container.Status == Created && container.ContainerType == ContainerType_Task && model.AllServiceDependenciesReadyAndAllTaskDependenciesStopped(container) {
			startableContainers = append(startableContainers, container)
		}
	}

	return startableContainers
}

func (model *Model) ContainerCreating(container *Container) {

	if container.Status == Creating {
		return
	}

	if container.Status != New {
		log.Warningf("Marking container %s as Creating, expected it to be New, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Creating
	log.Debugf("Container %s marked as creating.\n", container.Name)
}

func (model *Model) ContainerStarted(container *Container) {

	if container.Status == Started {
		return
	}

	if container.Status != Starting {
		log.Warningf("Marking container %s as Started, expected it to be Starting, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Started
	log.Infof("Container %s (%s) has started.", container.Name, container.ContainerId)
}

func (model *Model) ContainerReady(container *Container) {

	if container.Status == Ready {
		return
	}

	if container.Status != Started && container.Status != Starting {
		log.Warningf("Marking container %s as Ready, expected it to be Started or Starting, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Ready
	log.Infof("Container %s (%s) is ready.\n", container.Name, container.ContainerId)
}

func (model *Model) ContainerCreated(container *Container) {

	if container.Status == Created {
		return
	}

	if container.Status != Creating {
		log.Warningf("Marking container %s as Created, expected it to be Creating, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Created
	container.CreateStartedAt = time.Now()
	log.Debugf("Container %s (%s) has been created.", container.Name, container.ContainerId)
}

func (model *Model) ContainerStarting(container *Container) {

	if container.Status == Starting {
		return
	}

	if container.Status != Created {
		log.Warningf("Marking container %s as Starting, expected it to be Created, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Starting
	container.StartStartededAt = time.Now()
	log.Debugf("Container %s marked as starting.", container.Name)
}

func (model *Model) ContainerFailed(container *Container) {
	if container.Status == Failed {
		return
	}

	container.Status = Failed
	container.FailedAt = time.Now()
	log.Warningf("Container %s has failed.", container.Name)
}

func (model *Model) ContainerStopped(container *Container) {
	if container.Status == Stopped {
		return
	}

	container.Status = Stopped
	log.Infof("Container %s (%s) has stopped.", container.Name, container.ContainerId)
}

func (model *Model) ContainerDestroyed(container *Container) {

	if container.Status == Destroyed {
		return
	}

	if container.Status != Destroying {
		log.Warningf("Marking container %s as Destroyed, expected it to be Destroying, but is in Status %s.", container.Name, container.Status)
	}

	container.Status = Destroyed
	log.Debugf("Container %s (%s) has been destroyed.", container.Name, container.ContainerId)
}
