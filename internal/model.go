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
	return task.status == New
}

func (model *Model) AllDependenciesReady(task *Task) bool {
	for _, dependency := range task.config.Dependencies {
		if dependency == task.name {
			log.Printf("WARNING %s contains dependency on itself.\n", task.name)
		}

		if model.tasks[dependency].status != Ready {
			log.Printf("Dependency %s->%s not ready.\n", task.name, dependency)
			return false
		}
	}

	return true
}

func (model *Model) AllServiceDependenciesReadyAndAllTaskDependenciesStopped(task *Task) bool {
	for _, dependency := range task.config.Dependencies {
		if model.tasks[dependency].taskType == "service" && model.tasks[dependency].status != Ready {
			return false
		}
		if model.tasks[dependency].taskType == "task" && model.tasks[dependency].status != Stopped {
			return false
		}
	}

	return true
}

func (model *Model) getCreatableTasks(taskType string) []*Task {
	createableTasks := make([]*Task, 0)
	for _, task := range model.tasks {
		if task.status == New && task.taskType == taskType && model.CanCreateTask(task) {
			createableTasks = append(createableTasks, task)
		}
	}

	return createableTasks
}

func (model *Model) getStartableServices() []*Task {
	startableTasks := make([]*Task, 0)
	for _, task := range model.tasks {
		if task.status == Created && task.taskType == "service" && model.CanStartService(task) {
			startableTasks = append(startableTasks, task)
		}
	}

	return startableTasks
}

func (model *Model) getStartableTasks() []*Task {
	startableTasks := make([]*Task, 0)
	for _, task := range model.tasks {
		if task.status == Created && task.taskType == "task" && model.CanStartTask(task) {
			startableTasks = append(startableTasks, task)
		}
	}

	return startableTasks
}

func (model *Model) CanStartService(task *Task) bool {
	return task.status == Created && model.AllDependenciesReady(task)
}

func (model *Model) CanStartTask(task *Task) bool {
	return task.status == Created && model.AllServiceDependenciesReadyAndAllTaskDependenciesStopped(task)
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

func (model *Model) TaskStopped(task *Task) {
	task.status = Stopped
	log.Printf("Task %s marked as stopped.\n", task.name)
}

func (model *Model) TaskDestroyed(task *Task) {
	task.status = Destroyed
	log.Printf("Task %s marked as destroyed.\n", task.name)
}
