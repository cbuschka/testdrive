import datetime
import time


class ConsoleOutput(object):
    def print_banner(self):
        print(" _            _      _      _           \n" +
              "| |_ ___  ___| |_ __| |_ __(_)_   _____ \n" +
              "| __/ _ \\/ __| __/ _` | '__| \\ \\ / / _ \\\n" +
              "| ||  __/\\__ \\ || (_| | |  | |\\ V /  __/\n" +
              " \\__\\___||___/\\__\\__,_|_|  |_| \\_/ \\___|\n" +
              "                                        ", flush=True)

    def print(self, s, *args, trim=True):
        ts = time.time()
        formattedTs = datetime.datetime.fromtimestamp(ts).strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]
        print((formattedTs + " " + s.strip() if trim else s).format(*args), flush=True)


console_output = ConsoleOutput()
