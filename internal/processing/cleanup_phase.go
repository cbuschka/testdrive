package processing

type CleanupPhase struct {
	session *Session
}

func NewCleanupPhase(session *Session) *CleanupPhase {
	return &CleanupPhase{session: session}
}

func (phase *CleanupPhase) String() string {
	return "PHASE_CLEANUP"
}

func (phase *CleanupPhase) postHandle() (Phase, error) {

	err := phase.session.destroyAllNonDestroyedContainers()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersDestroyed() {
		return Phase(NewTerminatedPhase(phase.session)), nil
	}

	return Phase(phase), nil
}

func (phase *CleanupPhase) isDone() bool {
	return false
}
