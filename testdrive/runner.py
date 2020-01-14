import logging
from queue import Queue, Empty

from testdrive.callable import Callable

log = logging.getLogger(__name__)

from testdrive.console_output import console_output
from testdrive.run_model import Status


class Runner(object):
    def __init__(self, context, model):
        self.context = context
        self.model = model
        self.eventQueue = Queue()
        self.done = False

    def run(self):
        if not "driver" in self.model.services:
            raise ValueError("No driver.")

        while not self.done:
            try:
                event = self.eventQueue.get(timeout=0.3)
            except (Empty) as e:
                continue

            if event.type in ["tick", "serviceAdded", "driverAdded"]:
                self.__onTick()
            elif event.type in ["containerCreated"]:
                self.__onContainerCreated(event)
            elif event.type in ["containerStarted"]:
                self.__onContainerStarted(event)
            elif event.type in ["containerDied"]:
                self.__onContainerDied(event)
            else:
                # log.info("Event %s ignored.", event)
                pass

        return self.model.services["driver"].exitCode

    def shutdown(self):
        self.done = True
        services = [service for name, service in self.model.services.items()
                    if service.container != None
                    and service.status in [Status.STARTED, Status.START_IN_PROGRESS, Status.READY,
                                           Status.HEALTHCHECK_IN_PROGRESS, Status.CREATED, Status.STOPPED]]
        for service in services:
            self.model.stopServiceContainer(service)

    def __get_actions(self):
        actions = []
        containersLeft = False
        for name, service in self.model.services.items():
            if service.status == Status.NEW:
                actions.append(Callable(self.model.createServiceContainer, service))
                containersLeft = True
            elif self.model.canStart(service):
                actions.append(Callable(self.model.startServiceContainer, service))
                containersLeft = True
            elif service.status == Status.STARTED:
                actions.append(Callable(self.model.checkServiceContainer, service))
                containersLeft = True
            elif service.status == Status.READY:
                containersLeft = True
                pass
            elif service.status == Status.STOPPED:
                actions.append(Callable(self.model.removeServiceContainer, service))
            elif service.status == Status.DESTROYED:
                pass
            else:
                pass
        return (actions, containersLeft)

    def __onTick(self):
        if self.model.services["driver"].status in [Status.STOPPED, Status.DESTROYED]:
            self.done = True
            return

        (actions, containersLeft) = self.__get_actions()
        if len(actions) == 0 and not containersLeft:
            self.done = True

        for action in actions:
            action()

    def __onContainerCreated(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.model.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.CREATED
            console_output.print("Service {} created.", service.name)

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

    def __onContainerDied(self, event):
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
