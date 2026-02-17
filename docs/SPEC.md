# bc4 CLI Specification

## Overview

bc4 is a command-line interface for Basecamp 4, inspired by GitHub's `gh` CLI. It provides a modern, interactive experience for managing Basecamp projects, todos, messages, and campfires directly from the terminal.

Our NUMBER ONE PRIORITY is to deliver a CLI app that matches the GitHub CLI tool as closely as we can. Borrow from https://github.com/cli/cli as much as possible.

Instead of talking to GitHub, we're talking to Basecamp 3 (aka 4) whose API is documented at https://github.com/basecamp/bc3-api in detail.

## Core Principles


1. **User-Friendly**: Interactive prompts using Charm's Bubbletea for beautiful TUIs
2. **Secure**: OAuth2 authentication with secure token storage
3. **Efficient**: Respect API rate limits with intelligent retry logic
4. **Cross-Platform**: Single binary distribution, works on macOS/Linux/Windows
5. **Idiomatic Go**: Clean, well-documented, testable code
6. **URL-Aware**: Accept Basecamp URLs directly as command parameters

## Architecture

### Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra + Viper
- **HTTP Client**: Standard library with custom middleware
- **OAuth2**: golang.org/x/oauth2
- **Terminal UI**: Charm tools (Bubbletea, Bubbles, Lipgloss, Glamour)
- **Table Rendering**: GitHub CLI-inspired table system with intelligent column widths
- **Configuration**: JSON format with Viper
- **Testing**: Standard library + testify

### Project Structure

```
bc4/
├── cmd/                    # Command implementations
│   ├── auth/              # Authentication commands
│   ├── account/           # Account management
│   ├── project/           # Project commands
│   ├── todo/              # Todo commands
│   ├── message/           # Message commands
│   ├── campfire/          # Campfire commands
│   ├── card/              # Card table commands
│   └── root.go            # Root command setup
├── internal/              # Private packages
│   ├── api/               # API client and models
│   ├── auth/              # OAuth2 implementation
│   ├── config/            # Configuration management
│   ├── parser/            # URL parsing utilities
│   ├── tableprinter/      # Core GitHub CLI-style table rendering
│   ├── tui/               # Bubbletea components
│   ├── ui/                # UI utilities and table wrapper
│   │   └── tableprinter/  # bc4-specific table enhancements
│   └── utils/             # Helper functions
├── pkg/                   # Public packages (if needed)
├── main.go               # Entry point
├── SPEC.md               # This file
├── README.md             # User documentation
└── Makefile              # Build automation
```

## Authentication

### OAuth2 Flow

1. **Local HTTP Server** (primary method):
   - Start temporary HTTP server on `localhost:8888`
   - Open browser to Basecamp OAuth URL
   - Capture callback with authorization code
   - Exchange code for access token

2. **Manual Code Entry** (fallback):
   - Display OAuth URL for manual navigation
   - Prompt user to paste authorization code
   - Exchange code for access token

### First-Run Setup Wizard

Using Bubbletea, guide users through:

1. **Welcome Screen**
   - Explain what bc4 does
   - Check for existing configuration

2. **OAuth App Creation**
   - Direct to https://launchpad.37signals.com/integrations
   - Interactive prompt for Client ID and Secret
   - Validate and save credentials

3. **Authentication**
   - Perform OAuth2 flow
   - Test API connection
   - Save tokens securely

4. **Account Selection**
   - Fetch available accounts
   - Let user select default account
   - Option to skip for later

5. **Project Selection** (optional)
   - List projects in default account
   - Select default project
   - Can be changed later

### Token Storage

- **Location**: `~/.config/bc4/auth.json`
- **Format**: JSON with encryption for sensitive data
- **Permissions**: 0600 (user read/write only)
- **Structure**:
  ```json
  {
    "default_account": "account_id",
    "accounts": {
      "account_id": {
        "access_token": "encrypted_token",
        "refresh_token": "encrypted_token",
        "expiry": "2024-01-01T00:00:00Z",
        "account_name": "My Company"
      }
    }
  }
  ```

## Configuration

### Configuration File

