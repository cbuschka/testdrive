package processing

import (
	"context"
	"fmt"
	configPkg "github.com/cbuschka/testdrive/internal/config"
	"github.com/cbuschka/testdrive/internal/container_runtime"
	"github.com/cbuschka/testdrive/internal/dialog"
	dockerPkg "github.com/cbuschka/testdrive/internal/docker"
	"github.com/cbuschka/testdrive/internal/log"
	"github.com/cbuschka/testdrive/internal/model"
	"github.com/sheerun/queue"
	"os"
	"os/signal"
	"time"
)

type Session struct {
	id                    string
	config                *configPkg.TestdriveConfig
	model                 *model.Model
	eventQueue            *queue.Queue
	phase                 Phase
	ctx                   context.Context
	containerRuntime      container_runtime.ContainerRuntime
	resyncInterval        time.Duration
	containerEventTimeout time.Duration
	containerStopTimeout  time.Duration
	interruptCount        int
}

func NewSession() (*Session, error) {
	model := model.NewModel()
	docker, err := dockerPkg.NewDocker()
	if err != nil {
		return nil, err
	}
	session := Session{id: "1", config: nil, ctx: context.Background(),
		model: model, eventQueue: queue.New(),
		phase: nil, containerRuntime: container_runtime.ContainerRuntime(docker),
		resyncInterval:        1 * time.Second,
		containerEventTimeout: 1 * time.Second,
		containerStopTimeout:  5 * time.Second,
		interruptCount:        0}
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

	config, err := configPkg.LoadConfig(reader)
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

	go session.containerRuntime.AddEventListener(session.ctx, func(event container_runtime.ContainerEvent) {

		if event.Id() != "" {
			container := session.model.GetContainerByContainerId(event.Id())
			if container != nil {
				message := fmt.Sprintf("%s: %s", container.Name, event.Message())
				event = *container_runtime.NewContainerEvent(event.Type(), event.Id(), message)
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
		container := session.model.GetContainerByContainerId(event.Id())
		if container != nil {
			go session.containerRuntime.ReadContainerLogs(event.Id(), session.ctx, func(line string) {
				session.eventQueue.Append(&LogEvent{containerName: container.Name, line: line})
			})
			if container.Config.Healthcheck == nil {
				session.model.ContainerReady(container)
			} else {
				session.model.ContainerStarted(container)
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
				log.Warningf("Container %s has failed with exit code %d. Shutting down...", container.Name, exitCode)
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
		session.interruptCount++
		if session.interruptCount == 1 {
			log.Debugf("Shutting down because of user interrupt...")
			session.phase = NewShutdownPhase(session)
		} else if session.interruptCount == 2 {
			log.Debugf("Cleaning up because of multiple user interrupts...")
			session.phase = NewCleanupPhase(session)
		} else {
			log.Info("Aborted.")
			os.Exit(1)
		}
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

func (session *Session) addTaskContainersFrom(config *configPkg.TestdriveConfig) error {
	for taskName, taskConfig := range config.Tasks {
		err := session.model.AddContainer(&model.Container{Name: taskName, ContainerType: model.Task, Status: model.New, Config: taskConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) addServiceContainersFrom(config *configPkg.TestdriveConfig) error {
	for serviceName, serviceConfig := range config.Services {
		err := session.model.AddContainer(&model.Container{Name: serviceName, ContainerType: model.Service, Status: model.New, Config: serviceConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainersForCreatableContainers(containerType model.ContainerType) error {
	creatableContainers := session.model.GetCreatableContainers(containerType)
	for _, creatableContainer := range creatableContainers {
		log.Debugf("Found creatable container %s.", creatableContainer.Name)
		err := session.createContainer(creatableContainer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableServices() error {
	startableTasks := session.model.GetStartableServiceContainers()
	for _, startableTask := range startableTasks {
		log.Debugf("Found startable container %s.", startableTask.Name)
		err := session.startContainer(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableTasks() error {
	startableTasks := session.model.GetStartableTaskContainers()
	for _, startableTask := range startableTasks {
		log.Debugf("Found startable task %s.", startableTask.Name)
		err := session.startContainer(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainer(container *model.Container) error {

	log.Debugf("Creating container for %s...\n", container.Name)

	session.model.ContainerCreating(container)
	id, err := session.containerRuntime.CreateContainer(container.Name, container.Config.Image, container.Config.Command)
	if err != nil {
		session.model.ContainerFailed(container)
		return err
	}

	log.Debugf("Started container creation for %s (%s).", container.Name, id)
	container.ContainerId = id

	return nil
}

func (session *Session) startContainer(container *model.Container) error {

	log.Debugf("Starting container %s (%s).", container.Name, container.ContainerId)

	session.model.ContainerStarting(container)
	err := session.containerRuntime.StartContainer(container.ContainerId)
	if err != nil {
		session.model.ContainerFailed(container)
		return err
	}

	return nil
}

func (session *Session) allContainersReady(containerType model.ContainerType) bool {
	for _, container := range session.model.Containers {
		if container.ContainerType == containerType && container.Status != model.Ready {
			return false
		}
	}
	return true
}

func (session *Session) stopRunningContainers() error {
	for _, container := range session.model.Containers {
		if container.Status == model.Ready || container.Status == model.Started {
			log.Debugf("Stopping container %s (%s)...\n", container.Name, container.ContainerId)

			container.Status = model.Stopping
			container.StopStartedAt = time.Now()
			err := session.containerRuntime.StopContainer(container.ContainerId, session.containerStopTimeout)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) destroyNonRunningContainers() error {
	for _, container := range session.model.Containers {
		if container.Status != model.Destroyed && container.Status != model.Creating && container.Status != model.Starting && container.Status != model.Destroying && container.Status != model.Stopping {
			log.Debugf("Destroying container %s (%s)...\n", container.Name, container.ContainerId)

			container.Status = model.Destroying
			container.DestroyStartedAt = time.Now()
			err := session.containerRuntime.DestroyContainer(container.ContainerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) destroyAllNonDestroyedContainers() error {
	for _, container := range session.model.Containers {
		if container.Status != model.Destroyed && container.Status != model.New {
			log.Debugf("Destroying container %s (%s)...\n", container.Name, container.ContainerId)

			container.Status = model.Destroying
			container.DestroyStartedAt = time.Now()
			err := session.containerRuntime.DestroyContainer(container.ContainerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) allContainersDestroyed() bool {
	for _, container := range session.model.Containers {
		if container.Status != model.Destroyed && container.Status != model.New {
			log.Debugf("Container %s (%s) still not destroyed (%s).", container.Name, container.ContainerId, container.Status)
			return false
		}
	}

	return true
}

func (session *Session) allTaskContainersStopped() bool {
	for _, container := range session.model.Containers {
		if container.ContainerType == model.Task && (container.Status != model.Stopped && container.Status != model.Destroyed) {
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

	for _, container := range session.model.Containers {

		realState := stateByContainerId[container.ContainerId]
		if realState == "running" && (container.Status == model.Ready || container.Status == model.Started) {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.Name, container.ContainerId, container.Status)
		} else if realState == "" && container.Status == model.Destroyed {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.Name, container.ContainerId, container.Status)
		} else if realState == "" && (container.Status == model.New || container.Status == model.Creating) {
			log.Debugf("State sync: Container %s (%s) is %s - as expected, good.", container.Name, container.ContainerId, container.Status)
		} else if realState == "created" && container.Status == model.Creating && container.CreateStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as created.", container.Name, container.ContainerId, container.Status, realState)
			container.Status = model.Created
		} else if realState == "running" && (container.Status == model.New || container.Status == model.Creating || container.Status == model.Created || container.Status == model.Starting) && container.StartStartededAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: container %s (%s) running but in state %s. Marking as ready.", container.Name, container.ContainerId, container.Status)
			if container.Config.Healthcheck == nil {
				container.Status = model.Ready
			} else {
				container.Status = model.Started
			}
		} else if realState == "" && container.Status == model.Stopping && container.StopStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as destroyed.", container.Name, container.ContainerId, container.Status, realState)
			container.Status = model.Destroyed
		} else if realState == "" && container.Status == model.Destroying && container.DestroyStartedAt.Add(session.containerEventTimeout).Before(time.Now()) {
			log.Debugf("State sync: Container %s (%s) is %s - but container runtime shows %s for a long time. Marking as destroyed.", container.Name, container.ContainerId, container.Status, realState)
			container.Status = model.Destroyed
		} else if realState == "" && (container.Status == model.Ready || container.Status == model.Started || container.Status == model.Starting || container.Status == model.Stopped) {
			log.Debugf("State sync: container %s (%s) in state %s lost. Marking as destroyed.", container.Name, container.ContainerId, container.Status)
			container.Status = model.Destroyed
		} else {
			log.Debugf("State sync: Container %s (%s) is %s, but container runtime shows %s.", container.Name, container.ContainerId, container.Status, realState)
		}
	}

	return nil
}

func (session *Session) allContainersNonRunning() bool {
	for _, container := range session.model.Containers {
		if container.Status != model.Stopped && container.Status != model.Stopping && container.Status != model.Destroying && container.Status != model.Destroyed && container.Status != model.New && container.Status != model.Creating && container.Status != model.Created {
			return false
		}
	}

	return true
}
