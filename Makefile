.PHONEY: all build test docker

all: test build

build:
	go build -o bin/gatekeeper cmd/gatekeeper/main.go

test:
	go test -v ./...

docker: build
	docker build -t thatjames/gatekeeper .
