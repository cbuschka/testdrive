package internal

import (
	"context"
	"log"
	"time"
)

func tick(eventQueue chan Event) {
	for true {
		eventQueue <- &TickEvent{}
		time.Sleep(100 * time.Millisecond)
	}
}

func (session *Session) Run() (int, error) {

	config := session.config
	model := session.model
	for taskName, taskConfig := range config.Services {
		err := model.AddTask(&Task{name: taskName, dependencies: make([]string, 0), status: New, config: &taskConfig})
		if err != nil {
			return -1, err
		}
	}

	go tick(session.eventQueue)

	go session.containerRuntime.AddEventListener(context.TODO(), func(event ContainerEvent) {
		session.eventQueue <- &event
	})

	for event := range session.eventQueue {
		if event == nil {
			break
		}

		err := session.createContainersForCreatableTasks(model)
		if err != nil {
			return -1, err
		}

		err = session.startContainersForStartableTasks(model)
		if err != nil {
			return -1, err
		}

		if event.Type() == "container.create" {
			task := model.GetTaskByContainerId(event.Id())
			model.TaskCreated(task)
		} else if event.Type() == "container.start" {
			task := model.GetTaskByContainerId(event.Id())
			if task.config.Healthcheck == nil {
				model.TaskReady(task)
			} else {
				model.TaskStarted(task)
			}
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

	return 0, nil
}

func (session *Session) createContainersForCreatableTasks(model *Model) error {
	creatableTasks := model.getCreatableTasks()
	for _, creatableTask := range creatableTasks {
		log.Printf("Found creatable task %s.", creatableTask.name)
		err := session.createTask(creatableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) startContainersForStartableTasks(model *Model) error {
	startableTasks := model.getStartableTasks()
	for _, startableTask := range startableTasks {
		log.Printf("Found startable task %s.", startableTask.name)
		err := session.startTask(startableTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createTask(task *Task) error {

	session.model.TaskCreating(task)
	id, err := session.containerRuntime.CreateContainer(task.name, task.config.Image)
	if err != nil {
		session.model.TaskFailed(task)
		return err
	}

	log.Printf("Created container %s for task %s.", id, task.name)

	task.containerId = id

	return nil
}

func (session *Session) startTask(task *Task) error {
	session.model.TaskStarting(task)
	err := session.containerRuntime.StartContainer(task.containerId)
	if err != nil {
		session.model.TaskFailed(task)
		return err
	}

	log.Printf("Started container task %s.", task.name)

	return nil
}
