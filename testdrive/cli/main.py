import sys

from testdrive.console_output import console_output
from testdrive.run_command import RunCommand


def main():
    console_output.print_banner()

    try:
        command = RunCommand()
        exitCode = command.run()
        sys.exit(exitCode)
    except (KeyboardInterrupt):
        console_output.print("Aborting...")
        sys.exit(1)