- **Location**: `~/.config/bc4/config.json`
- **Structure**:
  ```json
  {
    "default_account": "account_id",
    "default_project": "project_id",
    "accounts": {
      "account_id": {
        "name": "My Company",
        "default_project": "project_id"
      }
    },
    "preferences": {
      "editor": "vim",
      "pager": "less",
      "color": "auto"
    }
  }
  ```

### Environment Variables

- `BC4_ACCOUNT_ID`: Override default account
- `BC4_PROJECT_ID`: Override default project
- `BC4_ACCESS_TOKEN`: Direct token (for CI/CD)
- `BC4_NO_COLOR`: Disable color output
- `BC4_CONFIG_DIR`: Alternative config directory

## Command Structure

### Global Flags

- `--account, -a`: Specify account ID
- `--project, -p`: Specify project ID
- `--json`: Output JSON format
- `--no-color`: Disable color output
- `--help, -h`: Show help

### Commands

#### auth
```bash
bc4 auth login              # Interactive OAuth flow
bc4 auth logout             # Remove stored credentials
bc4 auth status             # Show authentication status
bc4 auth refresh            # Manually refresh token
```

#### account
```bash
bc4 account list            # List all accounts
bc4 account select          # Interactive account selection
bc4 account set [ID]        # Set default account
bc4 account current         # Show current account
```

#### project
```bash
bc4 project list            # List projects in account
bc4 project select          # Interactive project selection (implemented with table UI)
bc4 project set [ID]        # Set default project (non-interactive)
bc4 project view [ID|URL]   # View project details (accepts ID or Basecamp URL)
bc4 project search [query]  # Search projects by name
```

Note: The `project select` command provides an interactive table-based UI for browsing and selecting a default project, while `project set` allows direct setting by ID.

#### todo
```bash
bc4 todo lists              # List all todo lists in project (GitHub CLI-style table)
bc4 todo list [ID|name]     # View todos in a specific list (GitHub CLI-style table)
bc4 todo list [ID|name] --all      # Include completed todos
bc4 todo list [ID|name] --grouped  # Show groups separately with headers instead of columns
bc4 todo view [ID|URL]      # View details of a specific todo (accepts ID or Basecamp URL)
bc4 todo select             # Interactive todo list selection (not yet implemented)
bc4 todo set [ID]           # Set default todo list
bc4 todo add "Task"         # Create a new todo (supports --list, --description, --due, --file flags)
bc4 todo check [ID|URL]     # Mark todo as complete (accepts ID or Basecamp URL)
bc4 todo uncheck [ID|URL]   # Mark todo as incomplete (accepts ID or Basecamp URL)
bc4 todo create-list "Name" # Create a new todo list (supports --description flag)
bc4 todo create-group "Name" # Create a new group within a todo list (supports --list flag)
bc4 todo reposition-group [ID] [position] # Reposition a group within a todo list
```

**Table Output Features:**
- Dynamic headers: TTY mode shows human-readable columns, non-TTY adds STATE column for scripts
- Status symbols: ✓ for completed, ○ for incomplete (TTY mode)
- Intelligent column widths with GitHub CLI's responsive algorithm
- Color coding: Green for incomplete, Red for completed, Cyan for names, Muted for timestamps
- Default indicators: * suffix for default todo lists/projects/accounts

**Todo List Display Modes:**
- **Default**: Single table with GROUP column for grouped todo lists
- **--grouped**: Separate sections for each group with group headers
- **--all**: Include completed todos (by default only shows open todos)
- **Combination**: Use `--grouped --all` to show all todos organized by group sections

**Todo Commands Details:**

- **`todo lists`**: Shows all todo lists in the project
- **`todo list [ID|name]`**: Shows todos within a specific list
- **`todo view [ID|URL]`**: Shows detailed information about a single todo
  - Accepts numeric ID or Basecamp URL (e.g., `https://3.basecamp.com/1234567/buckets/89012345/todos/12345`)
- **`todo add "Task"`**: Creates a new todo in the default list
  - `--list, -l`: Specify todo list by ID, name, or URL
  - `--description, -d`: Add a description to the todo (supports Markdown)
  - `--due`: Set due date (YYYY-MM-DD format)
  - `--assign`: Assign to team members (not yet implemented)
  - `--file, -f`: Read todo content from a Markdown file
- **`todo check [ID|URL]`**: Marks a todo as complete
  - Accepts #ID, ID, or Basecamp URL
