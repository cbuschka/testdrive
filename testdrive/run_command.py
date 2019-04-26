from __future__ import absolute_import
from __future__ import unicode_literals

import logging

from testdrive.config.config import Config
from testdrive.event_timer import EventTimer
from testdrive.event_watcher import EventWatcher
from testdrive.run_model import RunModel

log = logging.getLogger(__name__)


class RunCommand:
    def __init__(self, context):
        self.context = context

    def run(self):
        config = Config.from_file("testdrive.yml")
        self.run_model = RunModel(context=self.context)

        event_timer = EventTimer(queue=self.run_model.eventQueue)
        event_timer.start()

        event_watcher = EventWatcher(docker_client=self.context.docker_client, queue=self.run_model.eventQueue)
        event_watcher.start()

        self.run_model.set_driver(config.data["driver"])
        if "services" in config.data:
            self.__add_services_from(config.data["services"])

        return self.run_model.run()

    def __add_services_from(self, services):
        for name, config in services.items():
            self.run_model.add_service(name, config)
