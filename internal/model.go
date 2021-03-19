package internal

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
	container.status = Creating
	log.Debugf("Container %s marked as creating.\n", container.name)
}

func (model *Model) TaskStarted(container *Container) {
	container.status = Started
	log.Debugf("Container %s marked as started.\n", container.name)
}

func (model *Model) ContainerReady(container *Container) {
	container.status = Ready
	log.Debugf("Container %s marked as ready.\n", container.name)
}

func (model *Model) ContainerCreated(container *Container) {
	container.status = Created
	log.Debugf("Container %s marked as created.\n", container.name)
}

func (model *Model) ContainerStarting(container *Container) {
	container.status = Starting
	log.Debugf("Container %s marked as starting.\n", container.name)
}

func (model *Model) ContainerFailed(container *Container) {
	container.status = Failed
	log.Debugf("Container %s marked as failed.\n", container.name)
}

func (model *Model) ContainerStopped(container *Container) {
	container.status = Stopped
	log.Debugf("Container %s marked as stopped.\n", container.name)
}

func (model *Model) ContainerDestroyed(container *Container) {
	container.status = Destroyed
	log.Debugf("Container %s marked as destroyed.\n", container.name)
}
