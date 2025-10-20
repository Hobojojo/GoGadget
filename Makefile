# TUI App Launcher Makefile

# Variables
BINARY_NAME=tui-launcher
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for current platform
.PHONY: build-local
build-local:
	@echo "Building $(BINARY_NAME) for local platform..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v ./internal/integration/...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install the application (requires sudo)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the application (requires sudo)
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete"

# Install to user's local bin directory
.PHONY: install-user
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin/..."
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "User installation complete"
	@echo "Make sure ~/.local/bin is in your PATH"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet instead..."; \
		go vet ./...; \
	fi

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run .

# Development build with race detection
.PHONY: build-dev
build-dev:
	@echo "Building $(BINARY_NAME) with race detection..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev .
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)-dev"

# Cross-compilation targets
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows
	@echo "Cross-compilation complete"

# Check dependencies
.PHONY: deps
deps:
	@echo "Checking dependencies..."
	go mod tidy
	go mod verify

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Help target
.PHONY: help
help:
	@echo "TUI App Launcher - Available Make targets:"
	@echo ""
	@echo "Building:"
	@echo "  build         Build the application"
	@echo "  build-local   Build for current platform (no build directory)"
	@echo "  build-dev     Build with race detection for development"
	@echo "  build-all     Cross-compile for Linux, macOS, and Windows"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run all tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  test-integration Run integration tests only"
	@echo ""
	@echo "Installation:"
	@echo "  install       Install to /usr/local/bin (requires sudo)"
	@echo "  install-user  Install to ~/.local/bin"
	@echo "  uninstall     Remove from /usr/local/bin (requires sudo)"
	@echo ""
	@echo "Development:"
	@echo "  run           Run the application"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps          Check and tidy dependencies"
	@echo "  update-deps   Update all dependencies"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  build-linux   Build for Linux"
	@echo "  build-darwin  Build for macOS"
	@echo "  build-windows Build for Windows"