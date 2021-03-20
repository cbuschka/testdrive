package internal

import (
	"context"
	"fmt"
	"github.com/cbuschka/testdrive/internal/config"
	"github.com/cbuschka/testdrive/internal/log"
	"github.com/sheerun/queue"
	"os"
	"os/signal"
	"time"
)

type Session struct {
	id                    string
	config                *config.TestdriveConfig
	model                 *Model
	eventQueue            *queue.Queue
	phase                 Phase
	ctx                   context.Context
	containerRuntime      ContainerRuntime
	resyncInterval        time.Duration
	containerEventTimeout time.Duration
	containerStopTimeout  time.Duration
}

func NewSession() (*Session, error) {
	model := NewModel()
	docker, err := NewDocker()
	if err != nil {
		return nil, err
	}
	session := Session{id: "1", config: nil, ctx: context.Background(),
		model: model, eventQueue: queue.New(),
		phase: nil, containerRuntime: ContainerRuntime(docker),
		resyncInterval:        1 * time.Second,
		containerEventTimeout: 1 * time.Second,
		containerStopTimeout:  5 * time.Second}
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

	config, err := config.LoadConfig(reader)
	if err != nil {
		return err
	}

	session.config = config

	return nil
}

func (session *Session) tick(eventQueue *queue.Queue) {
	for true {
		eventQueue.Append(&TickEvent{})
		time.Sleep(session.resyncInterval)
	}
}

func handleSigint(signalChannel chan os.Signal, eventChannel *queue.Queue) {
	for range signalChannel {
		log.Info("SIGINT received.")
		eventChannel.Append(&SigintEvent{})
	}
}

