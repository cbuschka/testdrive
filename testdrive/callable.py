class Callable:
    def __init__(self, function, *args, skip=False):
        self.args = args
        self.function = function
        self.skip = skip

    def __call__(self, *args, **kwargs):
        if self.skip:
            return

        return self.function(*self.args)
