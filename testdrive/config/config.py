import io
import json
import os

import jsonschema
import yaml


def load_schema():
    script_dir = os.path.dirname(__file__)
    file_path = os.path.join(script_dir, 'config_schema_v1.0.json')
    with open(file_path, 'r') as file:
        return json.load(file)


SCHEMA_V1_0 = load_schema()


class Config(object):
    @classmethod
    def from_file(cls, filename):
        config = load_yaml(filename)
        jsonschema.validate(instance=config, schema=SCHEMA_V1_0)
        return cls(config)

    def __init__(self, data):
        self.data = data


def load_yaml(filename):
    try:
        with io.open(filename, 'r', encoding='utf-8') as fh:
            return yaml.safe_load(fh)
    except (IOError, yaml.YAMLError, UnicodeDecodeError) as e:
        raise e
