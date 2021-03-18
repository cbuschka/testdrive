package internal

type StartupPhase struct {
	session *Session
}

func (phase *StartupPhase) String() string {
	return "PHASE_STARTUP"
}

func (phase *StartupPhase) postHandle() (Phase, error) {

	err := phase.session.createContainersForCreatableContainers("service")
	if err != nil {
		return nil, err
	}

	err = phase.session.startContainersForStartableServices()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersReady("service") {
		return Phase(&RunningPhase{session: phase.session}), nil
	}

	return Phase(phase), nil
}

func (phase *StartupPhase) isDone() bool {
	return false
}
