package internal

type RunningPhase struct {
	session *Session
}

func (phase *RunningPhase) String() string {
	return "PHASE_RUNNING"
}

func (phase *RunningPhase) postHandle(event *Event) (Phase, error) {

	err := phase.session.createContainersForCreatableTasks("task")
	if err != nil {
		return nil, err
	}

	err = phase.session.startContainersForStartableTasks()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersStopped("task") {
		return Phase(&ShutdownPhase{session: phase.session}), nil
	}

	return Phase(phase), nil
}

func (phase *RunningPhase) isDone() bool {
	return false
}
