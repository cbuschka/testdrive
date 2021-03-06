import logging

log = logging.getLogger(__name__)


class Event(object):
    def __init__(self, type, data):
        self.type = type
        self.data = data

    def __str__(self):
        return "<Event type={},data={}>".format(self.type, self.data)
