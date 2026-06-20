# Recur Makefile
# Build and package automation for Unix-based systems

# Build variables
BINARY_NAME=recur
CMD_PATH=cmd/recur/main.go
BUILD_FLAGS=-tags fts5
LDFLAGS=-ldflags="-s -w"

# Version info (can be overridden: make VERSION=1.0.0)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build output directories
BUILD_DIR=build
DIST_DIR=dist

# Platforms for cross-compilation
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all build test test-verbose test-coverage clean install uninstall \
        build-all package package-all dist-clean help version

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build with version info embedded
build-release:
	@echo "Building $(BINARY_NAME) v$(VERSION) (release)..."
	go build $(BUILD_FLAGS) \
		-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BINARY_NAME) $(CMD_PATH)
	@echo "Release build complete: ./$(BINARY_NAME)"

# Build for all platforms
build-all: clean-build
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output="$(BUILD_DIR)/$(BINARY_NAME)-$${GOOS}-$${GOARCH}"; \
		echo "Building for $${GOOS}/$${GOARCH}..."; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} go build $(BUILD_FLAGS) $(LDFLAGS) -o $${output} $(CMD_PATH); \
		if [ $$? -eq 0 ]; then \
			echo "  ✓ $${output}"; \
		else \
			echo "  ✗ Failed to build for $${GOOS}/$${GOARCH}"; \
		fi; \
	done
	@echo "All builds complete!"

# Create distributable packages (tar.gz)
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		binary="$(BUILD_DIR)/$(BINARY_NAME)-$${GOOS}-$${GOARCH}"; \
		if [ -f "$${binary}" ]; then \
			package_name="$(BINARY_NAME)-$(VERSION)-$${GOOS}-$${GOARCH}"; \
			package_dir="$(BUILD_DIR)/$${package_name}"; \
			mkdir -p "$${package_dir}"; \
			cp "$${binary}" "$${package_dir}/$(BINARY_NAME)"; \
			cp README.md "$${package_dir}/" 2>/dev/null || true; \
			cp LICENSE "$${package_dir}/" 2>/dev/null || true; \
			echo "Creating $${package_name}.tar.gz..."; \
			tar -czf "$(DIST_DIR)/$${package_name}.tar.gz" -C "$(BUILD_DIR)" "$${package_name}"; \
			rm -rf "$${package_dir}"; \
			echo "  ✓ $(DIST_DIR)/$${package_name}.tar.gz"; \
		fi; \
	done
	@echo "Packaging complete! Packages in $(DIST_DIR)/"

# Create checksums for all packages
checksums: package
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum *.tar.gz > SHA256SUMS
	@echo "  ✓ $(DIST_DIR)/SHA256SUMS"

# Full distribution build with checksums
package-all: package checksums
	@echo "Distribution complete!"
	@echo ""
	@echo "Packages created:"
	@ls -lh $(DIST_DIR)/*.tar.gz
	@echo ""
	@echo "To verify a package:"
	@echo "  cd $(DIST_DIR) && sha256sum -c SHA256SUMS"

# Install to system (requires sudo on most systems)
install: build-release
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@install -d /usr/local/bin
	@install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete! Run '$(BINARY_NAME) help' to get started."

# Install to user bin (no sudo required)
install-user: build-release
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@install -m 755 $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo "Make sure ~/.local/bin is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

# Uninstall from system
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete."

# Uninstall from user bin
uninstall-user:
	@echo "Removing $(BINARY_NAME) from ~/.local/bin..."
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete."

# Run tests
test:
	@echo "Running tests..."
	cd tests/integration && go test $(BUILD_FLAGS) ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	cd tests/integration && go test -v $(BUILD_FLAGS) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd tests/integration && go test $(BUILD_FLAGS) -coverprofile=coverage.out ./...
	@echo "Generating coverage report..."
	go tool cover -html=tests/integration/coverage.out

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	cd tests/integration && go test $(BUILD_FLAGS) -race ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f tests/integration/coverage.out
	@echo "Clean complete."

# Clean build directory
clean-build:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Clean distribution directory
clean-dist:
	@echo "Cleaning distribution directory..."
	@rm -rf $(DIST_DIR)

# Clean everything
clean-all: clean clean-build clean-dist
	@echo "All artifacts cleaned."

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all quality checks
check: fmt vet test
	@echo "All checks passed!"

# Show version info
version:
	@echo "Recur Build Information"
	@echo "======================="
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(shell go version)"

# Development build with live reload (requires air: go install github.com/air-verse/air@latest)
dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install it with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

# Help target
help:
	@echo "Recur Makefile"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              Build binary for current platform"
	@echo "  make build-release      Build release binary with version info"
	@echo "  make build-all          Build for all supported platforms"
	@echo "  make install            Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user       Install to ~/.local/bin (no sudo)"
	@echo "  make uninstall          Remove from /usr/local/bin"
	@echo "  make uninstall-user     Remove from ~/.local/bin"
	@echo ""
	@echo "Package Commands:"
	@echo "  make package            Create tar.gz packages for all platforms"
	@echo "  make checksums          Generate SHA256 checksums"
	@echo "  make package-all        Create packages with checksums"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               Run tests"
	@echo "  make test-verbose       Run tests with verbose output"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make test-race          Run tests with race detection"
	@echo ""
	@echo "Quality Commands:"
	@echo "  make fmt                Format code"
	@echo "  make vet                Vet code"
	@echo "  make lint               Lint code (requires golangci-lint)"
	@echo "  make check              Run all quality checks"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make clean              Clean build artifacts"
	@echo "  make clean-all          Clean all artifacts and directories"
	@echo "  make version            Show version information"
	@echo "  make dev                Run with live reload (requires air)"
	@echo "  make help               Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make VERSION=1.0.0 package-all    Create v1.0.0 release packages"
	@echo "  make install-user                  Install for current user only"
	@echo "  make test && make build            Test then build"
