# bc4 - Basecamp Command Line Interface

A powerful command-line interface for [Basecamp](https://basecamp.com/), strongly inspired by [GitHub CLI](https://github.com/cli/cli). Manage your projects, todos, cards, messages, chats, and more, directly from your terminal. Made by your friends at [Needmore Designs](https://needmoredesigns.com).

## Features

- 🔐 **OAuth2 Authentication** - Secure authentication with token management
- 👥 **Multi-Account Support** - Manage multiple Basecamp accounts with ease
- 📁 **Project Management** - List, search, and select projects
- ✅ **Todo Management** - Create, list, edit, check/uncheck todos across projects (supports Markdown → rich text, grouping, and group management)
- 💬 **Message Posting** - Post messages to project message boards
- 📄 **Document Management** - Create, edit, and view documents in your projects
- 💭 **Comment Management** - View, create, edit, and delete comments on todos, messages, documents, and cards
- 🔥 **Campfire Integration** - Send updates to project campfire chats
- 🎯 **Card Management** - Manage cards, columns, and steps with kanban board view
- 👤 **Profile Information** - View your Basecamp profile and account details
- 🎨 **Beautiful TUI** - Interactive interface
- 🔍 **Smart Search** - Find projects by pattern matching
- 🔗 **URL Parameter Support** - Use Basecamp URLs directly as command arguments
- 📝 **Markdown Support** - Write in Markdown, automatically converted to Basecamp's rich text format
- 📊 **Activity Monitoring** - Track project activity with real-time watch mode and advanced filtering
- 🔄 **Shell Completion** - Tab-completion for bash, zsh, fish, and PowerShell
- 🖥️ **Cross-Platform** - Available for macOS, Linux, and Windows

## Installation

### Install with Homebrew (macOS and Linux)

```bash
brew tap needmore/bc4 https://github.com/needmore/bc4
brew install bc4
```

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/needmore/bc4/releases):

- **macOS** (Intel): `bc4_darwin_amd64`
- **macOS** (Apple Silicon): `bc4_darwin_arm64`
- **Linux** (64-bit): `bc4_linux_amd64`
- **Linux** (ARM64): `bc4_linux_arm64`
- **Windows** (64-bit): `bc4_windows_amd64.exe`
- **Windows** (ARM64): `bc4_windows_arm64.exe`

After downloading, make the binary executable (macOS/Linux):
```bash
chmod +x bc4_*
sudo mv bc4_* /usr/local/bin/bc4
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

# Refresh authentication token
bc4 auth refresh

# Log out of Basecamp
bc4 auth logout
```

### Account Management

```bash
# List all accounts
bc4 account list

# Show current account
bc4 account current

# Select default account interactively
bc4 account select

# Set default account by ID
bc4 account set 12345
```

### Profile

```bash
# View your Basecamp profile
bc4 profile

# Output as JSON
bc4 profile --json
```

### Project Management

```bash
# List all projects
bc4 project list

# Search for a project by name
bc4 project search "marketing"

# View project details by ID or URL
bc4 project view 12345
bc4 project view https://3.basecamp.com/1234567/projects/12345

# Interactively select a project
bc4 project select

# Or set project by ID
bc4 project set 12345
```

### Todo Management

```bash
# List all todo lists in the current project
bc4 todo lists

# View todos in a specific list
bc4 todo list [list-id|name]

# View todos with completed items included
bc4 todo list [list-id|name] --all

# View todos grouped by sections (for organized todo lists)
# Use --grouped to show each group with clear headers
bc4 todo list [list-id|name] --grouped

# View todos in a flat table with GROUP column (default for grouped lists)
bc4 todo list [list-id|name]

# View details of a specific todo
bc4 todo view 12345
bc4 todo view https://3.basecamp.com/1234567/buckets/89012345/todos/12345

# View a todo with its comments inline
bc4 todo view 12345 --with-comments

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

# Edit an existing todo
bc4 todo edit 12345 --title "Updated title"
bc4 todo edit 12345 --description "New description with **markdown**"
bc4 todo edit 12345 --due 2025-02-15
bc4 todo edit 12345 --assign user@example.com
bc4 todo edit 12345 --unassign user@example.com

# Move a todo to a different position within its list
bc4 todo move 12345 --position 1    # Move to first position
bc4 todo move 12345 --top           # Move to top of list
bc4 todo move 12345 --bottom        # Move to bottom of list

# List attachments for a todo

# Download all attachments from a todo
bc4 todo download-attachments 123456

# Download to specific directory
bc4 todo download-attachments 123456 --output-dir ~/Downloads
bc4 todo attachments 12345

# Create a new todo list
bc4 todo create-list "Sprint 1 Tasks"

# Create a todo list with description
bc4 todo create-list "Bug Fixes" --description "Critical bugs to fix before release"

# Create a new group within a todo list
bc4 todo create-group "In Progress"

# Create a group in a specific list (by name, ID, or URL)
bc4 todo create-group "Completed" --list "Sprint 1 Tasks"
bc4 todo create-group "Backlog" --list 12345

# Reposition a group within a todo list (position is 1-based)
bc4 todo reposition-group 12345 1  # Move to first position
bc4 todo reposition-group 12345 3  # Move to third position

# Edit a todo list's name or description
bc4 todo edit-list 12345 --name "Renamed List"
bc4 todo edit-list "Sprint Tasks" --description "Updated description"
bc4 todo edit-list 12345 --clear-description

# Select a default todo list interactively
bc4 todo select

# Set a default todo list by ID
bc4 todo set 12345
```

