# bc4 - Basecamp Command Line Interface

A powerful command-line interface for [Basecamp](https://basecamp.com/), inspired by GitHub's `gh` CLI. Manage your projects, todos, messages, and campfire chats directly from your terminal. Made by your friends at [Needmore Designs](https://needmoredesigns.com).

## Features

- 🔐 **OAuth2 Authentication** - Secure authentication with token management
- 👥 **Multi-Account Support** - Manage multiple Basecamp accounts with ease
- 📁 **Project Management** - List, search, and select projects
- ✅ **Todo Management** - Create, list, check/uncheck todos across projects (supports Markdown → rich text)
- 💬 **Message Posting** - Post messages to project message boards
- 🔥 **Campfire Integration** - Send updates to project campfire chats
- 🎯 **Card Management** - Manage cards with kanban board view
- 🎨 **Beautiful TUI** - Interactive interface powered by Charm tools
- 🔍 **Smart Search** - Find projects by pattern matching
- 🔗 **URL Parameter Support** - Use Basecamp URLs directly as command arguments
- 📝 **Markdown Support** - Write in Markdown, automatically converted to Basecamp's rich text format

## Installation

### Install with Homebrew (macOS)

```bash
brew tap needmore/bc4 https://github.com/needmore/bc4
brew install bc4
```

### Install from source

Prerequisites for building from source:
- Go 1.21 or later
- Git

```bash
# Clone the repository
git clone https://github.com/needmore/bc4.git
cd bc4

# Build the binary
go build -o bc4

# Install to your PATH
sudo mv bc4 /usr/local/bin/

# Or install with go install
go install github.com/needmore/bc4@latest
```

## Setup

### 1. Create a Basecamp OAuth App

1. Go to https://launchpad.37signals.com/integrations
2. Click "Register one now" to create a new integration
3. Fill in the details:
   - **Name**: Your app name (e.g., "BC4 CLI")
   - **Redirect URI**: `http://localhost:8888/callback`
   - **Company**: Your company name
4. Save the integration
5. Copy your **Client ID** and **Client Secret**

### 2. First Run

When you run bc4 for the first time, it will guide you through setup:

```bash
bc4
```

The interactive setup wizard will:
- Help you enter your OAuth app credentials
- Authenticate with Basecamp
- Let you select a default account
- Configure your preferences

### 3. Manual Setup (Optional)

If you prefer to set up manually, you can provide credentials via environment variables:

```bash
export BC4_CLIENT_ID='your_client_id_here'
export BC4_CLIENT_SECRET='your_client_secret_here'
```

Then authenticate:

```bash
bc4 auth login
```

This will open your browser for authentication. After authorizing, paste the redirect URL or authorization code back into the terminal.

## Usage

### Authentication

```bash
# Log in to Basecamp
bc4 auth login

# Check authentication status
bc4 auth status
```

### Account Management

```bash
# List all accounts
bc4 account list

# Select default account
bc4 account select
```

### Project Management

```bash
# List all projects
bc4 project list

# Search for a project by name
bc4 project "marketing"

# View project details by ID or URL
bc4 project view 12345
bc4 project view https://3.basecamp.com/1234567/projects/12345

# Interactively select a project
bc4 project select
```

### Todo Management

```bash
# List all todo lists in the current project
bc4 todo lists

# View todos in a specific list
bc4 todo list [list-id|name]

# View todos with completed items included
bc4 todo list [list-id|name] --all

# View todos grouped by sections (for grouped todo lists)
bc4 todo list [list-id|name] --grouped

# View details of a specific todo
bc4 todo view 12345
bc4 todo view https://3.basecamp.com/1234567/buckets/89012345/todos/12345

# Create a new todo (supports Markdown formatting)
bc4 todo add "Review **critical** pull request"

# Create a todo with Markdown description and due date
bc4 todo add "Deploy to production" --description "After all tests pass\n\n- Check staging\n- Run **final** tests" --due 2025-01-15

# Create a todo from a Markdown file
bc4 todo add --file todo-content.md

# Create a todo from stdin
echo "# Important Task\n\nThis needs **immediate** attention" | bc4 todo add

# Create a todo in a specific list (by name, ID, or URL)
bc4 todo add "Update documentation" --list "Documentation Tasks"
bc4 todo add "Fix bug" --list 12345
bc4 todo add "New feature" --list https://3.basecamp.com/1234567/buckets/89012345/todosets/12345

# Mark a todo as complete (by ID or URL)
bc4 todo check 12345
bc4 todo check #12345  # Also accepts # prefix
bc4 todo check https://3.basecamp.com/1234567/buckets/89012345/todos/12345

# Mark a todo as incomplete (by ID or URL)
bc4 todo uncheck 12345
bc4 todo uncheck https://3.basecamp.com/1234567/buckets/89012345/todos/12345

# Create a new todo list
bc4 todo create-list "Sprint 1 Tasks"

# Create a todo list with description
bc4 todo create-list "Bug Fixes" --description "Critical bugs to fix before release"

# Select a default todo list (interactive - not yet implemented)
bc4 todo select

# Set a default todo list by ID
bc4 todo set 12345
```

### Messaging

```bash
# Post a message to the message board
bc4 message post

# Post to campfire chat
bc4 campfire post "Quick update: deployment complete! 🚀"

# Post to a specific campfire (by ID, name, or URL)
bc4 campfire post "Status update" --campfire "Engineering"
bc4 campfire post "Done!" --campfire 12345
bc4 campfire post "Shipped!" --campfire https://3.basecamp.com/1234567/buckets/89012345/chats/12345

# View campfire messages (by ID, name, or URL)
bc4 campfire view 12345
bc4 campfire view "Engineering"
bc4 campfire view https://3.basecamp.com/1234567/buckets/89012345/chats/12345

# Post a formatted update to campfire
bc4 campfire update
```

### Card Management

```bash
# List card tables in project
bc4 card list

# View cards in a specific table
bc4 card table [ID]

# View a specific card (by ID or URL)
bc4 card view 12345
bc4 card view https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345

# Create a new card in a specific table (by ID or URL)
bc4 card add "New feature" --table 12345
bc4 card add "Bug fix" --table https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345

# Edit a card (by ID or URL)
bc4 card edit 12345
bc4 card edit https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345

# Move card between columns (by ID or URL)
bc4 card move 12345 --column "In Progress"
bc4 card move https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 --column "Done"

# Assign users to a card (by ID or URL)
bc4 card assign 12345
bc4 card assign https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345

# Work with card steps
bc4 card step check 12345 456  # Card ID and Step ID
bc4 card step check https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/456
```

## Examples

### Common Workflows

#### Daily Standup Updates

```bash
# Post a formatted standup update to campfire
bc4 campfire update
# This opens an interactive editor for a formatted daily update

# Quick status update to team campfire
bc4 campfire post "Team" "PR #123 is ready for review 👀"
```

#### Managing Development Tasks

```bash
# Create a todo from a Markdown file with rich formatting
cat > task.md << EOF
# Refactor Authentication Module

## Objectives
- Improve error handling
- Add retry logic for network failures
- Update to use new OAuth2 library

## Acceptance Criteria
- [ ] All tests pass
- [ ] No breaking changes to public API
- [ ] Documentation updated
EOF

bc4 todo add --file task.md --list "Sprint 2025-01" --due 2025-01-25

# Check off todos as you complete them
bc4 todo check #18234  # Using the # prefix
bc4 todo check https://3.basecamp.com/1234567/buckets/89012345/todos/18234  # Using URL
```

#### Project Navigation

```bash
# Quickly switch between projects using patterns
bc4 project marketing    # Switches to first project matching "marketing"
bc4 project "Q1 2025"   # Switches to project with "Q1 2025" in the name

# Set a default project to avoid constant switching
bc4 project select       # Interactive project selector
```

#### Card Board Management

```bash
# View kanban board status
bc4 card list            # Shows all card tables in project
bc4 card table 12345     # Shows cards in specific table

# Move cards through workflow
bc4 card move 45678 --column "In Progress"
bc4 card move 45678 --column "Review"
bc4 card move 45678 --column "Done"

# Assign team members to cards
bc4 card assign 45678    # Interactive assignee selector
```

#### Working with URLs

```bash
# bc4 accepts Basecamp URLs directly - just copy from your browser!
bc4 todo view https://3.basecamp.com/1234567/buckets/89012345/todos/12345
bc4 card edit https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345
bc4 campfire view https://3.basecamp.com/1234567/buckets/89012345/chats/12345
```

## Configuration

Configuration is stored in:
- `~/.config/bc4/auth.json` - OAuth tokens (auto-generated, secure)
- `~/.config/bc4/config.json` - Default account and project settings

## Tips

1. **Set defaults**: Use `bc4 account select` and `bc4 project select` to set defaults and avoid constant selection
2. **Project patterns**: Use partial project names with `bc4 project <pattern>` for quick access
3. **Multiple accounts**: The tool handles multiple Basecamp accounts seamlessly
4. **URL shortcuts**: Copy Basecamp URLs from your browser and use them directly in commands - no need to extract IDs manually

## Troubleshooting

### Authentication Issues

- Ensure your OAuth app's redirect URI is exactly `http://localhost:8888/callback`
- Check that your credentials are set correctly
- Try `bc4 auth login` to re-authenticate

### Network Issues

- bc4 respects HTTP proxy settings via standard environment variables
- Ensure you have a stable internet connection
- Check firewall settings if authentication fails

## Contributing

We welcome contributions from the community! Here's how you can help:

### Filing Issues

1. **Check existing issues**: Before filing a new issue, search the [issue tracker](https://github.com/needmore/bc4/issues) to see if it has already been reported.
2. **Use clear titles**: Summarize the issue in a clear, descriptive title.
3. **Provide details**: Include:
   - Steps to reproduce the issue
   - Expected behavior
   - Actual behavior
   - Your environment (OS, Go version, bc4 version)
   - Any error messages or logs
4. **Use issue templates**: If available, use the provided issue templates for bug reports or feature requests.

### Submitting Pull Requests

1. **Fork the repository**: Create your own fork of the bc4 repository.
2. **Create a feature branch**: Use a descriptive branch name (e.g., `fix-auth-timeout`, `add-document-support`).
3. **Follow code style**: Ensure your code follows the existing patterns and passes linting:
   ```bash
   golangci-lint run
   ```
4. **Write tests**: Add tests for any new functionality or bug fixes.
5. **Update documentation**: Update the README or other docs if your changes affect user-facing functionality.
6. **Commit messages**: Use clear, descriptive commit messages following conventional commit format:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test additions or fixes
   - `refactor:` for code refactoring
7. **Open a pull request**:
   - Provide a clear description of the changes
   - Reference any related issues (e.g., "Fixes #123")
   - Ensure all CI checks pass
8. **Be responsive**: Address any feedback or requested changes promptly.

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/bc4.git
cd bc4

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the binary
go build -o bc4
```

## License

MIT License - see LICENSE file for details.

## Markdown Support

bc4 supports Markdown input for creating content that gets automatically converted to Basecamp's rich text HTML format. This works for:

### Supported Resources
- ✅ **Todos** - Both title and description support Markdown
- 🔄 **Messages** - Coming soon
- 🔄 **Documents** - Coming soon
- 🔄 **Comments** - Coming soon
- ❌ **Campfire** - Plain text only (API limitation)

### Supported Markdown Elements
- **Bold** (`**text**`), *italic* (`*text*`), ~~strikethrough~~ (`~~text~~`)
- Headings (all levels converted to `<h1>` per Basecamp spec)
- [Links](url) and auto-links
- `Inline code` and code blocks
- Ordered and unordered lists with nesting
- > Blockquotes
- Line breaks and paragraphs

### Examples
```bash
# Markdown in todo titles and descriptions
bc4 todo add "Fix **critical** bug in `Parser.parse()` method"
bc4 todo add "Refactor code" --description "## Goals\n\n- Improve **performance**\n- Add tests"

# From a Markdown file
bc4 todo add --file detailed-task.md
```

## Acknowledgments

- Inspired by GitHub's `gh` CLI design
- Built for the Basecamp community