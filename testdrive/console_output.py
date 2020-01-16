import datetime

import time


class Color:
    def __init__(self, start, end):
        self.start = start
        self.end = end


DONT_COLORIZE = Color("", "")
DEFAULT = Color("\033[39;49m", "\033[39;49m")
LIGHT_RED = Color("\033[91m", DEFAULT.end)
LIGHT_GREEN = Color("\033[92m", DEFAULT.end)
LIGHT_YELLOW = Color("\033[93m", DEFAULT.end)
LIGHT_BLUE = Color("\033[94m", DEFAULT.end)
LIGHT_MAGENTA = Color("\033[95m", DEFAULT.end)
LIGHT_CYAN = Color("\033[96m", DEFAULT.end)
LIGHT_GREY = Color("\033[97m", DEFAULT.end)
RED = Color("\033[31m", DEFAULT.end)
GREEN = Color("\033[32m", DEFAULT.end)
YELLOW = Color("\033[33m", DEFAULT.end)
BLUE = Color("\033[34m", DEFAULT.end)
MAGENTA = Color("\033[35m", DEFAULT.end)
CYAN = Color("\033[36m", DEFAULT.end)
GREY = Color("\033[37m", DEFAULT.end)
ERROR = LIGHT_RED
IMPORTANT = LIGHT_RED
COLORS = [GREEN, BLUE, MAGENTA, CYAN, RED, GREY, YELLOW]


class ConsoleOutput(object):
    def __init__(self, verbose=True, colorize=True):
        self.verbose = verbose
        self.colorize = colorize

    def print_banner(self):
        print(" _            _      _      _           \n" +
              "| |_ ___  ___| |_ __| |_ __(_)_   _____ \n" +
              "| __/ _ \\/ __| __/ _` | '__| \\ \\ / / _ \\\n" +
              "| ||  __/\\__ \\ || (_| | |  | |\\ V /  __/\n" +
              " \\__\\___||___/\\__\\__,_|_|  |_| \\_/ \\___|\n" +
              "                                        ", flush=True)

    def print_verbose(self, s, *args, trim=True, color=None):
        if self.verbose:
            self.print(s, *args, trim=trim, color=color)

    def print_error(self, s, *args, trim=True, color=ERROR):
        self.print(s, *args, trim=trim, color=color)

    def print_important(self, s, *args, trim=True, color=IMPORTANT):
        self.print(s, *args, trim=trim, color=color)

    def print(self, s, *args, trim=True, color=DEFAULT):
        defaultColor = DEFAULT
        if type(color) == int:
            color = COLORS[color % len(COLORS)]
        color = color or DEFAULT
        if not self.colorize:
            color = DONT_COLORIZE
            defaultColor = DONT_COLORIZE

        ts = time.time()
        formattedTs = datetime.datetime.fromtimestamp(ts).strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]
        print((defaultColor.start + formattedTs + defaultColor.end + " " + color.start + (
            s.strip() if trim else s) + color.end).format(*args), flush=True)


console_output = ConsoleOutput()
