# Contributing to xsh

## Development

### Prerequisites

- Go 1.23 or later
- Make (optional)

### Build

```bash
go build -o xsh main.go
```

### Build with version

```bash
go build -ldflags="-s -w -X github.com/benoitpetit/xsh/cmd.Version=1.0.0" -o xsh main.go
```

### Run tests

```bash
go test ./...
```

### Run with verbose output

```bash
./xsh -v feed
```

## Releasing

### Automatic Release via GitHub Actions

1. Push to `prod` branch for continuous builds
2. Create a tag for official releases:

```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
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
