#!/bin/bash

PROJECT_DIR=$(cd `dirname $0`/.. && pwd)

cd ${PROJECT_DIR}

${VIRTUAL_ENV}/bin/pyinstaller --exclude-module pycrypto --exclude-module PyInstaller testdrive.spec
