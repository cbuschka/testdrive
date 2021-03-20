package processing

type Resource interface {
	Name() string
	Type() string
}
