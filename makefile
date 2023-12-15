# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean

# Output
BINARY_NAME = 直播间挂机助手
BUILD_BASE_DIR = build

# Builds the project for the current platform
build:
	$(GOBUILD) -o $(BUILD_BASE_DIR)/$(BINARY_NAME)

# Builds the project for multiple platforms
build-all: build-linux build-windows build-darwin

# Builds the project for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_BASE_DIR)/$(BINARY_NAME)-linux-amd64

# Builds the project for Windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_BASE_DIR)/$(BINARY_NAME)-windows-amd64.exe

# Builds the project for macOS (Darwin)
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_BASE_DIR)/$(BINARY_NAME)-darwin-arm64

# Cleans the project
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_BASE_DIR)

# Shows the help message
help:
	@echo "Available targets:"
	@echo "  build           - Builds the project for the current platform"
	@echo "  build-all       - Builds the project for multiple platforms"
	@echo "  build-linux     - Builds the project for Linux"
	@echo "  build-windows   - Builds the project for Windows"
	@echo "  build-darwin    - Builds the project for macOS (Darwin)"
	@echo "  clean           - Cleans the project"
	@echo "  help            - Shows this help message"

.PHONY: build build-all build-linux build-windows build-darwin clean help