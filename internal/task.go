package internal

type TaskStatus string

const (
	New        = "New"
	Creating   = "Creating"
	Created    = "Created"
	Starting   = "Starting"
	Started    = "Started"
	Ready      = "Ready"
	Stopping   = "Stopping"
	Stopped    = "Stopped"
	Destroying = "Destroying"
	Destroyed  = "Destroyed"
	Failed     = "Failed"
)

type Task struct {
	name         string
	status       TaskStatus
	config       *TaskConfig
	containerId  string
	taskType     string
}
