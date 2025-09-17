.PHONY: all build test docker version clean-binaries build-all-arch docker-multiarch
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BIN_NAME ?= gatekeeper

# Default architecture values
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Set GOARM for ARM builds
ifeq ($(GOARCH),arm)
	GOARM ?= 7
	export GOARM
endif

all: web test build

##@ Build
build: test web ## Builds the native go binary
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BIN_NAME) -tags live ./cmd/gatekeeper/

build-pipeline: ## Builds the binary for the pipeline (uses GOOS, GOARCH, BIN_NAME from environment)
	@echo "Building $(BIN_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p bin
	env CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BIN_NAME) -tags live ./cmd/gatekeeper/

web: install ## Builds the web ui
	$(MAKE) -C internal/web/ui build

docker-binary: web test ## Builds the docker binary
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper cmd/gatekeeper/main.go

docker-binary-multiarch: web test ## Builds docker binaries for all supported architectures
	@echo "Building binaries for multi-arch docker..."
	@mkdir -p bin
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper-linux-amd64 ./cmd/gatekeeper/
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper-linux-arm64 ./cmd/gatekeeper/
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper-linux-arm ./cmd/gatekeeper/
	env CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gatekeeper-linux-386 ./cmd/gatekeeper/

docker: docker-binary ## Builds the docker binary, the web ui and the docker image
	docker build -t smokeycircles/gatekeeper .

docker-multiarch: docker-binary-multiarch ## Builds and pushes multi-architecture docker images
	@echo "Building multi-architecture docker images..."
	@echo "Make sure docker buildx is enabled and you're logged into Docker Hub"
	docker buildx create --use --name multiarch-builder --driver docker-container 2>/dev/null || docker buildx use multiarch-builder
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7,linux/386 --push -t smokeycircles/gatekeeper:$(VERSION) -t smokeycircles/gatekeeper:latest -f Dockerfile.multiarch .

##@ Test
test: generate-mocks ## Runs the golang unit tests
	go test ./internal/dhcp -run TestFeatures -v
	go test ./internal/dns -run TestFeatures -v
	go test ./internal/dhcp/packet_test.go internal/dhcp/packet.go

test-report: generate-mocks ## Runs the golang unit tests and generates a test report
	@go test ./internal/dhcp -run TestFeaturesWithOutputFile -report-file=${PWD}/dhcp-test-report.xml
	@go test ./internal/dns -run TestFeaturesWithOutputFile -report-file=${PWD}/dns-test-report.xml

##@ Run
docker-run: ## Runs the docker image
	docker run --name gatekeeper --rm -ti -v $(PWD)/config/docker-config.yml:/app/config.yml -v /tmp/leases:/var/lib/gatekeeper -p 8085:8085 smokeycircles/gatekeeper -debug

##@ Web
install: ## Installs the web ui
	$(MAKE) -C internal/web/ui install

##@ Misc
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

version: ## Prints the build version
	@echo "Version: $(VERSION)"

clean: ## Cleans the build directories
	rm -rf bin
	rm -rf internal/datasource/mocks
	$(MAKE) -C internal/web/ui clean

clean-binaries: ## Cleans only the binary directory
	rm -rf bin

generate-mocks: install-mockgen ## Generates the test DB mocks
	@rm -rf internal/datasource/mocks
	@go generate ./...
	@go mod tidy

install-mockgen: ## Installs the mockgen tool if it's missing
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		go install github.com/golang/mock/mockgen@latest; \
	fi
