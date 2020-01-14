import logging
import signal

import docker
import sys

from testdrive.console_output import console_output

signal_handlers = []
SIGNAL_NAMES = {
    signal.SIGINT.value: "SIGINT",
    signal.SIGTERM.value: "SIGTERM",
}


def handle_signal(signum, frame):
    console_output.print("Received signal {}...", SIGNAL_NAMES.get(signum, signum))
    for signal_handler in signal_handlers:
        signal_handler()


class Context:
    def __init__(self, args=None):
        self.args = args or sys.argv[1:]
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