- **`todo uncheck [ID|URL]`**: Marks a todo as incomplete
  - Accepts #ID, ID, or Basecamp URL
- **`todo create-list "Name"`**: Creates a new todo list
  - `--description, -d`: Add a description to the list
- **`todo create-group "Name"`**: Creates a new group within a todo list
  - `--list, -l`: Specify todo list by ID, name, or URL (defaults to selected list)
- **`todo reposition-group [ID] [position]`**: Repositions a group within a todo list
  - Accepts group ID or Basecamp URL
  - Position is 1-based (1 = first position)

#### message
```bash
bc4 message list            # List recent messages
bc4 message post            # Interactive message creation
bc4 message post "Title" "Body"  # Quick message post
bc4 message view [ID]       # View message thread
```

#### campfire
```bash
bc4 campfire list                 # List all campfires in project (GitHub CLI-style table)
bc4 campfire select               # Interactive campfire selection to set default
bc4 campfire set [ID|name]        # Set default campfire (non-interactive)  
bc4 campfire view [ID|name|URL]   # View recent messages in a campfire (accepts ID, name, or URL)
bc4 campfire post                 # Interactive message composition
bc4 campfire post "Message"       # Quick message to default campfire
```

**Campfire Command Details:**

- **`campfire list`**: Shows all campfires in the current project
  - Table columns: ID, NAME, STATUS, LAST ACTIVITY
  - Default campfire marked with * suffix
  - Color coding: green for active, gray for inactive
  - `--all`: Show campfires from all projects (adds PROJECT column)

- **`campfire select`**: Interactive selection to set default campfire (not yet implemented)

- **`campfire set [ID|name]`**: Set default campfire by ID or name
  - Saves to per-project configuration
  - Accepts campfire ID or partial name match

- **`campfire view [ID|name|URL]`**: View recent messages from a campfire
  - Accepts numeric ID, name, or Basecamp URL (e.g., `https://3.basecamp.com/1234567/buckets/89012345/chats/12345`)
  - Shows last 50 messages by default
  - `--limit, -n`: Specify number of messages to show
  - `--since`: Show messages since timestamp
  - `--follow, -f`: Follow mode for live updates (future enhancement)
  - Uses campfire specified or default if none provided

- **`campfire post`**: Post messages to campfire
  - Interactive mode: Multi-line message composition with Bubbletea
  - Quick mode: Single argument for simple messages
  - `--campfire, -c`: Specify campfire by ID, name, or URL (overrides default)
  - Posts to default campfire if none specified

**Default Campfire Behavior:**
- Each project can have a default campfire set via `campfire set` 
- Commands use the default campfire when no explicit campfire is specified
- The `--campfire` flag overrides the default for any command
- Follows the same pattern as todo list defaults

#### card
```bash
bc4 card list                  # List card tables in project
bc4 card table [ID|name]       # View cards in a specific table
bc4 card select                # Interactive card table selection (not yet implemented)
bc4 card set [ID|name]         # Set default card table
bc4 card view [ID|URL]         # View card details including steps (accepts ID or URL)
bc4 card create                # Interactive card creation
bc4 card add "Title"           # Quick card creation
bc4 card edit [ID|URL]         # Edit card title/content (accepts ID or URL)
bc4 card move [ID|URL]         # Move card between columns (accepts ID or URL)
bc4 card assign [ID|URL]       # Assign people to card (accepts ID or URL)
bc4 card unassign [ID|URL]     # Remove assignees from card (accepts ID or URL)
bc4 card archive [ID|URL]      # Archive a card (accepts ID or URL)

# Column management
bc4 card column list        # List columns in current table
bc4 card column create      # Create new column
bc4 card column edit [ID]   # Edit column name/description
bc4 card column move [ID]   # Reorder columns
bc4 card column color [ID]  # Set column color

# Step management (subtasks within cards)
bc4 card step add [ID|URL] "Title"         # Add step to card (accepts card ID or URL)
bc4 card step list [ID|URL]                # List all steps in card (accepts card ID or URL)
bc4 card step check [ID|URL] [STEP|URL]    # Mark step as complete (accepts IDs or step URL)
bc4 card step uncheck [ID|URL] [STEP|URL]  # Mark step as incomplete (accepts IDs or step URL)
bc4 card step edit [ID|URL] [STEP|URL]     # Edit step details (accepts IDs or step URL)
bc4 card step move [ID|URL] [STEP|URL]     # Reorder steps (accepts IDs or step URL)
bc4 card step assign [ID|URL] [STEP|URL]   # Assign step to someone (accepts IDs or step URL)
bc4 card step delete [ID|URL] [STEP|URL]   # Delete a step (accepts IDs or step URL)
```

