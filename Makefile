.PHONY: build test lint install clean help release-snapshot release-check

BINARY_NAME=secctl
INSTALL_DIR=$(GOPATH)/bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the secctl binary
	go build -v -o $(BINARY_NAME) .

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...

coverage: test ## Run tests and show coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-func: test ## Show coverage by function
	go tool cover -func=coverage.out

lint: ## Run linters
	go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

fmt: ## Format code
	go fmt ./...

install: build ## Build and install the binary to $(INSTALL_DIR)
	go install .

clean: ## Remove build artifacts and coverage files
	rm -f $(BINARY_NAME) coverage.out coverage.html

release-snapshot: ## Test goreleaser with a snapshot build
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed. Install from https://goreleaser.com/install/"; exit 1; }
	goreleaser release --snapshot --clean --skip=publish

release-check: ## Check goreleaser configuration
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed. Install from https://goreleaser.com/install/"; exit 1; }
	goreleaser check

all: clean fmt lint test build ## Run all checks and build

.DEFAULT_GOAL := help
