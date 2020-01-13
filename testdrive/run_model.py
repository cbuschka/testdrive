import logging
from enum import Enum
from queue import Queue, Empty

from docker.errors import NotFound

from testdrive.callable import Callable
from testdrive.log_writer import LogWriter

log = logging.getLogger(__name__)

from testdrive.console_output import console_output


class Status(Enum):
    NEW = 1
    CREATE_IN_PROGRESS = 2
    CREATED = 3
    START_IN_PROGRESS = 4
    STARTED = 5
    AVAILABLE_IN_PROGRESS = 6
    AVAILABLE = 7
    STOP_IN_PROGRESS = 8
    STOPPED = 9
    DESTROY_IN_PROGRESS = 10
    DESTROYED = 11


class DriverOrService(object):
    def __init__(self, name, config):
        self.name = name
        self.config = config
        self.status = Status.NEW
        self.container = None
        self.exitCode = None


class RunModel(object):
    def __init__(self, context):
        self.context = context
        self.services = {}
        self.eventQueue = Queue()
        self.done = False

    def set_driver(self, config):
        self.services["driver"] = DriverOrService("driver", config)

    def add_service(self, name, config):
        self.services[name] = DriverOrService(name, config)

    def run(self):
        if not "driver" in self.services:
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

        return self.services["driver"].exitCode

    def shutdown(self):
        self.done = True
        services = [service for name, service in self.services.items()
                    if service.container != None
                    and service.status in [Status.STARTED, Status.START_IN_PROGRESS, Status.AVAILABLE,
                                           Status.AVAILABLE_IN_PROGRESS, Status.CREATED, Status.STOPPED]]
        for service in services:
            self.__stopServiceContainer(service)

    def __get_actions(self):
        actions = []
        containersLeft = False
        for name, service in self.services.items():
            if service.status == Status.NEW:
                actions.append(Callable(self.__createServiceContainer, service))
                containersLeft = True
            elif self.__canStart(service):
                actions.append(Callable(self.__startServiceContainer, service))
                containersLeft = True
            elif service.status == Status.STARTED:
                actions.append(Callable(self.__checkServiceContainer, service))
                containersLeft = True
            elif service.status == Status.AVAILABLE:
                containersLeft = True
                pass
            elif service.status == Status.STOPPED:
                actions.append(Callable(self.__removeServiceContainer, service))
            elif service.status == Status.DESTROYED:
                pass
            else:
                pass
        return (actions, containersLeft)

    def __canStart(self, service):
        if service.status != Status.CREATED:
            return False

        for dependency in service.config.get("depends_on", []):
            if self.services[dependency].status not in [Status.AVAILABLE]:
                return False

        return True

    def __createServiceContainer(self, service):
        if service.status != Status.NEW:
            log.warning("Cannot create container %s (%s) because not NEW.", service.name, service.status)
            return

        console_output.print("Creating service {}...", service.name)
        docker_client = self.context.docker_client
        service.status = Status.CREATE_IN_PROGRESS
        image = service.config["image"]
        healthcheck = service.config.get("healthcheck", {"test": []})
        command = service.config.get("command", None)
        service.container = docker_client.containers.create(image=image, command=command, healthcheck=healthcheck)

    def __startServiceContainer(self, service):
        if service.status != Status.CREATED:
            log.warning("Cannot create container %s (%s) because not CREATED.", service.name, service.status)
            return

        console_output.print("Starting service {}...", service.name)
        service.status = Status.START_IN_PROGRESS
        service.container.start()

        stream = service.container.logs(stdout=True, stderr=True, stream=True, follow=True)
        LogWriter(service.name, stream).start()

    def __checkServiceContainer(self, service):
        if service.status != Status.STARTED:
            log.warning("Cannot check container %s (%s) because not STARTED.", service.name, service.status)
            return

        console_output.print("Checking service {}...", service.name)
        try:
            service.status = Status.AVAILABLE_IN_PROGRESS
            readycheck = service.config["readycheck"]
            (exitCode, output) = service.container.exec_run(cmd=readycheck["command"],
                                                            user=readycheck.get("user", None))
            if exitCode == 0:
                service.status = Status.AVAILABLE
                console_output.print("Service {} now ready.", service.name)
            else:
                service.status = Status.STARTED
                console_output.print("Service {} still NOT ready. (exitCode={})", service.name, exitCode)
        except (NotFound) as e:
            service.status = Status.STARTED

    def __stopServiceContainer(self, service):
        if service.status not in [Status.STARTED, Status.START_IN_PROGRESS, Status.AVAILABLE,
                                  Status.AVAILABLE_IN_PROGRESS, Status.CREATED, Status.STOPPED]:
            log.warning("Cannot stop container %s (%s) because not stoppable.", service.name, service.status)
            return

        if service.status == Status.CREATED:
            service.status = Status.STOPPED
        elif service.status == Status.STOPPED:
            pass
        else:
            console_output.print("Stopping service {}...", service.name)
            service.status = Status.STOP_IN_PROGRESS
            service.container.stop()

    def __removeServiceContainer(self, service):
        if service.container == None:
            return

        console_output.print("Removing service {}...", service.name)
        try:
            service.status = Status.DESTROY_IN_PROGRESS
            service.container.remove(force=True, v=True)
        except (NotFound) as e:
            service.container = None
            service.status = Status.DESTROYED

    def __onTick(self):
        if self.services["driver"].status in [Status.STOPPED, Status.DESTROYED]:
            self.done = True
            return

        (actions, containersLeft) = self.__get_actions()
        if len(actions) == 0 and not containersLeft:
            self.done = True

        for action in actions:
            action()

    def __onContainerCreated(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.CREATED
            console_output.print("Service {} created.", service.name)

    def __onContainerStarted(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            if "readycheck" not in service.config.keys():
                service.status = Status.AVAILABLE
            else:
                service.status = Status.STARTED
            console_output.print("Service {} started.", service.name)

    def __onContainerDied(self, event):
        containerId = event.data["id"]
        services = [service for name, service in self.services.items() if
                    service.container is not None and service.container.id == containerId]
        for service in services:
            service.status = Status.STOPPED
            exitCodeStr = event.data.get("Actor", {}).get("Attributes", {}).get("exitCode", None)
            service.exitCode = int(exitCodeStr or "127")
            console_output.print("Service {} stopped (exit code={}).", service.name, service.exitCode)