**Card Command Details:**

- **`card list`**: Shows all card tables in the current project
  - Table columns: ID, NAME, DESCRIPTION, CARDS, UPDATED
  - Default card table marked with * suffix
  
- **`card table [ID|name]`**: Shows cards in a specific table
  - List view with columns showing card status
  - Shows step progress (e.g., "Steps: 2/5")
  - Groups cards by column
  - `--column`: Filter to show only specific column
  - `--format`: Output format (table, json, csv)
  
- **`card view [ID|URL]`**: Shows detailed card information
  - Accepts numeric ID or Basecamp URL (e.g., `https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345`)
  - Card title, description, assignees, due date
  - Complete list of steps with completion status
  - Step assignees and due dates
  - `--steps-only`: Show only the steps list
  
- **`card create`**: Interactive card creation with TUI
  - Select table and column
  - Add title, description, assignees
  - Add initial steps
  
- **`card add "Title"`**: Quick card creation
  - Creates in default table's first column
  - `--table`: Specify card table by ID or URL
  - `--column`: Specify target column
  - `--assign`: Add assignees
  - `--step`: Add steps (can be used multiple times)
  
- **`card move [ID|URL]`**: Move card to different column
  - Accepts numeric ID or Basecamp URL
  - Interactive column selection if not specified
  - `--column`: Target column name or ID
  
- **Column Management:**
  - Columns represent workflow stages in card tables
  - Support for custom colors (white, red, orange, yellow, green, blue, aqua, purple, gray, pink, brown)
  - Columns can be reordered to match workflow
  
- **Step Management:**
  - Steps are mini-todos within cards
  - Can be referenced by:
    - Card and step IDs: `bc4 card step check 123 456`
    - Step URL: `bc4 card step check https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/123/steps/456`
  - Steps can be individually assigned and have due dates
  - Step completion is independent of card status
  - Steps maintain order and can be reordered

**Default Card Table Behavior:**
- Each project can have a default card table set via `card set`
- Commands use the default table when none is specified
- The `--table` flag overrides the default for any command
- Follows the same pattern as todo list and campfire defaults

## Rich Text and Markdown Support

### Overview

bc4 supports automatic conversion between Markdown and Basecamp's rich text HTML format. Users can write in familiar Markdown syntax, and bc4 handles the conversion transparently.

### Supported Resources

According to the Basecamp API, the following resources support rich text content:

**Currently Implemented:**
- ✅ **Todo** - `content` (title) and `description` fields support Markdown input
- ✅ **Comment** - `content` field supports Markdown input and bc-attachment tags
- ✅ **Campfire line** - uses `content_type: "text/html"` to send rich text

**Future Implementation:**
- 🔄 **Card** - `title` and `content` fields will support Markdown
- 🔄 **Message** - `content` field will support Markdown
- 🔄 **Document** - `content` field will support Markdown
- 🔄 **Schedule entry** - `description` field will support Markdown
- 🔄 **Upload** - `description` field will support Markdown
- 🔄 **To-do list** - `description` field will support Markdown

### Markdown to Rich Text Conversion

The `internal/markdown` package provides bidirectional conversion:

1. **MarkdownToRichText**: Converts GitHub Flavored Markdown to Basecamp's HTML format
   - Used when creating or updating content
   - Supports all standard GFM elements
   - Automatically wraps content in appropriate HTML tags

2. **RichTextToMarkdown**: Converts Basecamp HTML back to Markdown (not yet implemented)
   - Will be used for displaying content in the terminal
   - Preserves formatting and structure

### Supported Markdown Elements

