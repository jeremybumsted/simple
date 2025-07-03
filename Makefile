# Simple CLI Makefile

# Variables
BINARY_NAME=simple
MAIN_PATH=.
BUILD_DIR=build
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} ${MAIN_PATH}

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ${MAIN_PATH}
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ${MAIN_PATH}
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 ${MAIN_PATH}
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ${MAIN_PATH}

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -v -cover ./...
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linting
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Vet code
.PHONY: vet
vet:
	go vet ./...

# Run all checks
.PHONY: check
check: fmt vet lint test

# Clean build artifacts
.PHONY: clean
clean:
	rm -f ${BINARY_NAME}
	rm -rf ${BUILD_DIR}
	rm -f coverage.out coverage.html

# Install the binary
.PHONY: install
install:
	go install ${LDFLAGS} ${MAIN_PATH}

# Run the application
.PHONY: run
run:
	go run ${MAIN_PATH}

# Run with TUI
.PHONY: run-tui
run-tui:
	go run ${MAIN_PATH} tui

# Create configuration file
.PHONY: configure
configure:
	@echo "Creating configuration file..."
	./simple configure
	@echo "Edit ~/.simple/config.yaml and set your PLAIN_API_KEY"

# Generate example config
.PHONY: config-example
config-example:
	@echo "Creating example config..."
	@mkdir -p ~/.simple
	@cp config.example.yaml ~/.simple/config.example.yaml
	@echo "Example config created at ~/.simple/config.example.yaml"

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Development environment setup complete!"
	@echo "1. Run 'make configure' to create a configuration file"
	@echo "2. Set your PLAIN_API_KEY environment variable"
	@echo "3. Run 'make run' to start the application"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         Build the binary"
	@echo "  build-all     Build for multiple platforms"
	@echo "  deps          Install dependencies"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  lint          Run linting"
	@echo "  fmt           Format code"
	@echo "  vet           Vet code"
	@echo "  check         Run all checks (fmt, vet, lint, test)"
	@echo "  clean         Clean build artifacts"
	@echo "  install       Install the binary"
	@echo "  run           Run the application"
	@echo "  run-tui       Run with TUI"
	@echo "  configure     Create configuration file"
	@echo "  config-example Generate example config"
	@echo "  dev-setup     Setup development environment"
	@echo "  help          Show this help"
