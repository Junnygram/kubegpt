.PHONY: build install clean test

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=kubegpt
BINARY_UNIX=$(BINARY_NAME)_unix
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/yourusername/kubegpt/cmd.Version=$(VERSION) -X github.com/yourusername/kubegpt/cmd.BuildDate=$(BUILD_DATE) -X github.com/yourusername/kubegpt/cmd.GitCommit=$(GIT_COMMIT)"

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)_mac -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME).exe -v

build-all: build-linux build-mac build-windows

# Install the binary to GOPATH/bin
install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Dependencies
deps:
	$(GOGET) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Mock mode for development without Amazon Q
mock:
	KUBEGPT_MOCK_AI=true $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	KUBEGPT_MOCK_AI=true ./$(BINARY_NAME)