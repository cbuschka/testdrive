package model

type ContainerType struct {
	name string
}

func (t *ContainerType) String() string {
	return t.name
}

var (
	Service = ContainerType{"service"}
	Task    = ContainerType{"task"}
)
