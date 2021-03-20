package processing

type TerminatedPhase struct {
	session *Session
}

func NewTerminatedPhase(session *Session) *TerminatedPhase {
	return &TerminatedPhase{session: session}
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
