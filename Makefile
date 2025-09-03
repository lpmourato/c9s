# Makefile for c9s - Cloud Run Status Monitor

# Variables
BINARY_NAME=c9s
BUILD_DIR=bin
CMD_DIR=cmd/c9s
MAIN_PATH=./$(CMD_DIR)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty)"

.PHONY: all build build-arm64 clean test deps run install dev lint fmt vet

# Default target
all: clean deps test build

# Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for ARM64 macOS (Apple Silicon)
build-arm64:
	mkdir -p $(BUILD_DIR)
	GOARCH=arm64 GOOS=darwin $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all: clean deps
	mkdir -p $(BUILD_DIR)
	# Linux amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# macOS amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application in development mode
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) gcp --project=${GOOGLE_CLOUD_PROJECT}

# Run with mock data
run-mock: build
	./$(BUILD_DIR)/$(BINARY_NAME) mock

# Install the binary to GOPATH/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)

# Development mode with file watching (requires entr)
dev:
	find . -name "*.go" | entr -r make run

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	$(GOCMD) fmt ./...

# Vet the code
vet:
	$(GOCMD) vet ./...

# Generate code (if needed)
generate:
	$(GOCMD) generate ./...

# Security scan
security:
	gosec ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-arm64  - Build for ARM64 macOS (Apple Silicon)"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  run          - Build and run with GCP (requires GOOGLE_CLOUD_PROJECT)"
	@echo "  run-mock     - Build and run with mock data"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  dev          - Development mode with file watching"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  generate     - Generate code"
	@echo "  security     - Run security scan"
	@echo "  help         - Show this help"
