PROJECT_DIR := $(shell pwd)

all:	test

test:	init
	@${PROJECT_DIR}/bin/testdrive --verbose

init:
	@[ -d "${PROJECT_DIR}/.venv/" ] || ${PROJECT_DIR}/scripts/setup-dev.sh
