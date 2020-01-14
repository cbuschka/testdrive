import logging

from testdrive.config.config import Config
from testdrive.console_output import console_output
from testdrive.docker_event_watcher import DockerEventWatcher
from testdrive.event_timer import EventTimer
from testdrive.run_model import RunModel
from testdrive.runner import Runner

log = logging.getLogger(__name__)


class RunCommand:
    def __init__(self, context):
        self.context = context
        self.run_model = RunModel(context=self.context)
        self.runner = Runner(context, self.run_model)
        self.event_timer = EventTimer(queue=self.runner.eventQueue)
        self.event_watcher = DockerEventWatcher(docker_client=self.context.docker_client,
                                                queue=self.runner.eventQueue)

    def run(self):
        config = Config.from_file("testdrive.yml")

        try:
            self.__start()
            self.context.add_signal_handler(self.__shutdown)

            self.run_model.set_driver(config.data["driver"])
            if "services" in config.data:
                self.__add_services_from(config.data["services"])

            return self.runner.run()
        finally:
            self.event_timer.stop()
            self.event_watcher.stop()

    def __start(self):
        self.event_timer.start()
        self.event_watcher.start()

    def __shutdown(self):
        console_output.print("Shutting down...")
        self.runner.shutdown()

    def __add_services_from(self, services):
        for name, config in services.items():
            self.run_model.add_service(name, config)
