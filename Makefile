.PHONY: test lint build build-image build-image-multiplatform push push-latest builder clean install test-build help

BINARY ?= artship
REGISTRY ?= ghcr.io/ipaqsa
IMAGE ?= artship:latest
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
PLATFORMS ?= linux/amd64,linux/arm64
LDFLAGS := -X artship/internal/version.Version=$(VERSION) -s -w

## Run tests
test:
	go test -v ./...

## Lint the codebase
lint:
	golangci-lint run

## Fix the codebase against lint
lint-fix:
	golang-lint run --fix

## Format code
fmt:
	go fmt ./...

## Run static analysis
vet:
	go vet ./...

## Build the CLI binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/artship

## Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-amd64 ./cmd/artship
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm64 ./cmd/artship
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY)-darwin-arm64 ./cmd/artship

## Install binary to local system
install: build
	sudo mv $(BINARY) /usr/local/bin/

## Build multi-platform container images
build-images: builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--tag $(REGISTRY)/$(IMAGE) \
		--tag $(REGISTRY)/$(BINARY):$(VERSION) \
		.

## Push multi-platform container images to registry
push: builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--tag $(REGISTRY)/$(IMAGE) \
		--tag $(REGISTRY)/$(BINARY):$(VERSION) \
		--push \
		.

## Build and push latest tag only (for development)
push-latest: builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--tag $(REGISTRY)/$(IMAGE) \
		--push \
		.

## Create or use multi-platform builder
builder:
	@if ! docker buildx ls | grep -q multiplatform-builder; then \
		echo "Creating new multiplatform builder..."; \
		docker buildx create --name multiplatform-builder --driver docker-container --use; \
	else \
		echo "Using existing multiplatform builder..."; \
		docker buildx use multiplatform-builder; \
	fi

## Test that the binary builds and works correctly
test-build: build
	./$(BINARY) version
	@echo "✅ Binary builds and runs successfully"

## Clean build artifacts
clean:
	rm -f $(BINARY) $(BINARY)-*

## Show help
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## /  /'