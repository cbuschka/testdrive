package processing

import "github.com/cbuschka/testdrive/internal/model"

type StartupPhase struct {
	session *Session
}

func (phase *StartupPhase) String() string {
	return "PHASE_STARTUP"
}

func (phase *StartupPhase) postHandle() (Phase, error) {

	err := phase.session.createContainersForCreatableContainers(model.Service)
	if err != nil {
		return nil, err
	}

	err = phase.session.startContainersForStartableServices()
	if err != nil {
		return nil, err
	}

	if phase.session.allContainersReady(model.Service) {
		return Phase(&RunningPhase{session: phase.session}), nil
	}

	return Phase(phase), nil
}

func (phase *StartupPhase) isDone() bool {
	return false
}
