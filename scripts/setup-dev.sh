#!/bin/bash

PROJECT_DIR=$(cd `dirname $0`/.. && pwd)

cd ${PROJECT_DIR}

VENV_DIR=${PROJECT_DIR}/.venv/3.7/

if [ ! -d "${VENV_DIR}" ]; then
  virtualenv -p python3.7 ${VENV_DIR}
fi

source ${VENV_DIR}/bin/activate

${VIRTUAL_ENV}/bin/pip3 install -r requirements.txt -r requirements-build.txt -r requirements-dev.txt

echo "Activate with 

source ${VENV_DIR}/bin/activate

export PYTHONPATH=${PROJECT_DIR}

"
