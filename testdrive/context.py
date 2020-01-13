import logging
import signal

import docker
import sys

signal_handlers = []


def handle_signal(signum, frame):
    for signal_handler in signal_handlers:
        signal_handler()


class Context:
    def __init__(self, args=None):
        self.args = sys.argv[1:] if args is None else args
        self.setup_logging()
        self.docker_client = docker.from_env()

        for curr_signal in [signal.SIGTERM, signal.SIGINT]:
            signal.signal(curr_signal, handle_signal)

    def setup_logging(self):
        logging.basicConfig(level=logging.INFO, stream=sys.stderr)

    def add_signal_handler(self, handler):
        signal_handlers.append(handler)

    def remove_signal_handler(self, handler):
        if handler in signal_handlers:
            signal_handlers.remove(handler)
