TOP_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

build:	test
	@cd ${TOP_DIR} && \
	mkdir -p dist/ && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a \
			-ldflags "-extldflags \"-static\"" -o dist/testdrive ./cmd/testdrive.go

clean:
	rm -rf ${TOP_DIR}/dist/ ${TOP_DIR}/.cache/

format:
	@cd ${TOP_DIR} && \
	gofmt -s -w .

test:
	@cd ${TOP_DIR} && \
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic ./cmd/... ./internal/...

lint:
	@cd ${TOP_DIR} && \
	go get -u golang.org/x/lint/golint && \
	${GOPATH}/bin/golint ./internal/... ./cmd/...

refresh:
	@cd ${TOP_DIR} && \
	go clean -modcache && go mod tidy 
