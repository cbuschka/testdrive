from __future__ import absolute_import
from __future__ import print_function
from __future__ import unicode_literals

import logging
import sys
import signal
import docker
from testdrive.callable import Callable

signal_handlers = []


class Context:
    def __init__(self, args=sys.argv[1:]):
        self.args = args
        self.setup_logging()
        self.docker_client = docker.from_env()

        for curr_signal in [signal.SIGTERM, signal.SIGINT]:
            signal.signal(curr_signal, Callable(self.__handle_signal))

    def setup_logging(self):
        logging.basicConfig(level=logging.INFO, stream=sys.stderr)

    def add_signal_handler(self, handler):
        signal_handlers.append(handler)

    def remove_signal_handler(self, handler):
        if handler in signal_handlers:
            signal_handlers.remove(handler)

    def __handle_signal(self):
        for signal_handler in signal_handlers:
            signal_handler()
