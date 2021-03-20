package processing

type Phase interface {
	postHandle() (Phase, error)
	isDone() bool
	String() string
}
