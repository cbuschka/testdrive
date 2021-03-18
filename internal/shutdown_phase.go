package internal

type ShutdownPhase struct {
	session *Session
}

func (phase *ShutdownPhase) String() string {
	return "PHASE_SHUTDOWN"
}

func (phase *ShutdownPhase) postHandle(event *Event) (Phase, error) {

	err := phase.session.stopRunningContainers()
	if err != nil {
		return nil, err
	}

	err = phase.session.destroyStoppedContainers()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersDestroyed() {
		return Phase(&ShutdownPhase{session: phase.session}), nil
	}

	return Phase(&TerminatedPhase{session: phase.session}), nil
}

func (phase *ShutdownPhase) isDone() bool {
	return false
}