- **Text formatting**: Bold (`**`), italic (`*`), strikethrough (`~~`)
- **Headings**: All levels (converted to `<h1>` per Basecamp limitations)
- **Links**: Inline links and auto-links
- **Code**: Inline code and fenced code blocks (converted to `<pre>`)
- **Lists**: Ordered and unordered, with nesting support
- **Blockquotes**: Standard markdown quotes
- **Line breaks**: Hard breaks with two spaces

### Input Methods

1. **Command arguments**: Direct markdown in quotes
2. **File input**: `--file` flag to read from .md files
3. **Stdin**: Pipe markdown content into commands
4. **Interactive**: Future TUI editor with markdown preview

### Comments and Attachments

- `comment create`: accepts `--attach <path>` to upload a single file and append a `<bc-attachment sgid="...">` tag after Markdown-to-rich-text conversion.
- `comment attach <recording-id|url>`: uploads a file and appends it to an existing comment (latest by default, or `--comment-id` to target a specific one).
- Attachment upload flow: `POST /attachments.json?name=<filename>` with raw body and content type, then embed the returned `attachable_sgid` in the comment HTML.
- Attachment download flow: `GET /buckets/{bucket}/uploads/{id}.json` to get metadata including `download_url`, then `GET {download_url}` with OAuth Bearer token to download the file. Supported via `bc4 card download-attachments <card-id>` command.

## API Integration

### HTTP Client Configuration

```go
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    UserAgent  string
    RateLimit  *RateLimiter
}
```

### Required Headers

- `Authorization: Bearer {token}`
- `User-Agent: bc4-cli/1.0.0 (github.com/user/bc4)`
- `Content-Type: application/json; charset=utf-8`

### Rate Limiting

- **Limit**: 50 requests per 10 seconds
- **Implementation**:
  - Token bucket algorithm
  - Automatic retry with exponential backoff
  - Show progress with Bubbletea spinner
  - Honor `Retry-After` header

### Error Handling

- **Network Errors**: Retry with backoff
- **401 Unauthorized**: Trigger re-authentication
- **404 Not Found**: Clear error message
- **429 Too Many Requests**: Wait and retry
- **5xx Server Errors**: Retry with backoff

## Table Rendering System

### GitHub CLI-Inspired Architecture

bc4 implements a comprehensive table rendering system modeled after GitHub CLI's proven approach, providing professional, responsive table output across all commands.

#### Two-Layer Design

1. **Core Layer** (`internal/tableprinter`):
   - `TablePrinter` interface with fluent API (AddHeader/AddField/EndRow)
   - `ttyTablePrinter`: Terminal output with formatting, colors, intelligent column widths
   - `csvTablePrinter`: Comma-separated output for scripts/pipes
   - Field-level formatting options (WithColor, WithTruncate, WithPadding)

2. **Integration Layer** (`internal/ui/tableprinter`):
   - bc4-specific wrapper with Basecamp entity helpers
   - Time formatting utilities (relative for TTY, RFC3339 for non-TTY)
   - Color scheme with state-based functions
   - Convenience methods: AddIDField, AddProjectField, AddTodoField, etc.

#### Key Features

**TTY vs Non-TTY Behavior:**
- **TTY Mode**: Human-readable with colors, symbols (✓/○), relative times, #ID prefixes
- **Non-TTY Mode**: Machine-readable CSV, RFC3339 times, additional STATE columns

**Intelligent Column Widths:**
- GitHub CLI's exact proportional distribution algorithm
- 8-character minimums, 3-space separators, responsive to terminal width
- Natural width measurement with ANSI code stripping

**Dynamic Headers:**
- TTY mode: User-friendly headers like ["ID", "NAME", "DESCRIPTION", "UPDATED"]
- Non-TTY mode: Adds "STATE" column for programmatic access

**Color Scheme:**
- Green: active/open items • Red: completed/closed items
- Cyan: names/identifiers • Gray: inactive/draft items  
- Muted: timestamps/secondary info
- Respects NO_COLOR, CLICOLOR, FORCE_COLOR environment variables

#### Usage Examples

