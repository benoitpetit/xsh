# Contributing to xsh

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional)

### Build

```bash
go build -o xsh main.go
```

### Build with version

```bash
make build VERSION=v0.1.0
```

### Run tests

```bash
testhome=$(mktemp -d)
env -u XSH_CONFIG_DIR -u XDG_CONFIG_HOME HOME="$testhome" go test ./...
rm -rf "$testhome"
go test -race ./...
go vet ./...
```

For browser compatibility, also run `CGO_ENABLED=0 go test ./browser`. The
release artifacts use the CGO-free path for all six advertised targets.

Fuzz smoke tests are bounded and offline:

```bash
go test ./core -run=^$ -fuzz=Fuzz -fuzztime=5s
go test ./browser -run=^$ -fuzz=Fuzz -fuzztime=5s
go test ./cmd -run=^$ -fuzz=Fuzz -fuzztime=5s
```

Keep `XSH_CONFIG_DIR` and `HOME` isolated for tests. Never place real cookies,
tokens, or endpoint cache files in fixtures.

### Run with verbose output

```bash
./xsh -v feed
```

## Releasing

### Automatic Release via GitHub Actions

1. Merge the reviewed changes into `master`
2. Create the GitHub release with `gh`:

```bash
gh release create v0.1.0 --target master --generate-notes
```

The GitHub Actions workflow will automatically:
- Build binaries for Linux, Windows, and macOS (amd64 & arm64)
- Create a GitHub Release with all binaries
- Compress binaries (.tar.gz for Unix, .zip for Windows)

### Manual Release

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o xsh-linux-amd64 main.go

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o xsh-windows-amd64.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o xsh-darwin-amd64 main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o xsh-darwin-arm64 main.go
```

The reproducible release path is `make release VERSION=0.1.0`. It builds
Linux/Windows/macOS amd64 and arm64 artifacts with `-trimpath -buildvcs=false`
and `CGO_ENABLED=0`, then verifies every binary and archive exists.

## Project Structure

```
.
├── browser/       # Browser cookie extraction
├── cmd/           # CLI commands
├── core/          # Core API logic
├── display/       # Output formatting
├── models/        # Data models
├── tests/         # Test suite
├── utils/         # Utilities
└── main.go        # Entry point
```

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Run `go vet` to check for issues
- Add tests for new features
