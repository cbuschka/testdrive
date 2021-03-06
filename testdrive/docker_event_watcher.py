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

        self.thread = Thread(target=self.__read_events)
        self.thread.setDaemon(True)
        self.thread.start()

    def __read_events(self):
        myThread = self.thread

        for dockerEvent in self.docker_client.events(decode=True):
            if myThread != self.thread:
                break

            event = self.__toEvent(dockerEvent)
            if event is not None:
                self.queue.put(event)

    def __toEvent(self, event):
        if event["Type"] == 'container' and event["Action"] == 'create':
            return Event(type='containerCreated', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'start':
            return Event(type='containerStarted', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'kill':
            return Event(type='containerStopping', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'die':
            return Event(type='containerStopped', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'stop':
            return Event(type='containerStopping', data=event)
        elif event["Type"] == 'container' and event["Action"] == 'destroy':
            return Event(type='containerDestroyed', data=event)
        elif event["Type"] == 'container' and event["Action"].startswith('exec_create'):
            pass
        elif event["Type"] == 'container' and event["Action"].startswith('exec_start'):
            pass
        elif event["Type"] == 'container' and event["Action"].startswith('exec_die'):
            pass
        elif event["Type"] == 'network':
            pass
        elif event["Type"] == 'volume':
            pass
        else:
            log.warning("Event %s ignored.", event)
            pass

    def stop(self):
        self.thread = None
