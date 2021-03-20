package processing

type Event interface {
	Type() string
	String() string
	Id() string
}
