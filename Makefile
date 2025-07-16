# bc4 Makefile

# Get version from git tag or use dev
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/needmore/bc4/internal/version.Version=$(VERSION) \
	-X github.com/needmore/bc4/internal/version.GitCommit=$(COMMIT) \
	-X github.com/needmore/bc4/internal/version.BuildDate=$(BUILD_DATE)"

# Binary name
BINARY := bc4

# Build directories
BUILD_DIR := ./build
DIST_DIR := ./dist

.PHONY: all build build-all clean test fmt vet lint install version help

# Default target
all: clean fmt vet test build

# Build for current platform
build:
	@echo "Building $(BINARY) $(VERSION) for current platform..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .

# Build for all supported platforms
build-all: clean
	@echo "Building $(BINARY) $(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# macOS Intel
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-amd64 .
	
	# macOS Apple Silicon
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY)-darwin-arm64 .
	
	# Create universal binary for macOS
	@echo "Creating universal binary..."
	@lipo -create -output $(DIST_DIR)/$(BINARY)-darwin-universal \
		$(DIST_DIR)/$(BINARY)-darwin-amd64 \
		$(DIST_DIR)/$(BINARY)-darwin-arm64

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run linting (requires golangci-lint)
lint:
	@echo "Running linters..."
	@golangci-lint run

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY) to GOPATH/bin..."
	@go install $(LDFLAGS) .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@go clean

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build       - Build for current platform"
	@echo "  make build-all   - Build for all supported platforms"
	@echo "  make test        - Run tests"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run go vet"
	@echo "  make lint        - Run linters (requires golangci-lint)"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make version     - Show version info"
	@echo "  make help        - Show this help message"