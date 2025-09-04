.PHONY: all build test docker version
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

all: test build

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper cmd/gatekeeper/main.go

test:
	go test -v ./...

version:
	@echo "Version: $(VERSION)"

build-docker: test
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper cmd/gatekeeper/main.go


docker: build-docker
	docker build -t thatjames/gatekeeper .

