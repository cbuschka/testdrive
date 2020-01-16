import logging
from enum import Enum
from queue import Queue, Empty

from testdrive.callable import Callable

log = logging.getLogger(__name__)

from testdrive.console_output import console_output
from testdrive.run_model import Status
from testdrive.epoch_time import now


def nop():
    pass


class Phase(Enum):
    EXECUTION = 1
    SHUTDOWN = 2
    DONE = 3


class Runner(object):
    def __init__(self, context, model):
        self.context = context
        self.model = model
        self.eventQueue = Queue()
        self.done = False
        self.phase = Phase.EXECUTION

    def run(self):
        if not "driver" in self.model.services:
            raise ValueError("No driver.")

        while self.phase != Phase.DONE:
            try:
                event = self.eventQueue.get(timeout=0.3)
            except (Empty):
                continue

            if event.type in ["tick", "serviceAdded", "driverAdded"]:
                pass
            elif event.type in ["containerCreated"]:
                self.__onContainerCreated(event)
            elif event.type in ["containerStarted"]:
                self.__onContainerStarted(event)
            elif event.type in ["containerStopping"]:
                self.__onContainerStopping(event)
            elif event.type in ["containerStopped"]:
                self.__onContainerStopped(event)
            elif event.type in ["containerDestroyed"]:
                self.__onContainerDestroyed(event)
            else:
                log.warning("Unknown event %s ignored.", event)
                pass

            self.__onTick()

        return self.model.services["driver"].exitCode

    def shutdown(self):
        if self.phase == Phase.SHUTDOWN:
            console_output.print_important("Already shutting down. Please be patient...")
            return

        console_output.print_important("Shutting down...")
        self.phase = Phase.SHUTDOWN
        services = [service for name, service in self.model.services.items()
                    if service.container != None
                    and service.status in [Status.STARTED, Status.READY, Status.HEALTHCHECK_IN_PROGRESS]]
        for service in services:
            self.model.stopServiceContainer(service)

    def __get_actions(self):
        if self.phase == Phase.EXECUTION:
            return self.__get_actions_for_execution()
        elif self.phase == Phase.SHUTDOWN:
            return self.__get_actions_for_shutdown()
        else:
            return []

    def __get_actions_for_execution(self):
        actions = []
        for name, service in self.model.services.items():
            if service.status == Status.NEW:
                actions.append(Callable(self.model.createServiceContainer, service))
            elif service.status == Status.CREATE_IN_PROGRESS:
                actions.append(
                    Callable(console_output.print_verbose, "Waiting for {} to be created...".format(service.name)))
            elif service.status == Status.CREATED:
                if self.model.canStart(service):
                    actions.append(Callable(self.model.startServiceContainer, service))
                else:
                    pass
            elif service.status == Status.START_IN_PROGRESS:
                actions.append(Callable(console_output.print, "Waiting for {} to be started...".format(service.name)))
            elif service.status == Status.STARTED:
                actions.append(Callable(self.model.startHealthcheckForServiceContainer, service))
            elif service.status == Status.HEALTHCHECK_IN_PROGRESS:
                actions.append(Callable(self.model.checkServiceContainer, service))
            elif service.status == Status.READY:
                if service.name == 'driver':
                    actions.append(Callable(nop))
                else:
                    pass
            elif service.status == Status.STOP_IN_PROGRESS:
                actions.append(
                    Callable(console_output.print, "Waiting for {} to be stopped...".format(service.name)))
            elif service.status == Status.STOPPED:
                actions.append(Callable(self.model.removeServiceContainer, service))
            elif service.status == Status.DESTROY_IN_PROGRESS:
                actions.append(
                    Callable(console_output.print, "Waiting for {} to be destroyed...".format(service.name)))
            elif service.status == Status.DESTROYED:
                pass
            else:
                log.warning("Service {} has unknown status {}.", service.name, service.status)
        return actions

    def __get_actions_for_shutdown(self):
        actions = []
        for name, service in self.model.services.items():
            if service.status == Status.NEW:
                pass
            elif service.status == Status.CREATE_IN_PROGRESS:
                actions.append(
                    Callable(console_output.print, "Waiting for {} to be created...".format(service.name)))
            elif service.status == Status.CREATED:
                actions.append(Callable(self.model.removeServiceContainer, service))
            elif service.status == Status.START_IN_PROGRESS:
                actions.append(
                    Callable(console_output.print, "Waiting for {} to finish startup...".format(service.name)))
            elif service.status == Status.STARTED:
                actions.append(Callable(self.model.stopServiceContainer, service))
            elif service.status == Status.HEALTHCHECK_IN_PROGRESS:
                actions.append(Callable(self.model.stopServiceContainer, service))
            elif service.status == Status.READY:
                actions.append(Callable(self.model.stopServiceContainer, service))
            elif service.status == Status.STOP_IN_PROGRESS:
                if now() > service.timeout:
                    actions.append(Callable(self.model.killServiceContainer, service))
                else:
                    actions.append(
                        Callable(console_output.print_verbose, "Waiting for {} to be stopped...".format(service.name)))
            elif service.status == Status.STOPPED:
                actions.append(Callable(self.model.removeServiceContainer, service))
            elif service.status == Status.DESTROY_IN_PROGRESS:
                if now() > service.timeout:
                    actions.append(
                        Callable(console_output.print_verbose, "Destroying {} delayed...".format(service.name)))
                else:
                    actions.append(
                        Callable(console_output.print_verbose,
                                 "Waiting for {} to be destroyed...".format(service.name)))
            elif service.status == Status.DESTROYED:
                pass
            else:
                pass

        return actions

    def __onTick(self):
        actions = self.__get_actions()
        if len(actions) == 0:
            self.phase = Phase.DONE

        for action in actions:
            action()

    def __onContainerCreated(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.CREATED
            console_output.print_verbose("Service {} created.", service.name)

    def __onContainerStarted(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            if "readycheck" not in service.config.keys():
                service.status = Status.READY
            else:
                service.status = Status.STARTED
            console_output.print("Service {} started.", service.name)

    def __onContainerStopping(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.STOP_IN_PROGRESS

    def __onContainerStopped(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.STOPPED
            exitCodeStr = event.data.get("Actor", {}).get("Attributes", {}).get("exitCode", None)
            try:
                service.exitCode = int(exitCodeStr)
            except (TypeError):
                service.exitCode = 127
            console_output.print("Service {} stopped (exit code={}).", service.name, service.exitCode)
            if service.name == 'driver':
                self.phase = Phase.SHUTDOWN

    def __onContainerDestroyed(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.DESTROYED
            console_output.print_verbose("Service {} destroyed.", service.name)
