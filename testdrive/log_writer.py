from __future__ import absolute_import
from __future__ import unicode_literals

from threading import Thread

from testdrive.console_output import console_output


class LogWriter(object):
    def __init__(self, name, stream):
        self.name = name
        self.thread = None
        self.stream = stream

    def start(self):
        if self.thread is not None:
            return

        self.thread = Thread(target=self.__read_logs)
        self.thread.setDaemon(True)
        self.thread.start()

    def __read_logs(self):
        for line in self.stream:
            console_output.print("[{}] {}", self.name, line.decode('unicode_escape').strip())

    def stop(self):
        if self.thread is None:
            return

        self.thread.stop()
        self.thread = None
