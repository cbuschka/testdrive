from __future__ import absolute_import
from __future__ import print_function
from __future__ import unicode_literals

import logging
import sys

from testdrive.context import Context
from testdrive.run_command import RunCommand

log = logging.getLogger(__name__)


def main():
    try:
        context = Context()
        command = RunCommand(context)
        exitCode = command.run()
        sys.exit(exitCode)
    except (KeyboardInterrupt):
        log.error("Aborting.")
        sys.exit(1)
