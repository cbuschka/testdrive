package internal

type Phase interface {
	postHandle() (Phase, error)
	isDone() bool
	String() string
}