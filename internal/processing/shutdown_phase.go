package processing

type ShutdownPhase struct {
	session *Session
}

func NewShutdownPhase(session *Session) *ShutdownPhase {
	return &ShutdownPhase{session: session}
}

func (phase *ShutdownPhase) String() string {
	return "PHASE_SHUTDOWN"
}

func (phase *ShutdownPhase) postHandle() (Phase, error) {

	err := phase.session.stopRunningContainers()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersNonRunning() {
		return Phase(NewCleanupPhase(phase.session)), nil
	}

	return Phase(phase), nil
}

func (phase *ShutdownPhase) isDone() bool {
	return false
}
