# bc4 - Basecamp Command Line Interface

A powerful command-line interface for [Basecamp](https://basecamp.com/), inspired by GitHub's `gh` CLI. Manage your projects, todos, messages, and campfire chats directly from your terminal.

## Features

- 🔐 **OAuth2 Authentication** - Secure authentication with token management
- 👥 **Multi-Account Support** - Manage multiple Basecamp accounts with ease
- 📁 **Project Management** - List, search, and select projects
- ✅ **Todo Management** - Create and list todos across projects
- 💬 **Message Posting** - Post messages to project message boards
- 🔥 **Campfire Integration** - Send updates to project campfire chats
- 🔍 **Smart Search** - Find projects by pattern matching

## Installation

### Prerequisites

- Python 3.7+
- pip

### Install from source

```bash
# Clone the repository
git clone https://github.com/yourusername/bc4.git
cd bc4

# Install dependencies
pip install -r requirements.txt

# Make the script executable
chmod +x bc4.py

# Add to your PATH (choose one):
# Option 1: Symlink
ln -s $(pwd)/bc4.py /usr/local/bin/bc4

# Option 2: Add alias to your shell config
echo "alias bc4='$(pwd)/bc4.py'" >> ~/.bashrc  # or ~/.zshrc
```

## Setup

### 1. Create a Basecamp OAuth App

1. Go to https://launchpad.37signals.com/integrations
2. Click "Register one now" to create a new integration
3. Fill in the details:
   - **Name**: Your app name (e.g., "BC4 CLI")
   - **Redirect URI**: `http://localhost`
   - **Company**: Your company name
4. Save the integration
5. Copy your **Client ID** and **Client Secret**

### 2. Set Environment Variables

Add these to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export BASECAMP_CLIENT_ID='your_client_id_here'
export BASECAMP_CLIENT_SECRET='your_client_secret_here'
```

Then reload your shell configuration:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### 3. Authenticate

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
- `~/.basecamp_token.json` - OAuth tokens (auto-generated, secure)
- `~/.basecamp_config.json` - Default account and project settings

## Tips

1. **Set defaults**: Use `bc4 account select` and `bc4 project select` to set defaults and avoid constant selection
2. **Project patterns**: Use partial project names with `bc4 project <pattern>` for quick access
3. **Multiple accounts**: The tool handles multiple Basecamp accounts seamlessly

## Troubleshooting

### SSL Certificate Error

If you encounter SSL errors, ensure your Python installation has proper SSL support:
```bash
# For macOS with pyenv
pyenv uninstall 3.x.x
pyenv install 3.x.x
```

### Authentication Issues

- Ensure your OAuth app's redirect URI is exactly `http://localhost`
- Check that your environment variables are set correctly
- Try `bc4 auth login` to re-authenticate

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Inspired by GitHub's `gh` CLI design
- Built for the Basecamp community