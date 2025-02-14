# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=bidprentjes-api

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build clean test coverage fmt lint run deps help

all: clean build test ## Build and run tests

build: ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) -v

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out

test: ## Run tests
	$(GOTEST) -v ./...

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

fmt: ## Format code
	$(GOFMT) ./...

lint: ## Run linter
	golangci-lint run

run: ## Run the application
	$(GORUN) main.go

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help 