func (session *Session) Run() error {

	config := session.config
	err := session.addServiceContainersFrom(config)
	if err != nil {
		return err
	}

	err = session.addTaskContainersFrom(config)
	if err != nil {
		return err
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go handleSigint(signalChannel, session.eventQueue)

	go session.tick(session.eventQueue)

	go session.containerRuntime.AddEventListener(session.ctx, func(event ContainerEvent) {

		id := event.Id()
		if id != "" {
			container := session.model.GetContainerByContainerId(id)
			if container != nil {
				message := fmt.Sprintf("%s: %s", container.name, event.message)
				event = ContainerEvent{eventType: event.eventType, id: event.id, message: message}
			}
		}

		session.eventQueue.Append(&event)
	})

	session.phase = Phase(&StartupPhase{session: session})

	for {
		event := session.eventQueue.Pop().(Event)
		log.Debugf("Phase is %s, event is %v.", session.phase.String(), event)

		if event == nil {
			session.phase = Phase(&ShutdownPhase{session: session})
		}

		if session.phase.isDone() {
			break
		}

		if event != nil {
			err = session.handleEvent(event)
			if err != nil {
				return err
			}
		}

		session.phase, err = session.phase.postHandle()
		if err != nil {
			return err
		}
	}

	session.ctx.Done()

	log.Info("Finished.")

	return nil
}

func (session *Session) handleEvent(event Event) error {
	if event.Type() == "container.create" {
		task := session.model.GetContainerByContainerId(event.Id())
		if task != nil {
			session.model.ContainerCreated(task)
		} else {
			log.Warningf("Saw container created event for unknown container %s.", event.Id())
		}
	} else if event.Type() == "container.start" {
		task := session.model.GetContainerByContainerId(event.Id())
		if task != nil {
			go session.containerRuntime.ReadContainerLogs(event.Id(), session.ctx, func(line string) {
				session.eventQueue.Append(&LogEvent{containerName: task.name, line: line})
			})
			if task.config.Healthcheck == nil {
				session.model.ContainerReady(task)
			} else {
				session.model.ContainerStarted(task)
			}
		} else {
			log.Warningf("Saw container started event for unknown container %s.", event.Id())
		}
	} else if event.Type() == "container.stop" || event.Type() == "container.kill" {
		// inored
	} else if event.Type() == "container.die" {
		container := session.model.GetContainerByContainerId(event.Id())
		if container != nil {
			exitCode, err := session.containerRuntime.GetContainerExitCode(event.Id())
			if err != nil {
				return err
			}

			if exitCode == 0 {
				session.model.ContainerStopped(container)
			} else {
				session.model.ContainerFailed(container)
				log.Warningf("Container %s has failed with exit code %d. Shutting down...", container.name, exitCode)
				session.phase = &ShutdownPhase{session: session}
			}
		} else {
			log.Warningf("Saw container die/stop/kill event for unknown container %s.", event.Id())
		}
	} else if event.Type() == "container.destroy" {
		container := session.model.GetContainerByContainerId(event.Id())
		if container != nil {
			session.model.ContainerDestroyed(container)
		} else {
			log.Warningf("Saw container destroyed event for unknown container %s.", event.Id())
		}
	} else if event.Type() == "image.pull" {
		log.Debugf("Event seen: %s\n", event.Type())
	} else if event.Type() == "network.connect" {
		log.Debugf("Event seen: %s\n", event.Type())
	} else if event.Type() == "log" {
		dialog.ContainerOutput(&event.(*LogEvent).containerName, &event.(*LogEvent).line)
	} else if event.Type() == "sigint" {
		log.Debugf("Sigint seen, shutting down...")
		session.phase = &ShutdownPhase{session: session}
	} else if event.Type() == "tick" {
		err := session.resyncContainerStates()
		if err != nil {
			return err
		}
	} else {
		log.Debugf("Unhandled event: %s\n", event.Type())
	}

	return nil
}

func (session *Session) addTaskContainersFrom(config *config.TestdriveConfig) error {
	for taskName, taskConfig := range config.Tasks {
		err := session.model.AddContainer(&Container{name: taskName, containerType: ContainerType_Task, status: New, config: taskConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) addServiceContainersFrom(config *config.TestdriveConfig) error {
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
		log.Debugf("Found creatable container %s.", creatableContainer.name)
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
		log.Debugf("Found startable container %s.", startableTask.name)
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

func (session *Session) createContainer(container *Container) error {

	log.Debugf("Creating container for %s...\n", container.name)

	session.model.ContainerCreating(container)
	id, err := session.containerRuntime.CreateContainer(container.name, container.config.Image, container.config.Command)
	if err != nil {
		session.model.ContainerFailed(container)
		return err
	}

	log.Debugf("Started container creation for %s (%s).", container.name, id)
	container.containerId = id

	return nil
}

func (session *Session) startContainer(container *Container) error {

	log.Debugf("Starting container %s (%s).", container.name, container.containerId)

	session.model.ContainerStarting(container)
	err := session.containerRuntime.StartContainer(container.containerId)
	if err != nil {
		session.model.ContainerFailed(container)
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
		if container.status == Ready || container.status == Started {
			log.Debugf("Stopping container %s (%s)...\n", container.name, container.containerId)

			container.status = Stopping
			container.stoppStartedAt = time.Now()
			err := session.containerRuntime.StopContainer(container.containerId, session.containerStopTimeout)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) destroyNonRunningContainers() error {
	for _, container := range session.model.containers {
		if container.status != Destroyed && container.status != Creating && container.status != Starting && container.status != Destroying && container.status != Stopping {
			log.Debugf("Destroying container %s (%s)...\n", container.name, container.containerId)

			container.status = Destroying
			container.destroyStartedAt = time.Now()
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
		if container.status != Destroyed && container.status != New {
			log.Debugf("Container %s (%s) still not destroyed (%s).", container.name, container.containerId, container.status)
			return false
		}
	}

	return true
}

func (session *Session) allTaskContainersStopped() bool {
	for _, container := range session.model.containers {
		if container.containerType == "task" && (container.status != Stopped && container.status != Destroyed) {
			return false
		}
	}

	return true
}

func (session *Session) resyncContainerStates() error {

	log.Debugf("Resynchronizing container states...")

	stateByContainerId, err := session.containerRuntime.ListContainers()
	if err != nil {
		return err
	}

	for _, container := range session.model.containers {

		realState := stateByContainerId[container.containerId]
		if realState == "running" && (container.status == Ready || container.status == Started) {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.name, container.containerId, container.status)
		} else if realState == "" && container.status == Destroyed {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.name, container.containerId, container.status)
		} else if realState == "" && (container.status == New || container.status == Creating) {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.name, container.containerId, container.status)
		} else if realState == "created" && container.status == Creating && container.createStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as created.", container.name, container.containerId, container.status, realState)
			container.status = Created
		} else if realState == "running" && (container.status == New || container.status == Creating || container.status == Created || container.status == Starting) && container.startStartededAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: container %s (%s) running but in state %s. Marking as ready.", container.name, container.containerId, container.status)
			if container.config.Healthcheck == nil {
				container.status = Ready
			} else {
				container.status = Started
			}
		} else if realState == "" && container.status == Stopping && container.stoppStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as destroyed.", container.name, container.containerId, container.status, realState)
			container.status = Destroyed
		} else if realState == "" && container.status == Destroying && container.destroyStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as destroyed.", container.name, container.containerId, container.status, realState)
			container.status = Destroyed
		} else if realState == "" && (container.status == Ready || container.status == Started || container.status == Starting || container.status == Stopped) {
			log.Debugf("State sync: container %s (%s) in state %s lost. Marking as destroyed.", container.name, container.containerId, container.status)
			container.status = Destroyed
		} else {
			log.Debugf("State sync: Container %s (%s) is %s, but container runtime shows %s.", container.name, container.containerId, container.status, realState)
		}
	}

	return nil
}
