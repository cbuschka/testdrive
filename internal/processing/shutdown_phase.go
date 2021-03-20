package processing

type ShutdownPhase struct {
	session *Session
}

func (phase *ShutdownPhase) String() string {
	return "PHASE_SHUTDOWN"
}

func (phase *ShutdownPhase) postHandle() (Phase, error) {

	err := phase.session.stopRunningContainers()
	if err != nil {
		return nil, err
	}

	err = phase.session.destroyNonRunningContainers()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersDestroyed() {
		return Phase(&TerminatedPhase{session: phase.session}), nil
	}

	return Phase(phase), nil
}

func (phase *ShutdownPhase) isDone() bool {
	return false
}
