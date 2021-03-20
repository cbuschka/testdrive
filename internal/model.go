package internal

import "time"

type Model struct {
	containers map[string]*Container
}

func NewModel() *Model {
	return &Model{containers: make(map[string]*Container, 0)}
}

func (model *Model) GetContainerByContainerId(containerId string) *Container {
	for _, task := range model.containers {
		if task.containerId == containerId {
			return task
		}
	}

	return nil
}

func (model *Model) AddContainer(container *Container) error {
	model.containers[container.name] = container
	return nil
}

func (model *Model) AllDependenciesReady(container *Container) bool {
	for _, dependency := range container.config.Dependencies {
		if model.containers[dependency].status != Ready {
			log.Debugf("Dependency %s->%s not ready.\n", container.name, dependency)
			return false
		}
	}

	return true
}

func (model *Model) AllServiceDependenciesReadyAndAllTaskDependenciesStopped(container *Container) bool {
	for _, dependency := range container.config.Dependencies {
		if model.containers[dependency].containerType == ContainerType_Service && model.containers[dependency].status != Ready {
			return false
		}
		if model.containers[dependency].containerType == ContainerType_Task && model.containers[dependency].status != Stopped {
			return false
		}
	}

	return true
}

func (model *Model) getCreatableContainers(containerType string) []*Container {
	createableContainers := make([]*Container, 0)
	for _, container := range model.containers {
		if container.status == New && container.containerType == containerType {
			createableContainers = append(createableContainers, container)
		}
	}

	return createableContainers
}

func (model *Model) getStartableServiceContainers() []*Container {
	startableContainers := make([]*Container, 0)
	for _, container := range model.containers {
		if container.status == Created && container.containerType == ContainerType_Service && model.AllDependenciesReady(container) {
			startableContainers = append(startableContainers, container)
		}
	}

	return startableContainers
}

func (model *Model) getStartableTaskContainers() []*Container {
	startableContainers := make([]*Container, 0)
	for _, container := range model.containers {
		if container.status == Created && container.containerType == ContainerType_Task && model.AllServiceDependenciesReadyAndAllTaskDependenciesStopped(container) {
			startableContainers = append(startableContainers, container)
		}
	}

	return startableContainers
}

func (model *Model) ContainerCreating(container *Container) {
	if container.status != New {
		log.Warningf("Marking container %s as Creating, expected it to be New, but is in status %s.", container.name, container.status)
	}

	container.status = Creating
	log.Debugf("Container %s marked as creating.\n", container.name)
}

func (model *Model) ContainerStarted(container *Container) {

	if container.status != Starting {
		log.Warningf("Marking container %s as Started, expected it to be Starting, but is in status %s.", container.name, container.status)
	}

	container.status = Started
	log.Infof("Container %s (%s) has started.\n", container.name, container.containerId)
}

func (model *Model) ContainerReady(container *Container) {

	if container.status != Started && container.status != Starting {
		log.Warningf("Marking container %s as Ready, expected it to be Started or Starting, but is in status %s.", container.name, container.status)
	}

	container.status = Ready
	log.Infof("Container %s (%s) is ready.\n", container.name, container.containerId)
}

func (model *Model) ContainerCreated(container *Container) {

	if container.status != Creating {
		log.Warningf("Marking container %s as Created, expected it to be Creating, but is in status %s.", container.name, container.status)
	}

	container.status = Created
	container.createStartedAt = time.Now()
	log.Debugf("Container %s (%s) has been created.\n", container.name, container.containerId)
}

func (model *Model) ContainerStarting(container *Container) {

	if container.status != Created {
		log.Warningf("Marking container %s as Starting, expected it to be Created, but is in status %s.", container.name, container.status)
	}

	container.status = Starting
	container.startStartededAt = time.Now()
	log.Debugf("Container %s marked as starting.\n", container.name)
}

func (model *Model) ContainerFailed(container *Container) {
	container.status = Failed
	container.failedAt = time.Now()
	log.Warningf("Container %s has failed.\n", container.name)
}

func (model *Model) ContainerStopped(container *Container) {
	container.status = Stopped
	log.Infof("Container %s (%s) has stopped.\n", container.name, container.containerId)
}

func (model *Model) ContainerDestroyed(container *Container) {

	if container.status != Destroying {
		log.Warningf("Marking container %s as Destroyed, expected it to be Destroying, but is in status %s.", container.name, container.status)
	}

	container.status = Destroyed
	log.Debugf("Container %s (%s) has been destroyed.\n", container.name, container.containerId)
}
