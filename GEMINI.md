## Project Overview

bc4 is a Basecamp 4 CLI tool written in Go, inspired by GitHub's `gh` CLI. It provides terminal access to Basecamp features with beautiful TUIs using Charm tools.

## Development Commands

### Build & Run
```bash
go build -o bc4              # Build binary
go run main.go              # Run without building
go install                  # Install to GOPATH/bin
```

### Testing & Quality
```bash
go test ./...               # Run all tests
go test -v ./internal/api   # Test specific package
go fmt ./...                # Format code
go vet ./...                # Run static analysis
golangci-lint run           # Run comprehensive linting
```

### Development Workflow
1. Make changes to code
2. Run `go fmt ./...` to format
3. Run `go vet ./...` to check for issues
4. Run `go test ./...` to ensure tests pass
5. Build with `go build -o bc4` to test binary

## Architecture

### Command Structure (Cobra)
- Commands are in `cmd/` directory
- Each command file implements a cobra.Command
- Root command is in `cmd/root.go`
- Use interactive prompts with Bubbletea for user input

### API Client (`internal/api/`)
- HTTP client with OAuth2 authentication
- Rate limiting: 50 requests per 10 seconds
- Automatic retry with exponential backoff
- Base URL: `https://3.basecampapi.com/{account_id}`

### Authentication (`internal/auth/`)
- OAuth2 flow with local HTTP server (port 8888)
- Tokens stored in `~/.config/bc4/auth.json`
- Encrypted token storage with 0600 permissions

### Configuration (`internal/config/`)
- Uses Viper for configuration management
- Config file: `~/.config/bc4/config.json`
- Environment variable overrides supported

### UI Components (`internal/tui/`)
- Bubbletea for interactive terminal UIs
- Lipgloss for styling
- Glamour for markdown rendering
- Consistent color scheme and responsive layouts

## Key Implementation Notes

1. **OAuth Setup Required**: Users need to create OAuth app at https://launchpad.37signals.com/integrations
2. **First-Run Wizard**: Guide users through OAuth setup on first use
3. **Multi-Account Support**: Handle multiple Basecamp accounts with defaults
4. **Rate Limiting**: Implement token bucket algorithm for API calls
5. **Error Handling**: Show user-friendly errors with actionable suggestions

## Current State

The project structure is set up but most commands are not yet implemented. Focus areas:
1. Complete OAuth2 authentication flow
2. Implement basic API client functionality
3. Build interactive project/account selectors
4. Create todo list and create commands

## Testing Guidelines

- Mock API responses for unit tests
- Use table-driven tests for command parsing
- Test interactive components with Bubbletea's testing utilities
- Keep integration tests separate with build tags

## Security Considerations

- Never log OAuth tokens or sensitive data
- Use HTTPS only for API communication
- Store tokens with proper file permissions (0600)
- Support environment variables for CI/CD scenarios