from __future__ import print_function
from __future__ import unicode_literals

import datetime
import time


class ConsoleOutput(object):
    def print(self, s, *args):
        ts = time.time()
        formattedTs = datetime.datetime.fromtimestamp(ts).strftime('%Y-%m-%d %H:%M:%S')
        print((formattedTs + " " + s.strip()).format(*args), flush=True)


console_output = ConsoleOutput()
