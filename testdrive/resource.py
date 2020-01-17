import logging
from enum import Enum

import docker
from docker.errors import NotFound

from testdrive.epoch_time import now
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


class Resource(object):
    def __init__(self, type, name, config, seqNum, testSessionId, dependencies=None):
        self.seqNum = seqNum
        self.type = type
        self.name = name
        self.config = config
        self.status = Status.NEW
        self.container = None
        self.exitCode = None
        self.timeout = 0
        self.dependencies = config.get("depends_on", []).copy()
        if dependencies is not None:
            self.dependencies.extend(dependencies)
        self.container_name = "{}_{}_{}".format(testSessionId, self.name, self.seqNum)

    def create(self, context):
        if self.status != Status.NEW:
            log.warning("Cannot create container %s (%s) because not NEW.", self.name, self.status)
            return

        console_output.print_verbose("Creating service {}...", self.name)
        self.status = Status.CREATE_IN_PROGRESS
        image = self.config["image"]
        healthcheck = self.config.get("healthcheck", {"test": []})
        command = self.config.get("command", None)
        self.container = context.docker_client.containers.create(image=image, command=command,
                                                                 healthcheck=healthcheck,
                                                                 name=self.container_name,
                                                                 auto_remove=True,
                                                                 tty=False, detach=True,
                                                                 oom_kill_disable=False,
                                                                 labels={}, init=False, cap_add=[],
                                                                 devices=[],
                                                                 domainname=None, entrypoint=None,
                                                                 environment={},
                                                                 extra_hosts={},
                                                                 hostname=self.config.get("hostname", None),
                                                                 links={},
                                                                 mounts=self.config.get("volumes", []),
                                                                 volumes_from=self.config.get("volumes_from", []), )

    def start(self, context):
        if self.status != Status.CREATED:
            log.warning("Cannot start container %s (%s) because not CREATED.", self.name, self.status)
            return

        console_output.print_verbose("Starting service {}...", self.name)
        self.status = Status.START_IN_PROGRESS
        self.timeout = now() + 10_000
        self.container.start()

        stream = self.container.logs(stdout=True, stderr=True, stream=True, follow=True)
        LogWriter(self.name, stream, color=self.seqNum).start()

    def stop(self, context):
        if self.status not in [Status.STARTED, Status.START_IN_PROGRESS, Status.READY,
                               Status.HEALTHCHECK_IN_PROGRESS, Status.CREATED,
                               Status.CREATE_IN_PROGRESS,
                               Status.STOPPED, Status.STOP_IN_PROGRESS]:
            log.warning("Cannot stop container %s (%s) because not stoppable.", self.name, self.status)
            return

        if self.status == Status.CREATED or self.status == Status.CREATE_IN_PROGRESS:
            self.status = Status.STOPPED
        elif self.status == Status.STOPPED or self.status == Status.STOP_IN_PROGRESS:
            pass
        else:
            try:
                console_output.print_verbose("Stopping service {}...", self.name)
                self.status = Status.STOP_IN_PROGRESS
                self.timeout = now() + 10_000
                self.container.stop()
            except (docker.errors.NotFound):
                self.container = None
                self.status = Status.DESTROYED

    def kill(self, context):
        if self.container == None:
            return

        console_output.print("Killing service {}...", self.name)
        try:
            self.status = Status.STOP_IN_PROGRESS
            self.timeout = now() + 3_000
            self.container.kill()
        except (docker.errors.NotFound):
            self.container = None
            self.status = Status.DESTROYED

    def remove(self, context):
        if self.container == None:
            return

        console_output.print("Removing service {}...", self.name)
        try:
            self.status = Status.DESTROY_IN_PROGRESS
            self.container.remove(force=True, v=True)
        except (docker.errors.NotFound):
            self.container = None
            self.status = Status.DESTROYED
        except (docker.errors.APIError) as e:
            if e.status_code != 409:
                raise e

    def startHealthcheck(self, context):
        if self.status != Status.STARTED:
            log.warning("Cannot check container %s (%s) because not STARTED.", self.name, self.status)
            return

        self.status = Status.HEALTHCHECK_IN_PROGRESS

    def check(self, context):
        console_output.print_verbose("Checking service {}...", self.name)
        try:
            readycheck = self.config["readycheck"]
            (exitCode, output) = self.container.exec_run(cmd=readycheck["command"],
                                                         user=readycheck.get("user", None))
            if exitCode == 0:
                self.status = Status.READY
                console_output.print_verbose("Service {} now ready.", self.name)
            else:
                self.status = Status.STARTED
                console_output.print_verbose("Service {} still NOT ready. (exitCode={})", self.name, exitCode)
        except (docker.errors.NotFound):
            console_output.print("Service {} not found.", self.name)
