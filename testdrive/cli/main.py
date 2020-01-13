import sys

from testdrive.console_output import console_output
from testdrive.context import Context
from testdrive.run_command import RunCommand


def main():
    console_output.print_banner()

    try:
        context = Context()
        command = RunCommand(context)
        exitCode = command.run()
        sys.exit(exitCode)
    except (KeyboardInterrupt):
        console_output.print("Aborting...")
        sys.exit(1)
