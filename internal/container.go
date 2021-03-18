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

type Container struct {
	name          string
	status        TaskStatus
	config        *ContainerConfig
	containerId   string
	containerType string
}

const (
	ContainerType_Service = "service"
	ContainerType_Task    = "task"
)
