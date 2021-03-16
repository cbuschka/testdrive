package internal

type Resource interface {
	Name() string
	Type() string
}
