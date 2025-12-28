# Simple Makefile for Atelier Go

# Binary output path
BINARY_NAME=bin/atelier-go

# Version info for local builds
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Build flags
LDFLAGS=-ldflags "-s -w -X atelier-go/internal/cli.Version=$(VERSION)"

.PHONY: all build clean test run help

all: build

help: ## Show this help message
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary into bin/
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/atelier-go

test: ## Run tests
	$(GOTEST) -v ./...

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f atelier-go

run: build ## Run the application
	./$(BINARY_NAME)

