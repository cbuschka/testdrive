package internal

import (
	"context"
	"log"
	"os"
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

func (session *Session) Run() (int, error) {

	config := session.config
	err := session.addServiceTasksFrom(config)
	if err != nil {
		return -1, err
	}

	err = session.assTaskTasksFrom(config)
	if err != nil {
		return -1, err
	}

	go tick(session.eventQueue)

	go session.containerRuntime.AddEventListener(context.TODO(), func(event ContainerEvent) {
		session.eventQueue <- &event
	})

	session.phase = Phase(&StartupPhase{session: session})

	for event := range session.eventQueue {
		log.Printf("Phase is %s, event is %v.\n", session.phase.String(), event)

		if event == nil {
			break
		}

		if session.phase.isDone() {
			break
		}

		session.handleEvent(event)

		session.phase, err = session.phase.postHandle(&event)
		if err != nil {
			return -1, err
		}
	}

	session.ctx.Done()

	return 0, nil
}

func (session *Session) handleEvent(event Event) {
	if event.Type() == "container.create" {
		task := session.model.GetTaskByContainerId(event.Id())
		if task != nil {
			session.model.TaskCreated(task)
		}
	} else if event.Type() == "container.start" {
		task := session.model.GetTaskByContainerId(event.Id())
		if task != nil && task.config.Healthcheck == nil {
			session.model.TaskReady(task)
		} else {
			session.model.TaskStarted(task)
		}
	} else if event.Type() == "container.die" {
		task := session.model.GetTaskByContainerId(event.Id())
		session.model.TaskStopped(task)
	} else if event.Type() == "container.kill" {
		task := session.model.GetTaskByContainerId(event.Id())
		session.model.TaskDestroyed(task)
	} else if event.Type() == "image.pull" {
		log.Printf("Event seen: %s\n", event.Type())
	} else if event.Type() == "network.connect" {
		log.Printf("Event seen: %s\n", event.Type())
	} else if event.Type() == "tick" {
		// ignored
	} else {
		log.Printf("Unhandled event: %s\n", event.Type())
	}
}

func (session *Session) assTaskTasksFrom(config *Config) error {
	for taskName, taskConfig := range config.Tasks {
		err := session.model.AddTask(&Task{name: taskName, taskType: "task", dependencies: make([]string, 0), status: New, config: &taskConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) addServiceTasksFrom(config *Config) error {
	for taskName, taskConfig := range config.Services {
		err := session.model.AddTask(&Task{name: taskName, taskType: "service", dependencies: make([]string, 0), status: New, config: &taskConfig})
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainersForCreatableTasks(taskType string) error {
	creatableTasks := session.model.getCreatableTasks(taskType)
	for _, creatableTask := range creatableTasks {
		log.Printf("Found creatable task %s.", creatableTask.name)
		err := session.createContainer(creatableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableTasks(taskType string) error {
	startableTasks := session.model.getStartableTasks(taskType)
	for _, startableTask := range startableTasks {
		log.Printf("Found startable task %s.", startableTask.name)
		err := session.startContainer(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createContainer(task *Task) error {

	log.Printf("Creating container for %s...\n", task.name)

	session.model.TaskCreating(task)
	id, err := session.containerRuntime.CreateContainer(task.name, task.config.Image, task.config.Command)
	if err != nil {
		session.model.TaskFailed(task)
		return err
	}

	log.Printf("Created container %s for task %s.", id, task.name)

	task.containerId = id

	return nil
}

func (session *Session) startContainer(task *Task) error {

	log.Printf("Starting container task %s.", task.name)

	session.model.TaskStarting(task)
	err := session.containerRuntime.StartContainer(task.containerId)
	if err != nil {
		session.model.TaskFailed(task)
		return err
	}

	return nil
}

func (session *Session) allContainersReady(taskType string) bool {
	for _, task := range session.model.tasks {
		if task.taskType == taskType && task.status != Ready {
			return false
		}
	}
	return true
}

func (session *Session) stopRunningContainers() error {
	for _, task := range session.model.tasks {
		if task.status == Ready {
			log.Printf("Stopping container for %s...\n", task.name)

			task.status = Stopping
			err := session.containerRuntime.StopContainer(task.containerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) destroyStoppedContainers() error {
	for _, task := range session.model.tasks {
		if task.status == Stopped {
			log.Printf("Destroying container for %s...\n", task.name)

			task.status = Destroying
			err := session.containerRuntime.DestroyContainer(task.containerId)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (session *Session) allContainersDestroyed() bool {
	for _, task := range session.model.tasks {
		if task.status != Destroyed {
			return false
		}
	}

	return true
}

func (session *Session) allContainersStopped(taskType string) bool {
	for _, task := range session.model.tasks {
		if task.taskType == taskType && task.status != Stopped {
			return false
		}
	}

	return true
}
