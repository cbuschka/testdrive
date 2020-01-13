import logging
import time
from threading import Thread

from testdrive.event import Event

log = logging.getLogger(__name__)


class EventTimer(object):
    def __init__(self, queue):
        self.queue = queue
        self.thread = None

    def start(self):
        if self.thread is not None:
            return

        self.thread = Thread(target=self.__run_loop)
        self.thread.setDaemon(True)
        self.thread.start()

    def __run_loop(self):
        myThread = self.thread

        while myThread == self.thread:
            time.sleep(1)
            if self.queue is not None:
                self.queue.put(Event(type="tick", data=None))

    def stop(self):
        self.thread = None
