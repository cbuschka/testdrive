package internal

type Phase interface {
	postHandle(event *Event) (Phase, error)
	isDone() bool
	String() string
}
