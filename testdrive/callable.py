class Callable:
    def __init__(self, function, *args):
        self.args = args
        self.function = function

    def __call__(self, *args, **kwargs):
        return self.function(*self.args)
