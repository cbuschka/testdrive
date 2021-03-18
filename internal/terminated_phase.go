package internal

type TerminatedPhase struct {
	session *Session
}

func (phase *TerminatedPhase) String() string {
	return "PHASE_TERMINATED"
}

func (phase *TerminatedPhase) postHandle() (Phase, error) {

	return Phase(phase), nil
}

func (phase *TerminatedPhase) isDone() bool {
	return true
}
