package model

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
	Name             string
	Status           ContainerStatus
	CreateStartedAt  time.Time
	StartStartededAt time.Time
	StopStartedAt    time.Time
	DestroyStartedAt time.Time
	FailedAt         time.Time
	Config           *config.ContainerConfig
	ContainerId      string
	ContainerType    ContainerType
}