### Messaging

```bash
# List messages in the current project
bc4 message list

# Post a message interactively
bc4 message post

# Post a message with title and content
bc4 message post --title "Project Update" --content "# Status\nThings are going well!"

# Post from a markdown file
cat update.md | bc4 message post --title "Weekly Update"

# View a specific message
bc4 message view 12345

# View a message with its comments inline
bc4 message view 12345 --with-comments

# Edit an existing message
bc4 message edit 12345

# Download all attachments from a message
bc4 message download-attachments 123456

# Download to specific directory
bc4 message download-attachments 123456 --output-dir ~/Downloads

# Pin a message to the top of the message board
bc4 message pin 12345
bc4 message pin https://3.basecamp.com/.../messages/12345

# Unpin a message
bc4 message unpin 12345
bc4 message unpin https://3.basecamp.com/.../messages/12345

# List all campfires in the project
bc4 campfire list

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

# Set default campfire for the project
bc4 campfire set 12345
```

### Document Management

```bash
# List all documents in the project
bc4 document list

# View a specific document
bc4 document view 12345
bc4 document view https://3.basecamp.com/1234567/buckets/89012345/documents/12345

# View a document with its comments inline
bc4 document view 12345 --with-comments

# Create a new document
bc4 document create "Meeting Notes"
bc4 document create "Spec Document" --content "# Overview\n\nThis is the spec..."

# Edit an existing document
bc4 document edit 12345
bc4 document edit 12345 --title "Updated Title"
```

### Card Management

```bash
# List card tables in project
bc4 card list

# View cards in a specific table
bc4 card table [ID]

# Set default card table
bc4 card set 12345

# View a specific card (by ID or URL)
bc4 card view 12345
bc4 card view https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345

# View a card with its comments inline
bc4 card view 12345 --with-comments

# Create a new card (quick add)
bc4 card add "New feature" --table 12345
bc4 card add "Bug fix" --table https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345

# Create a card interactively
bc4 card create

# Edit a card (by ID or URL)
bc4 card edit 12345
bc4 card edit https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345

# Move card between columns (by ID or URL)
bc4 card move 12345 --column "In Progress"
bc4 card move https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 --column "Done"

# Assign users to a card (by ID or URL)
bc4 card assign 12345

# Remove assignees from a card
bc4 card unassign 12345

# Archive a card
bc4 card archive 12345

# List attachments for a card
bc4 card attachments 12345

# Download all attachments from a card
bc4 card download-attachments 123456

# Download to specific directory
bc4 card download-attachments 123456 --output-dir ~/Downloads

# Download specific attachment only
bc4 card download-attachments 123456 --attachment 1
```

#### Card Columns

```bash
# List columns in a card table
bc4 card column list 12345

# Create a new column
bc4 card column create 12345 "In Review"

# Edit a column
bc4 card column edit 12345 --name "Code Review"

# Move a column to a different position
bc4 card column move 12345 --position 2

# Set column color
bc4 card column color 12345 blue
```

#### Card Steps

```bash
# List steps in a card
bc4 card step list 12345

# Add a step to a card
bc4 card step add 12345 "Review the code"

# Check/uncheck a step
bc4 card step check 12345 456    # Card ID and Step ID
bc4 card step uncheck 12345 456

# Edit a step
bc4 card step edit 456 --content "Updated step content"

# Assign a step to a user
bc4 card step assign 456

# Move a step to a different position
bc4 card step move 456 --position 1

# Delete a step
bc4 card step delete 456
```

### Comment Management

```bash
# List comments on a recording (todo, message, document, or card)
bc4 comment list 12345  # Using recording ID
bc4 comment list https://3.basecamp.com/1234567/buckets/89012345/todos/12345  # Using URL

# View a specific comment (by ID or URL)
bc4 comment view 67890
bc4 comment view https://3.basecamp.com/1234567/buckets/89012345/comments/67890

# Create a comment interactively
bc4 comment create 12345
bc4 comment create https://3.basecamp.com/1234567/buckets/89012345/todos/12345

# Create a comment with inline content (supports Markdown)
bc4 comment create 12345 --content "Great work on this! **Approved** ✅"

# Create a comment from stdin
echo "# Review Notes\n\nLooks good to me!" | bc4 comment create 12345

# Create a comment with an attachment (single file)
bc4 comment create 12345 --attach ./diagram.png

# Edit a comment (by ID or URL)
bc4 comment edit 67890
bc4 comment edit https://3.basecamp.com/1234567/buckets/89012345/comments/67890

# Edit with inline content
bc4 comment edit 67890 --content "Updated: this is now **complete**"

# Delete a comment (with confirmation prompt)
bc4 comment delete 67890
bc4 comment delete https://3.basecamp.com/1234567/buckets/89012345/comments/67890

# Delete without confirmation
bc4 comment delete 67890 --yes

# Append an attachment to the latest comment on a recording
bc4 comment attach 12345 --attach ./log.txt

# Append an attachment to a specific comment by ID
bc4 comment attach 12345 --comment-id 67890 --attach ./screenshot.png
```


