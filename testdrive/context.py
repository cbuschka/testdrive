from __future__ import absolute_import
from __future__ import print_function
from __future__ import unicode_literals

import logging
import sys

import docker


class Context:
    def __init__(self, args=sys.argv[1:]):
        self.args = args
        self.setup_logging()
        self.docker_client = docker.from_env()

    def setup_logging(self):
        logging.basicConfig(level=logging.INFO, stream=sys.stderr)
