package internal

import (
	"context"
	"os"
	"os/signal"
	"time"
)

type Session struct {
	id               string
	config           *Config
	model            *Model
	eventQueue       chan Event
	phase            Phase
	ctx              context.Context
	containerRuntime ContainerRuntime
}

func NewSession() (*Session, error) {
	model := NewModel()
	docker, err := NewDocker()
	if err != nil {
		return nil, err
	}
	session := Session{id: "1", config: nil, ctx: context.Background(),
		model: model, eventQueue: make(chan Event, 100),
		phase: nil, containerRuntime: ContainerRuntime(docker)}
	return &session, nil
}

func (session *Session) LoadConfig(file string) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	config, err := LoadConfig(reader)
	if err != nil {
		return err
	}

	session.config = config

	return nil
}

func tick(eventQueue chan Event) {
	for true {
		eventQueue <- &TickEvent{}
		time.Sleep(100 * time.Millisecond)
	}
}

func handleSigint(signalChannel chan os.Signal, eventChannel chan Event) {
	for range signalChannel {
		log.Debugf("SIGINT received.\n")
		eventChannel <- nil
	}
}

func (session *Session) Run() (int, error) {

	config := session.config
	err := session.addServiceContainersFrom(config)
	if err != nil {
		return -1, err
	}

	err = session.addTaskContainersFrom(config)
	if err != nil {
		return -1, err
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go handleSigint(signalChannel, session.eventQueue)

	go tick(session.eventQueue)

	go session.containerRuntime.AddEventListener(context.TODO(), func(event ContainerEvent) {
		session.eventQueue <- &event
	})

	session.phase = Phase(&StartupPhase{session: session})

	for event := range session.eventQueue {
		log.Debugf("Phase is %s, event is %v.\n", session.phase.String(), event)

		if event == nil {
			session.phase = Phase(&ShutdownPhase{session: session})
		}

		if session.phase.isDone() {
			break
		}

		if event != nil {
			session.handleEvent(event)
		}

		session.phase, err = session.phase.postHandle()
		if err != nil {
			return -1, err
		}
	}

	session.ctx.Done()

	return 0, nil
}

func (session *Session) handleEvent(event Event) {
	if event.Type() == "container.create" {
		task := session.model.GetContainerByContainerId(event.Id())
		if task != nil {
			session.model.ContainerCreated(task)
		}
	} else if event.Type() == "container.start" {
		task := session.model.GetContainerByContainerId(event.Id())
		if task != nil {
			go session.containerRuntime.ReadContainerLogs(event.Id(), session.ctx, func(line string) {
				session.eventQueue <- &LogEvent{containerName: task.name, line: line}
			})
			if task.config.Healthcheck == nil {
				session.model.ContainerReady(task)
			} else {
				session.model.TaskStarted(task)
			}
		}
	} else if event.Type() == "container.die" {
		task := session.model.GetContainerByContainerId(event.Id())
		session.model.ContainerStopped(task)
	} else if event.Type() == "container.kill" {
		task := session.model.GetContainerByContainerId(event.Id())
		session.model.ContainerDestroyed(task)
	} else if event.Type() == "image.pull" {
		log.Debugf("Event seen: %s\n", event.Type())
	} else if event.Type() == "network.connect" {
		log.Debugf("Event seen: %s\n", event.Type())
	} else if event.Type() == "log" {
		dialog.ContainerOutput(&event.(*LogEvent).containerName, &event.(*LogEvent).line)
	} else if event.Type() == "tick" {
		// ignored
	} else {
		log.Debugf("Unhandled event: %s\n", event.Type())
	}
}

func (session *Session) addTaskContainersFrom(config *Config) error {
	for taskName, taskConfig := range config.Tasks {
		err := session.model.AddContainer(&Container{name: taskName, containerType: ContainerType_Task, status: New, config: taskConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) addServiceContainersFrom(config *Config) error {
	for serviceName, serviceConfig := range config.Services {
		err := session.model.AddContainer(&Container{name: serviceName, containerType: ContainerType_Service, status: New, config: serviceConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainersForCreatableContainers(containerType string) error {
	creatableContainers := session.model.getCreatableContainers(containerType)
	for _, creatableContainer := range creatableContainers {
		log.Debugf("Found creatable task %s.", creatableContainer.name)
		err := session.createContainer(creatableContainer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableServices() error {
	startableTasks := session.model.getStartableServiceContainers()
	for _, startableTask := range startableTasks {
		log.Debugf("Found startable task %s.", startableTask.name)
		err := session.startContainer(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableTasks() error {
	startableTasks := session.model.getStartableTaskContainers()
	for _, startableTask := range startableTasks {
		log.Debugf("Found startable task %s.", startableTask.name)
		err := session.startContainer(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainer(task *Container) error {

	log.Debugf("Creating container for %s...\n", task.name)

	session.model.ContainerCreating(task)
	id, err := session.containerRuntime.CreateContainer(task.name, task.config.Image, task.config.Command)
	if err != nil {
		session.model.ContainerFailed(task)
		return err
	}

	log.Debugf("Created container %s for task %s.", id, task.name)

	task.containerId = id

	return nil
}

func (session *Session) startContainer(task *Container) error {

	log.Debugf("Starting container task %s.", task.name)

	session.model.ContainerStarting(task)
	err := session.containerRuntime.StartContainer(task.containerId)
	if err != nil {
		session.model.ContainerFailed(task)
		return err
	}

	return nil
}

func (session *Session) allContainersReady(taskType string) bool {
	for _, task := range session.model.containers {
		if task.containerType == taskType && task.status != Ready {
			return false
		}
	}
	return true
}

func (session *Session) stopRunningContainers() error {
	for _, container := range session.model.containers {
		if container.status == Ready {
			log.Debugf("Stopping container for %s...\n", container.name)

			container.status = Stopping
			err := session.containerRuntime.StopContainer(container.containerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) destroyStoppedContainers() error {
	for _, container := range session.model.containers {
		if container.status == Stopped {
			log.Debugf("Destroying container for %s...\n", container.name)

			container.status = Destroying
			err := session.containerRuntime.DestroyContainer(container.containerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) allContainersDestroyed() bool {
	for _, container := range session.model.containers {
		if container.status != Destroyed {
			return false
		}
	}

	return true
}

func (session *Session) allContainersStopped(containerType string) bool {
	for _, container := range session.model.containers {
		if container.containerType == containerType && container.status != Stopped {
			return false
		}
	}

	return true
}
