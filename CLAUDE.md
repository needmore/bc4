## Project Overview

bc4 is a command-line interface for Basecamp 4, written in Go and inspired by GitHub CLI. It provides OAuth2 authentication, multi-account support, project/todo/message/campfire/card management with Markdown support

- It implements the Basecamp 3/4 API at https://github.com/basecamp/bc3-api
- It is inspired by, and should behave similarly to, GitHub CLI: https://github.com/cli/cli

## Build and Test Commands

```bash
# Build
make build                    # Build binary to ./build/bc4 (<1s)
./build/bc4 --version        # Verify build

# Test
go test -v ./...             # Run all tests (2-30s, NEVER CANCEL - set 60s+ timeout)
go test -v ./internal/api    # Run specific package tests
go test -tags=integration -v ./test/integration  # Integration tests (requires auth tokens)

# Lint (use these instead of golangci-lint due to version compatibility issues)
go fmt ./...                 # Format code
go vet ./...                 # Static analysis
```

Integration tests require `BC4_TEST_ACCOUNT_ID` and `BC4_TEST_ACCESS_TOKEN` environment variables.

## Architecture

### Directory Structure
- `cmd/` - CLI commands (Cobra-based): account, activity, auth, campfire, card, checkin, comment, document, message, people, profile, project, schedule, search, todo
- `internal/` - Private packages:
  - `api/` - Modular Basecamp API client with retry logic and rate limiting
  - `auth/` - OAuth2 authentication
  - `config/` - JSON config at `~/.config/bc4/` (falls back to `~/Library/Application Support/bc4/` on macOS for existing users)
  - `errors/` - Custom error types with user-friendly messages
  - `factory/` - Dependency injection (used by all commands)
  - `markdown/` - Markdown to Basecamp HTML conversion (goldmark)
  - `models/` - Data models
  - `parser/` - Basecamp URL parsing (extracts account/project/resource IDs)
  - `tableprinter/` - GitHub CLI-style table rendering with TTY detection
  - `tui/` - Terminal UI components (Bubbletea/Lipgloss)

### Key Patterns

**Factory Pattern**: All commands receive `*factory.Factory` for lazy-initialized API client, auth, and config access.

**Command Structure**: Each command follows:
```go
func NewXxxCmd(f *factory.Factory) *cobra.Command
```

**API Client**: Two-layer design in `internal/api/`:
- `client.go` - Low-level HTTP with retry/backoff
- `modular.go` - High-level operations grouped by resource (Projects, Todos, Messages, etc.)

**URL Support**: Commands accept Basecamp URLs directly as arguments. The parser extracts IDs automatically.

**Table Rendering**: `internal/tableprinter/` adapts output for TTY (colors, relative times) vs non-TTY (plain text, RFC3339 timestamps).

## Validation After Changes

```bash
make build
./build/bc4 --help           # Verify all commands listed
./build/bc4 --version        # Verify version display
./build/bc4 auth status      # Verify graceful error handling
./build/bc4 project --help   # Verify subcommand help
```

## Technology Stack

- Go 1.23+ (toolchain 1.24.5)
- CLI: Cobra + Viper
- Terminal UI: Charm tools (Bubbletea, Bubbles, Lipgloss, Glamour)
- Auth: golang.org/x/oauth2
- Markdown: goldmark, html-to-markdown/v2
- Build: Make + GoReleaser

## CI/CD

- `test.yml` - Tests, vet, build on every push/PR
- `lint.yml` - golangci-lint
- `release.yml` - GoReleaser on version tags (v*)

Always run `go fmt ./...` and `go vet ./...` before committing.