```go
// Create table with automatic TTY detection
table := tableprinter.New(os.Stdout)

// Dynamic headers based on mode
if table.IsTTY() {
    table.AddHeader("ID", "NAME", "STATUS", "UPDATED")
} else {
    table.AddHeader("ID", "NAME", "STATUS", "STATE", "UPDATED")
}

// Add data with field-specific formatting
table.AddIDField("123", "active")           // #123 in TTY, 123 in non-TTY
table.AddProjectField("My Project", "active") // Green color for active
table.AddStatusField(false)                  // ○ symbol for incomplete
table.AddTimeField(now, updatedAt)          // "2 hours ago" vs RFC3339
table.EndRow()

table.Render()
```

## Security Considerations

1. **Token Storage**
   - Never log tokens
   - Encrypt at rest
   - Clear on logout
   - Validate token expiry

2. **API Keys**
   - Store in secure config
   - Never commit to git
   - Support environment variables

3. **HTTPS Only**
   - Verify certificates
   - No HTTP fallback

## Distribution

### Build Process

```makefile
# Makefile targets
build:          # Build for current platform
build-all:      # Build for all platforms
test:           # Run tests
lint:           # Run linters
install:        # Install locally
release:        # Create GitHub release
```

### Homebrew Distribution

1. **Personal Tap**: `homebrew-bc4`
   ```ruby
   class Bc4 < Formula
     desc "Basecamp 4 command-line interface"
     homepage "https://github.com/user/bc4"
     url "https://github.com/user/bc4/archive/v1.0.0.tar.gz"
     sha256 "..."
     
     depends_on "go" => :build
     
     def install
       system "go", "build", "-o", bin/"bc4"
     end
   end
   ```

2. **Installation**: `brew tap user/bc4 && brew install bc4`

### GitHub Releases

- Pre-built binaries for:
  - macOS (amd64, arm64)
  - Linux (amd64, arm64)
  - Windows (amd64)
- Automated via GitHub Actions
- Include checksums

## Testing Strategy

1. **Unit Tests**
   - API client mocking
   - Configuration handling
   - OAuth flow

2. **Integration Tests**
   - Real API calls (test account)
   - Full command flows

3. **UI Tests**
   - Bubbletea component testing
   - Snapshot testing for output

## Future Enhancements

- Check-in support
- Schedule management
- People directory
- File uploads/downloads
- Webhooks integration
- Plugin system
- Shell completions

## Development Guidelines

1. **Code Style**
   - Follow standard Go conventions
   - Use `gofmt` and `golangci-lint`
   - Meaningful variable names
   - Comprehensive comments

2. **Error Messages**
   - User-friendly language
   - Actionable suggestions
   - Include relevant context

3. **Documentation**
   - Inline code comments
   - README for users
   - CONTRIBUTING guide
   - Example scripts

## Version 1.0 Deliverables

- [x] **GitHub CLI-style table rendering system** - Complete two-layer architecture with intelligent column widths, TTY/non-TTY behavior, and responsive design
- [x] **OAuth2 authentication with first-run wizard** - Interactive setup flow with account selection
- [x] **Multi-account support with defaults** - Account management with default indicators
- [x] **Project listing and selection** - Interactive and direct selection with table output
- [x] **Todo list management** - List, view, and set default todo lists
- [x] **Todo viewing** - Display todos in lists with status indicators and grouping support
- [x] **Todo creation and completion commands** - add, check, uncheck, create-list implemented
- [x] **URL parameter support** - Accept Basecamp URLs directly as command arguments
- [ ] Message posting
- [ ] **Campfire messaging** (in progress - spec defined, implementation started)
- [x] **Card table management specification** - Complete command structure with step support
- [x] **Professional table output** - GitHub CLI-quality formatting across all commands
- [x] **Rate limiting and pagination** - Global pagination utility with token bucket rate limiter
- [ ] Homebrew distribution
- [ ] Comprehensive documentation

### Current Implementation Status

**Completed Core Features:**
- Full table rendering system matching GitHub CLI standards
- Account management (list, select, set, current)
- Project management (list, select, set, view, search)
- Todo list management (list, view, select, set)
- Todo operations (add, check, uncheck, create-list)
- OAuth2 authentication flow with secure token storage
- Multi-account configuration with defaults

**Table System Features:**
- Dynamic headers based on output mode (TTY vs scripts)
- Intelligent column width distribution
- Color-coded status indicators (✓/○ symbols)
- State-based coloring (green/red/cyan/gray/muted)
- Default item indicators (* suffix)
- Responsive design adapting to terminal width
