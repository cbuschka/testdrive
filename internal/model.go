package internal

import "log"

type Model struct {
	tasks map[string]*Task
}

func NewModel() *Model {
	return &Model{tasks: make(map[string]*Task, 0)}
}

func (model *Model) GetTaskByContainerId(containerId string) *Task {
	for _, task := range model.tasks {
		if task.containerId == containerId {
			return task
		}
	}

	return nil
}

func (model *Model) AddTask(task *Task) error {
	model.tasks[task.name] = task
	return nil
}

func (model *Model) CanCreateTask(task *Task) bool {
	return task.status == New && model.AllDependenciesReady(task)
}

func (model *Model) AllDependenciesReady(task *Task) bool {
	for _, dependency := range task.dependencies {
		if model.tasks[dependency].status != Ready {
			return false
		}
	}

	return true
}

func (model *Model) getCreatableTasks() []*Task {
	createableTasks := make([]*Task, 0)
	for _, task := range model.tasks {
		if task.status == New && model.CanCreateTask(task) {
			createableTasks = append(createableTasks, task)
		}
	}

	return createableTasks
}

func (model *Model) getStartableTasks() []*Task {
	startableTasks := make([]*Task, 0)
	for _, task := range model.tasks {
		if task.status == Created && model.CanStartTask(task) {
			startableTasks = append(startableTasks, task)
		}
	}

	return startableTasks
}

func (model *Model) CanStartTask(task *Task) bool {
	return task.status == Created && len(task.dependencies) == 0
}

func (model *Model) TaskCreating(task *Task) {
	task.status = Creating
	log.Printf("Task %s marked as creating.\n", task.name)
}

func (model *Model) TaskStarted(task *Task) {
	task.status = Started
	log.Printf("Task %s marked as started.\n", task.name)
}

func (model *Model) TaskReady(task *Task) {
	task.status = Ready
	log.Printf("Task %s marked as ready.\n", task.name)
}

func (model *Model) TaskCreated(task *Task) {
	task.status = Created
	log.Printf("Task %s marked as created.\n", task.name)
}

func (model *Model) TaskStarting(task *Task) {
	task.status = Starting
	log.Printf("Task %s marked as starting.\n", task.name)
}

func (model *Model) TaskFailed(task *Task) {
	task.status = Failed
	log.Printf("Task %s marked as failed.\n", task.name)
}