### Downloading Attachments

bc4 can download images and files attached to cards, todos, and messages using OAuth authentication:

```bash
# Download all attachments from a card
bc4 card download-attachments 123456

# Download from a todo to specific directory
bc4 todo download-attachments 789012 --output-dir ~/Downloads

# Download from a message (only first attachment)
bc4 message download-attachments 345678 --attachment 1

# Overwrite existing files
bc4 card download-attachments 123456 --overwrite

# Works with Basecamp URLs too
bc4 card download-attachments https://3.basecamp.com/123/buckets/456/card_tables/cards/789
```

**Common options:**
- `--output-dir, -o` - Directory to save attachments (default: current directory)
- `--attachment N` - Download only the Nth attachment (1-based index)
- `--overwrite` - Replace existing files without prompting

**Note:** Comment attachments use blob storage that requires browser authentication and cannot be downloaded via OAuth. To download comment attachments, access them through your web browser while logged into Basecamp.

### Activity & Events

```bash
# List recent activity in the current project
bc4 activity list

# List activity from the last 24 hours
bc4 activity list --since "24h"

# List activity from the last 7 days
bc4 activity list --since "7d"

# List activity since a specific date
bc4 activity list --since "2025-01-01"
bc4 activity list --since "yesterday"
bc4 activity list --since "this week"

# Filter activity by type
bc4 activity list --type todo
bc4 activity list --type message
bc4 activity list --type "todo,message,document"

# Filter activity by person (by name, email, or ID)
bc4 activity list --person "John Doe"
bc4 activity list --person john@example.com
bc4 activity list --person 12345

# Combine filters
bc4 activity list --since "24h" --type todo --person "John Doe"

# Limit the number of results
bc4 activity list --limit 50

# Output as JSON
bc4 activity list --format json

# Watch for real-time activity (polls every 30 seconds)
bc4 activity watch

# Watch with custom polling interval (in seconds)
bc4 activity watch --interval 10

# Watch with filters
bc4 activity watch --type todo --person "John Doe"
```

## Examples

### Common Workflows

#### Daily Standup Updates

```bash
# Quick status update to team campfire
bc4 campfire post "PR #123 is ready for review 👀"

# Post to a specific campfire
bc4 campfire post "Team standup: All tests passing ✅" --campfire "Engineering"
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
2. **Override defaults**: Use `--account` and `--project` flags on any command to temporarily override your defaults
3. **Project patterns**: Use partial project names with `bc4 project <pattern>` for quick access
4. **Multiple accounts**: The tool handles multiple Basecamp accounts seamlessly
5. **URL shortcuts**: Copy Basecamp URLs from your browser and use them directly in commands - no need to extract IDs manually
6. **View with comments**: Use `--with-comments` on view commands to see comments inline
7. **Shell completion**: Enable tab completion for faster command entry (see below)

## Shell Completion

bc4 supports tab-completion for bash, zsh, fish, and PowerShell. This makes it faster to type commands and discover available options.

### Bash

```bash
# Add to ~/.bashrc
source <(bc4 completion bash)

# Or install permanently
bc4 completion bash > /etc/bash_completion.d/bc4
```

### Zsh

```bash
# Add to ~/.zshrc (before compinit)
source <(bc4 completion zsh)

# Or install permanently
bc4 completion zsh > "${fpath[1]}/_bc4"
```

### Fish

```bash
bc4 completion fish | source

# Or install permanently
bc4 completion fish > ~/.config/fish/completions/bc4.fish
```

### PowerShell

```powershell
bc4 completion powershell | Out-String | Invoke-Expression

# Or add to your PowerShell profile
bc4 completion powershell >> $PROFILE
```

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
- ✅ **Messages** - List, post, view, and edit messages on project message boards
- ✅ **Comments** - Create and edit comments with Markdown formatting
- ✅ **Documents** - Create and edit documents with Markdown formatting
- ✅ **Campfire** - Post messages with Markdown formatting

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

#### Known Limitations

**Comment Attachments:** Comment attachments use blob storage URLs that require browser session cookies and do not support OAuth Bearer token authentication. The `download-attachments` command works for card body attachments, todo attachments, and message attachments, but not for attachments added in comments. To download comment attachments, use a web browser while authenticated to Basecamp.
