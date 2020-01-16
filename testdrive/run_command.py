import logging
from optparse import OptionParser

from testdrive.config.config import Config
from testdrive.console_output import console_output
from testdrive.context import Context
from testdrive.docker_event_watcher import DockerEventWatcher
from testdrive.event_timer import EventTimer
from testdrive.run_model import RunModel
from testdrive.runner import Runner

log = logging.getLogger(__name__)


class RunCommand:
    def __init__(self):
        self.context = Context()
        self.run_model = RunModel(self.context)
        self.runner = Runner(self.context, self.run_model)
        self.event_timer = EventTimer(queue=self.runner.eventQueue)
        self.event_watcher = DockerEventWatcher(docker_client=self.context.docker_client,
                                                queue=self.runner.eventQueue)

    def run(self):
        self.__configure()
        config = Config.from_file("testdrive.yml")

        try:
            console_output.print("Test session id: {}", self.context.testSessionId)

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
        self.runner.shutdown()

    def __add_services_from(self, services):
        for name, config in services.items():
            self.run_model.add_service(name, config)

    def __configure(self):
        parser = OptionParser(prog="testdrive",
                              description="A test driver for docker container based integration tests.",
                              version="devel",
                              epilog="For more information visit https://github.io/cbuschka/testdrive")
        parser.add_option("-v", "--verbose",
                          action="store_true", dest="verbose", default=False,
                          help="verbose output, default not verbose")
        parser.add_option("--no-color",
                          action="store_false", dest="colorize", default=True,
                          help="colorize output, default enabled")
        (options, args) = parser.parse_args()

        if len(args) > 0:
            log.warning("Args {} ignored.", args)

        console_output.verbose = options.verbose
        console_output.colorize = options.colorize
