# bc4 - Basecamp 4 Command Line Interface

bc4 is a powerful command-line interface for Basecamp 4, strongly inspired by GitHub CLI. It's written in Go and provides OAuth2 authentication, multi-account support, project management, todo management with Markdown support, messaging, campfire integration, and card management.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

- Bootstrap, build, and test the repository:
  - `go mod download` -- downloads dependencies in < 1 second
  - `go test -v ./...` -- runs test suite, takes ~30 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
  - `make build` -- builds the binary, takes < 1 second
  - `./build/bc4 --version` -- test the built binary
  - `./build/bc4 --help` -- verify help system works

- Linting and code quality:
  - `go fmt ./...` -- format code, takes < 1 second
  - `go vet ./...` -- run static analysis, takes ~4 seconds
  - **NOTE**: golangci-lint has version compatibility issues in some environments. Always use `go fmt` and `go vet` instead.

- Integration testing:
  - `go test -tags=integration -v ./test/integration` -- runs integration tests (skips without auth tokens)
  - Integration tests require `BC4_TEST_ACCOUNT_ID` and `BC4_TEST_ACCESS_TOKEN` environment variables

## Validation

- **CRITICAL**: Always manually validate any new code by building and testing core functionality.
- Build the application: `make build` (< 1 second)
- Test basic commands:
  - `./build/bc4 --version` -- should show version info
  - `./build/bc4 --help` -- should show comprehensive help
  - `./build/bc4 project list` -- should show authentication error with helpful message
  - `./build/bc4 auth status` -- should show authentication status

- **VALIDATION SCENARIOS**: Test complete user workflows after making changes:
  - **Help system**: Run `./build/bc4 --help` and verify all commands are listed
  - **Version display**: Run `./build/bc4 --version` and `./build/bc4 version` 
  - **Authentication flow**: Run `./build/bc4 auth status` to verify graceful error handling
  - **Command structure**: Test subcommands like `./build/bc4 project --help`, `./build/bc4 todo --help`

- Always run `go fmt ./...` and `go vet ./...` before committing changes or the CI (.github/workflows/test.yml, .github/workflows/lint.yml) will fail.

## Common Tasks

The following are outputs from frequently run commands. Reference them instead of viewing, searching, or running bash commands to save time.

### Repository root
```
ls -la
.env.example
.git
.github
.gitignore
.golangci.yml
.goreleaser.yml
CHANGELOG.md
Formula
LICENSE
Makefile
README.md
RELEASING.md
SPEC.md
cmd
docs
examples
go.mod
go.sum
internal
main.go
test
```

### Go module info
```
cat go.mod
module github.com/needmore/bc4

go 1.23.0

toolchain go1.24.5
```

### Build targets
```
make help
Available targets:
  make build       - Build for current platform
  make build-all   - Build for all supported platforms  
  make test        - Run tests
  make fmt         - Format code
  make vet         - Run go vet
  make lint        - Run linters (requires golangci-lint)
  make install     - Install to GOPATH/bin
  make clean       - Clean build artifacts
  make version     - Show version info
  make help        - Show this help message
```

### Project structure
- `cmd/` - CLI command implementations (account, auth, campfire, card, message, project, todo)
- `internal/` - Internal packages (api, auth, config, errors, factory, markdown, models, parser, tableprinter, tui, ui, utils, version)
- `test/integration/` - Integration tests requiring Basecamp API credentials
- `examples/` - Example files and documentation
- `docs/` - Documentation
- `.github/workflows/` - CI/CD workflows (test.yml, lint.yml, release.yml)

## Build and Test Timing

**CRITICAL - NEVER CANCEL**: These are the expected times for common operations:

- `go mod download` -- < 1 second (0.01s measured)
- `go test -v ./...` -- ~2-30 seconds (faster with cache). NEVER CANCEL. Set timeout to 60+ seconds.
- `make build` -- < 1 second (0.7s measured)
- `go fmt ./...` -- < 1 second (0.1s measured)
- `go vet ./...` -- < 1 second (0.4s measured)
- `make test` -- < 1 second (0.5s with cache). NEVER CANCEL. Set timeout to 60+ seconds.
- `make fmt` -- < 1 second (0.1s measured)
- `make vet` -- < 1 second (0.4s measured)
- `make clean` -- < 1 second

## Technology Stack

- **Language**: Go 1.23+ (toolchain 1.24.5)
- **CLI Framework**: Cobra + Viper for command-line interface
- **Terminal UI**: Charm tools (Bubbletea, Bubbles, Lipgloss, Glamour) for interactive interfaces
- **HTTP Client**: Standard library with OAuth2 support
- **Authentication**: golang.org/x/oauth2 for Basecamp API
- **Markdown Processing**: goldmark for Markdown to HTML conversion
- **Build System**: Make + GoReleaser for cross-platform builds

## Key Commands and Usage

- **Authentication**: `bc4 auth login`, `bc4 auth status`
- **Account Management**: `bc4 account list`, `bc4 account select`
- **Project Management**: `bc4 project list`, `bc4 project view [ID]`, `bc4 project select`
- **Todo Management**: `bc4 todo list`, `bc4 todo add`, `bc4 todo check/uncheck [ID]`
- **Messaging**: `bc4 message list`, `bc4 message post`
- **Campfire**: `bc4 campfire post`, `bc4 campfire view`
- **Cards**: `bc4 card list`, `bc4 card view [ID]`, `bc4 card move [ID]`

## Authentication and Configuration

- Configuration stored in `~/.config/bc4/auth.json` (OAuth tokens) and `~/.config/bc4/config.json` (settings)
- Requires Basecamp OAuth app setup with redirect URI: `http://localhost:8888/callback`
- Supports multiple Basecamp accounts
- Environment variables: `BC4_CLIENT_ID`, `BC4_CLIENT_SECRET`

## Common Issues

- **golangci-lint compatibility**: May fail with "unsupported version" error in some environments. Always use `go fmt` and `go vet` instead.
- **Authentication required**: Most commands require valid Basecamp OAuth tokens
- **Integration tests**: Skip automatically when `BC4_TEST_ACCOUNT_ID` and `BC4_TEST_ACCESS_TOKEN` are not set

## CI/CD Workflows

- **test.yml**: Runs tests, vet, build, and binary validation
- **lint.yml**: Runs golangci-lint with timeout settings
- **release.yml**: Uses GoReleaser for cross-platform builds and GitHub releases

## Development Notes

- Uses GitHub CLI patterns and design principles
- Supports Basecamp URLs as direct command arguments
- Markdown support for todos and messages (converted to Basecamp rich text)
- Cross-platform builds for macOS (Intel/ARM), Linux, Windows
- Comprehensive test suite with unit and integration tests