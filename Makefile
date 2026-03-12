.PHONY: build test clean release all

VERSION ?= dev
LDFLAGS = -ldflags="-s -w -X github.com/benoitpetit/xsh/cmd.Version=$(VERSION)"

BINARY_NAME = xsh
MAIN_FILE = main.go

# Default build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

# Build all platforms
all: build-linux build-windows build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe $(MAIN_FILE)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)

# Testing
test:
	go test ./...

test-v:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Install locally
install:
	go install $(LDFLAGS) .

# Development run
dev:
	go run $(LDFLAGS) $(MAIN_FILE)

# Lint (requires staticcheck)
lint:
	staticcheck ./...

# Format code
fmt:
	go fmt ./...

# Check for issues
vet:
	go vet ./...

# Full check
check: fmt vet lint test

# Create release directory and build
release: clean
	mkdir -p dist
	$(MAKE) all VERSION=$(VERSION)
	cd dist && for f in *; do if [[ "$$f" == *.exe ]]; then zip "$$f.zip" "$$f"; else tar czf "$$f.tar.gz" "$$f"; fi; done
