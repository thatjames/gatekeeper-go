.PHONY: all build test docker version
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

all: web test build

build: test web
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper cmd/gatekeeper/main.go

test:
	go test -v ./...

web:
	$(MAKE) -C internal/web/ui build

version:
	@echo "Version: $(VERSION)"

build-docker: web test
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper cmd/gatekeeper/main.go


docker: build-docker
	docker build -t thatjames/gatekeeper .

docker-run: 
	docker run --name gatekeeper --rm -ti -v $(PWD)/config/docker-config.yml:/app/config.yml -p 8085:8085 thatjames/gatekeeper -debug

install:
	$(MAKE) -C internal/web/ui install
