.PHONY: build test test-race clean release all install verify-build vet lint fmt fmt-check check

VERSION ?= dev
CGO_ENABLED ?= 0
COMMIT ?= $(shell commit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown); if git diff --quiet --ignore-submodules HEAD -- 2>/dev/null && git diff --cached --quiet --ignore-submodules 2>/dev/null; then echo $$commit; else echo $$commit-dirty; fi)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -ldflags="-s -w -X github.com/benoitpetit/xsh/cmd.Version=$(VERSION) -X github.com/benoitpetit/xsh/cmd.Commit=$(COMMIT) -X github.com/benoitpetit/xsh/cmd.BuildDate=$(BUILD_DATE)"
GO_BUILD_FLAGS = -trimpath -buildvcs=false

BINARY_NAME = xsh
MAIN_FILE = main.go

# Default build for current platform
build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

# Build all platforms
all: build-linux build-windows build-darwin

build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)

build-windows:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe $(MAIN_FILE)

build-darwin:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)

# Testing
test:
	go test ./...

test-race:
	go test -race ./...

test-v:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Install locally
install:
	go install $(LDFLAGS) .

# Verify the locally built binary exposes the current command tree.
verify-build: build
	./$(BINARY_NAME) version
	./$(BINARY_NAME) user --help >/dev/null
	./$(BINARY_NAME) lists --help >/dev/null

# Development run
dev:
	go run $(LDFLAGS) $(MAIN_FILE)

# Lint (requires staticcheck)
lint:
	staticcheck ./...

# Format code
fmt:
	go fmt ./...

fmt-check:
	@test -z "$$(gofmt -l .)"

# Check for issues
vet:
	go vet ./...

# Full check
check: fmt-check vet test

# Create release directory and build
release: clean
	mkdir -p dist
	$(MAKE) all VERSION=$(VERSION)
	cd dist && for f in *; do case "$$f" in *.exe) zip -q "$$f.zip" "$$f";; *) tar czf "$$f.tar.gz" "$$f";; esac; done
	for f in dist/$(BINARY_NAME)-linux-amd64 dist/$(BINARY_NAME)-linux-arm64 dist/$(BINARY_NAME)-windows-amd64.exe dist/$(BINARY_NAME)-windows-arm64.exe dist/$(BINARY_NAME)-darwin-amd64 dist/$(BINARY_NAME)-darwin-arm64; do test -f "$$f"; done
