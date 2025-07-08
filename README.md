# bc4 - Basecamp Command Line Interface

A powerful command-line interface for [Basecamp](https://basecamp.com/), inspired by GitHub's `gh` CLI. Manage your projects, todos, messages, and campfire chats directly from your terminal.

## Features

- 🔐 **OAuth2 Authentication** - Secure authentication with token management
- 👥 **Multi-Account Support** - Manage multiple Basecamp accounts with ease
- 📁 **Project Management** - List, search, and select projects
- ✅ **Todo Management** - Create and list todos across projects
- 💬 **Message Posting** - Post messages to project message boards
- 🔥 **Campfire Integration** - Send updates to project campfire chats
- 🎯 **Card Management** - Manage cards with kanban board view
- 🎨 **Beautiful TUI** - Interactive interface powered by Charm tools
- 🔍 **Smart Search** - Find projects by pattern matching

## Installation

### Prerequisites

- Go 1.21 or later
- Git

### Install with Homebrew (macOS)

```bash
brew tap needmore/bc4
brew install bc4
```

### Install from source

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

# Interactively select a project
bc4 project select
```

### Todo Management

```bash
# Create new todos interactively
bc4 todo create

# List todos in a project
bc4 todo list "project name"
```

### Messaging

```bash
# Post a message to the message board
bc4 message post

# Post to campfire chat
bc4 campfire post "Quick update: deployment complete! 🚀"

# Post a formatted update to campfire
bc4 campfire update
```

### Card Management

```bash
# List card tables in project
bc4 card list

# View cards in a specific table
bc4 card table [ID]

# Create a new card interactively
bc4 card create

# Move card between columns
bc4 card move [ID]
```

## Examples

### Quick Project Access

Find and set a project as default by pattern:
```bash
bc4 project executive
# Finds first project with "executive" in the name
```

### Post a Campfire Message

```bash
# Interactive mode
bc4 campfire post

# Direct message
bc4 campfire post "marketing" "New campaign is live!"
```

### Create Todos

```bash
bc4 todo create
# Follow the prompts to select project, list, and add todos
```

## Configuration

Configuration is stored in:
- `~/.config/bc4/auth.json` - OAuth tokens (auto-generated, secure)
- `~/.config/bc4/config.json` - Default account and project settings

## Tips

1. **Set defaults**: Use `bc4 account select` and `bc4 project select` to set defaults and avoid constant selection
2. **Project patterns**: Use partial project names with `bc4 project <pattern>` for quick access
3. **Multiple accounts**: The tool handles multiple Basecamp accounts seamlessly

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

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Inspired by GitHub's `gh` CLI design
- Built for the Basecamp community