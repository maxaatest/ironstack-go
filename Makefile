# IronStack WP - Makefile
# Build for multiple platforms

BINARY_NAME=ironstack
VERSION=1.0.0
BUILD_DIR=build
MAIN_PATH=./cmd/ironstack

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build linux darwin windows clean install release

all: clean build

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Build for Linux (production target)
linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Built Linux binaries in $(BUILD_DIR)/"

# Build for macOS
darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Built macOS binaries in $(BUILD_DIR)/"

# Build for Windows
windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Built Windows binary in $(BUILD_DIR)/"

# Build all platforms
release: clean linux darwin windows
	@echo "Release builds complete!"
	@ls -lh $(BUILD_DIR)/

# Install locally
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Run tests
test:
	go test -v ./...

# Run with race detector
test-race:
	go test -race -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	@echo "Generating documentation..."
	go doc -all ./... > docs/API.md

# Development run
dev:
	go run $(MAIN_PATH)

# Docker build
docker:
	docker build -t ironstack:$(VERSION) .

# Create release tarball
tarball: linux
	cd $(BUILD_DIR) && tar -czvf ironstack-$(VERSION)-linux-amd64.tar.gz ironstack-linux-amd64
	cd $(BUILD_DIR) && tar -czvf ironstack-$(VERSION)-linux-arm64.tar.gz ironstack-linux-arm64
	@echo "Created release tarballs"

# Show help
help:
	@echo "IronStack WP - Build Commands"
	@echo ""
	@echo "  make build    - Build for current platform"
	@echo "  make linux    - Build for Linux (amd64 + arm64)"
	@echo "  make darwin   - Build for macOS (amd64 + arm64)"
	@echo "  make windows  - Build for Windows"
	@echo "  make release  - Build all platforms"
	@echo "  make install  - Install to /usr/local/bin"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make dev      - Run in development mode"
