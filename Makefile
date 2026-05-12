# Infra CLI Makefile
BINARY_NAME=infra
VERSION?=0.1.0
BUILD_DIR=bin
MAIN_PATH=./main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X github.com/duclm2609/infra-cli/cmd/infra.Version=$(VERSION)"

.PHONY: all build build-all test clean deps lint

all: clean deps build

# Build for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for all platforms - each platform gets its own folder with binary named "infra"
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-amd64:
	mkdir -p $(BUILD_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/darwin-amd64/$(BINARY_NAME) $(MAIN_PATH)

build-darwin-arm64:
	mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/darwin-arm64/$(BINARY_NAME) $(MAIN_PATH)

build-linux-amd64:
	mkdir -p $(BUILD_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) $(MAIN_PATH)

build-linux-arm64:
	mkdir -p $(BUILD_DIR)/linux-arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/linux-arm64/$(BINARY_NAME) $(MAIN_PATH)

build-windows-amd64:
	mkdir -p $(BUILD_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/windows-amd64/$(BINARY_NAME).exe $(MAIN_PATH)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Install binary locally
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
