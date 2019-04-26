from __future__ import absolute_import
from __future__ import print_function
from __future__ import unicode_literals

import logging
from threading import Thread

from testdrive.event import Event

log = logging.getLogger(__name__)


class DockerEventWatcher(object):
    def __init__(self, docker_client, queue):
        self.queue = queue
        self.thread = None
        self.docker_client = docker_client

    def start(self):
        if self.thread is not None:
            return

        self.thread = Thread(target=self.read_events)
        self.thread.setDaemon(True)
        self.thread.start()

    def read_events(self):
        for dockerEvent in self.docker_client.events(decode=True):
            event = self.__toEvent(dockerEvent)
            if event is not None:
                # log.info("Event from docker: %s", event)
                self.queue.put(event)

    def __toEvent(self, event):
        if event["Type"] == 'container' and event["Action"] == 'create':
            return Event(type='containerCreated', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'start':
            return Event(type='containerStarted', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'kill':
            return Event(type='containerDied', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'die':
            return Event(type='containerDied', data=event)
        else:
            # log.info("Docker event ignored: %s", event)
            pass

    def stop(self):
        if self.thread is None:
            return

        self.thread.stop()
        self.thread = None
