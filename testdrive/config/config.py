from __future__ import absolute_import
from __future__ import unicode_literals

import io

import yaml


class Config(object):
    @classmethod
    def from_file(cls, filename):
        return cls(load_yaml(filename))

    def __init__(self, data):
        self.data = data


def load_yaml(filename):
    try:
        with io.open(filename, 'r', encoding='utf-8') as fh:
            return yaml.safe_load(fh)
    except (IOError, yaml.YAMLError, UnicodeDecodeError) as e:
        raise e
