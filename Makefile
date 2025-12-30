# Makefile for hashpwd2

# Binary name
BINARY_NAME=hashpwd2

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Version information from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags to inject version information
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build clean test run deps install help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

# Build with optimizations (smaller binary)
build-optimized:
	@echo "Building optimized $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BINARY_NAME) -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run the application
run: build
	./$(BINARY_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install to $GOPATH/bin or $GOBIN
install:
	@echo "Installing $(BINARY_NAME) $(VERSION)..."
	$(GOCMD) install $(LDFLAGS) -v

# Install to system location (requires sudo)
install-system: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo install -m 755 $(BINARY_NAME) /usr/local/bin/

# Display version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build            - Build the binary"
	@echo "  make build-optimized  - Build optimized binary (smaller size)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make test             - Run tests"
	@echo "  make run              - Build and run the application"
	@echo "  make deps             - Download and tidy dependencies"
	@echo "  make install          - Install to GOPATH/bin"
	@echo "  make install-system   - Install to /usr/local/bin (requires sudo)"
	@echo "  make version          - Display version information"
	@echo "  make help             - Show this help message"
