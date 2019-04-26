#!/bin/bash

PROJECT_DIR=$(cd `dirname $0`/.. && pwd)

cd ${PROJECT_DIR}

if [ ! -d "${PROJECT_DIR}/.venv/3.7/" ]; then
  virtualenv -p python3.7 ${PROJECT_DIR}/.venv/3.7/
fi

source ${PROJECT_DIR}/.venv/3.7/bin/activate

${VIRTUAL_ENV}/bin/pip3 install -r requirements.txt -r requirements-build.txt -r requirements-dev.txt

echo "Activate with 

source ${PROJECT_DIR}/.venv/3.7/bin/activate

"
