package processing

import "github.com/cbuschka/testdrive/internal/model"

type RunningPhase struct {
	session *Session
}

func (phase *RunningPhase) String() string {
	return "PHASE_RUNNING"
}

func (phase *RunningPhase) postHandle() (Phase, error) {

	err := phase.session.createContainersForCreatableContainers(model.Task)
	if err != nil {
		return nil, err
	}

	err = phase.session.startContainersForStartableTasks()
	if err != nil {
		return nil, err
	}

	if phase.session.allTaskContainersStopped() {
		return Phase(&ShutdownPhase{session: phase.session}), nil
	}

	return Phase(phase), nil
}

func (phase *RunningPhase) isDone() bool {
	return false
}
