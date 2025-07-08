# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BC4 is a command-line interface for Basecamp 4, inspired by GitHub's `gh` CLI. The project has two parallel implementations:
- **Python**: Complete implementation in `bc4.py` (primary/original version)
- **Go**: Implementation in progress using Cobra framework

## Common Development Commands

### Python Implementation
```bash
# Install dependencies
pip install -r requirements.txt

# Run the CLI
python bc4.py <command>
# or if executable
./bc4.py <command>
```

### Go Implementation
```bash
# Install dependencies
go mod download

# Build the project
go build -o bc4

# Run directly
go run main.go <command>
```

## Architecture Overview

### Python Implementation (`bc4.py`)
- Single-file CLI with complete Basecamp API integration
- OAuth2 authentication with token storage in `~/.basecamp_token.json`
- Configuration in `~/.basecamp_config.json`
- Interactive project/todo selection using arrow keys
- Pattern-based project search functionality

### Go Implementation
Follows standard Go project layout:
- `cmd/`: Command definitions using Cobra
  - `root.go`: Main command setup
  - `projects.go`: Project management commands
  - `todos.go`: Todo management commands
- `internal/`: Internal packages
  - `api/client.go`: HTTP client for Basecamp API
  - `config/config.go`: Configuration management
  - `models/`: Data structures for projects, todos
- `main.go`: Entry point

## Key API Endpoints

The Basecamp 4 API is accessed at `https://3.basecampapi.com/<account_id>/`:
- Projects: `projects.json`
- Todos: `buckets/<project_id>/todolists/<todolist_id>/todos.json`
- Messages: `buckets/<project_id>/message_boards/<board_id>/messages.json`
- Campfire: `buckets/<project_id>/chats/<chat_id>/lines.json`

## Configuration

Both implementations use:
- OAuth token storage: `~/.basecamp_token.json`
- Account configuration: `~/.basecamp_config.json`
- The token file stores OAuth credentials per account ID
- The config file stores the default account ID

## Development Notes

1. When modifying API calls, ensure proper OAuth2 authentication headers are included
2. Both implementations support multiple Basecamp accounts
3. Interactive features use pattern matching for project selection
4. The Python version is feature-complete while the Go version is being developed
5. No test suite currently exists - consider manual testing of API interactions
6. When adding new commands, maintain consistency with the existing command structure