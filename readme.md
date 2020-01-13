# Testdrive

### Multi-docker-container-orchestration optimized for integration tests

## Setup dev

### Prerequesites
* Linux
* python 3.7 with virtualenv
* GNU make
* docker

### Set up dev env
```
./scripts/setup-dev.sh
```

```
source ./.venv/3.7/bin/active
```

=> Ready to develop.

## Build

(Assuming dev env is set up.)

```
./scripts/build-linux.sh
```

## License
Copyright (c) 2016-2020 by the [Cornelius Buschka](https://github.com/cbuschka).

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a [copy of the GNU General Public License](./license.txt)
along with this program.  If not, see [GNU GPL v3 at http://www.gnu.org](http://www.gnu.org/licenses/gpl-3.0.txt).
