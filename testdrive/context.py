import logging
import signal

import docker
import random_name
import sys

signal_handlers = []
SIGNAL_NAMES = {
    signal.SIGINT.value: "SIGINT",
    signal.SIGTERM.value: "SIGTERM",
}


def handle_signal(signum, frame):
    # console_output.print("Received signal {}...", SIGNAL_NAMES.get(signum, signum))
    for signal_handler in signal_handlers:
        signal_handler()


class Context:
    def __init__(self, testSessionId=None, args=None, verbose=False):
        self.args = args
        self.testSessionId = testSessionId or random_name.generate_name(separator='_', lists=[random_name.ADJECTIVES,
                                                                                              random_name.ANIMALS])
        self.setup_logging()
        self.docker_client = docker.from_env()
        self.sequence = 1
        self.verbose = verbose

        for curr_signal in [signal.SIGTERM, signal.SIGINT]:
            signal.signal(curr_signal, handle_signal)

    def next_seqnum(self):
        value = self.sequence
        self.sequence += 1
        return value

    def setup_logging(self):
        logging.basicConfig(level=logging.INFO, stream=sys.stderr)

    def add_signal_handler(self, handler):
        signal_handlers.append(handler)

    def remove_signal_handler(self, handler):
        if handler in signal_handlers:
            signal_handlers.remove(handler)
