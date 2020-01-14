import logging
from enum import Enum

from docker.errors import NotFound

from testdrive.log_writer import LogWriter

log = logging.getLogger(__name__)

from testdrive.console_output import console_output


class Status(Enum):
    NEW = 1
    CREATE_IN_PROGRESS = 2
    CREATED = 3
    START_IN_PROGRESS = 4
    STARTED = 5
    HEALTHCHECK_IN_PROGRESS = 6
    READY = 7
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

    def set_driver(self, config):
        self.services["driver"] = DriverOrService("driver", config)

    def add_service(self, name, config):
        self.services[name] = DriverOrService(name, config)

    def canStart(self, service):
        if service.status != Status.CREATED:
            return False

        for dependency in service.config.get("depends_on", []):
            if self.services[dependency].status not in [Status.READY]:
                return False

        return True

    def createServiceContainer(self, service):
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

    def startServiceContainer(self, service):
        if service.status != Status.CREATED:
            log.warning("Cannot create container %s (%s) because not CREATED.", service.name, service.status)
            return

        console_output.print("Starting service {}...", service.name)
        service.status = Status.START_IN_PROGRESS
        service.container.start()

        stream = service.container.logs(stdout=True, stderr=True, stream=True, follow=True)
        LogWriter(service.name, stream).start()

    def checkServiceContainer(self, service):
        if service.status != Status.STARTED:
            log.warning("Cannot check container %s (%s) because not STARTED.", service.name, service.status)
            return

        console_output.print("Checking service {}...", service.name)
        try:
            service.status = Status.HEALTHCHECK_IN_PROGRESS
            readycheck = service.config["readycheck"]
            (exitCode, output) = service.container.exec_run(cmd=readycheck["command"],
                                                            user=readycheck.get("user", None))
            if exitCode == 0:
                service.status = Status.READY
                console_output.print("Service {} now ready.", service.name)
            else:
                service.status = Status.STARTED
                console_output.print("Service {} still NOT ready. (exitCode={})", service.name, exitCode)
        except (NotFound):
            console_output.print("Service {} not found.", service.name)

    def stopServiceContainer(self, service):
        if service.status not in [Status.STARTED, Status.START_IN_PROGRESS, Status.READY,
                                  Status.HEALTHCHECK_IN_PROGRESS, Status.CREATED, Status.STOPPED]:
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

    def removeServiceContainer(self, service):
        if service.container == None:
            return

        console_output.print("Removing service {}...", service.name)
        try:
            service.status = Status.DESTROY_IN_PROGRESS
            service.container.remove(force=True, v=True)
        except (NotFound) as e:
            service.container = None
            service.status = Status.DESTROYED
