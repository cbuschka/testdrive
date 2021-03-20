package internal

import (
	"github.com/cbuschka/testdrive/internal/config"
	"time"
)

type ContainerStatus struct {
	name string
}

func (status ContainerStatus) String() string {
	return status.name
}

var (
	New        = ContainerStatus{name: "New"}
	Creating   = ContainerStatus{name: "Creating"}
	Created    = ContainerStatus{name: "Created"}
	Starting   = ContainerStatus{name: "Starting"}
	Started    = ContainerStatus{name: "Started"}
	Ready      = ContainerStatus{name: "Ready"}
	Stopping   = ContainerStatus{name: "Stopping"}
	Stopped    = ContainerStatus{name: "Stopped"}
	Destroying = ContainerStatus{name: "Destroying"}
	Destroyed  = ContainerStatus{name: "Destroyed"}
	Failed     = ContainerStatus{name: "Failed"}
)

type Container struct {
	name             string
	status           ContainerStatus
	createStartedAt  time.Time
	startStartededAt time.Time
	stoppStartedAt   time.Time
	destroyStartedAt time.Time
	failedAt         time.Time
	config           *config.ContainerConfig
	containerId      string
	containerType    string
}

const (
	ContainerType_Service = "service"
	ContainerType_Task    = "task"
)
