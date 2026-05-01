DOCKER_DESKTOP_CLI := /Applications/Docker.app/Contents/Resources/bin/docker
DOCKER_CLI := $(shell command -v docker 2>/dev/null)
CONTAINER_RUNTIME ?= $(if $(DOCKER_CLI),$(DOCKER_CLI),$(if $(wildcard $(DOCKER_DESKTOP_CLI)),$(DOCKER_DESKTOP_CLI),docker))
BINARY_CLI_NAME := issue2md
BINARY_WEB_NAME := issue2mdweb
BUILD_DIR := bin
DOCKER_IMAGE_NAME := issue2md
DOCKER_TAG := local

.PHONY: all build test lint format verify clean web docker-build docker-run-cli docker-run-web check-container-runtime dev-setup help

all: build ## Build both binaries

build: ## Build CLI and web binaries into ./bin
	@echo "Building $(BINARY_CLI_NAME) and $(BINARY_WEB_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_CLI_NAME) ./cmd/issue2md
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_WEB_NAME) ./cmd/issue2mdweb
	@echo "Build completed"

test: ## Run all Go tests
	go test ./...

lint: ## Run golangci-lint when available
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found; skipping"; \
	fi

format: ## Format Go code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found; skipping"; \
	fi

verify: format test ## Run formatting and tests

clean: ## Remove local build artifacts
	rm -rf $(BUILD_DIR)

web: ## Build the web entrypoint package
	go build ./cmd/issue2mdweb

check-container-runtime:
	@command -v $(CONTAINER_RUNTIME) >/dev/null 2>&1 || (echo 'container runtime "$(CONTAINER_RUNTIME)" not found; install Docker or run with CONTAINER_RUNTIME=podman' >&2; exit 1)

docker-build: check-container-runtime ## Build the local Docker image
	$(CONTAINER_RUNTIME) build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

docker-run-cli: check-container-runtime ## Run the CLI image; requires URL=... and optional OUTPUT=/work/out.md
	@test -n "$(URL)" || (echo 'usage: make docker-run-cli URL=https://github.com/OWNER/REPO/issues/123 [OUTPUT=/work/out.md]' >&2; exit 1)
	$(CONTAINER_RUNTIME) run --rm -v "$$(pwd):/work" -e GITHUB_TOKEN="$${GITHUB_TOKEN:-}" $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) "$(URL)" $(OUTPUT)

docker-run-web: check-container-runtime ## Run the image in web mode
	$(CONTAINER_RUNTIME) run --rm $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) web

dev-setup: ## Install optional local developer tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_.-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)