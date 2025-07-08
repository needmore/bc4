#!/usr/bin/env python3
"""
Basecamp 4 CLI Tool - A command-line interface for Basecamp

This tool provides a simple way to interact with Basecamp from the command line,
similar to GitHub's 'gh' command.

Features:
    - OAuth2 authentication with token management
    - List and select projects interactively
    - Create and manage todo lists
    - Add todos with rich text descriptions
    - Post messages to message boards
    - Post updates to Campfire chat
    - View project information

Usage:
    bc4 auth login              # Authenticate with Basecamp
    bc4 auth status            # Show authentication status
    bc4 account list           # List all accounts
    bc4 account select         # Set default account
    bc4 project list           # List all projects
    bc4 project select         # Interactive project selection
    bc4 project <pattern>      # Find and select project by name
    bc4 todo create            # Create todos interactively
    bc4 todo list <project>    # List todos in a project
    bc4 message post           # Post a message to a project
    bc4 campfire post          # Post to project campfire

Author: Ray
License: MIT
"""

import os
import sys
import argparse
import requests
import webbrowser
import json
import time
from datetime import datetime
from urllib.parse import urlencode

# Configuration
DEFAULT_CLIENT_ID = os.environ.get('BASECAMP_CLIENT_ID', None)
DEFAULT_CLIENT_SECRET = os.environ.get('BASECAMP_CLIENT_SECRET', None)
DEFAULT_REDIRECT_URI = os.environ.get('BASECAMP_REDIRECT_URI', 'http://localhost')
TOKEN_FILE = os.path.expanduser("~/.basecamp_token.json")
CONFIG_FILE = os.path.expanduser("~/.basecamp_config.json")

class BasecampAPI:
    """Handle Basecamp API interactions"""
    
    def __init__(self, token_data=None):
        self.token_data = token_data
        self.base_url = "https://3.basecampapi.com"
        self.auth_url = "https://launchpad.37signals.com"
        
    @property
    def headers(self):
        """Get headers for API requests"""
        if not self.token_data:
            raise ValueError("No authentication token available")
        
        return {
            'Authorization': f"Bearer {self.token_data['access_token']}",
            'Content-Type': 'application/json',
            'User-Agent': 'Basecamp4 CLI Tool'
        }
    
    def get_identity(self):
        """Get user identity and accounts"""
        response = requests.get(f"{self.auth_url}/authorization.json", headers=self.headers)
        response.raise_for_status()
        return response.json()
    
    def get_projects(self, account_id, get_all=True):
        """Get projects for an account with optional pagination"""
        all_projects = []
        page = 1
        
        while True:
            url = f"{self.base_url}/{account_id}/projects.json?page={page}"
            response = requests.get(url, headers=self.headers)
            response.raise_for_status()
            
            projects = response.json()
            if not projects:
                break
                
            all_projects.extend(projects)
            
            # If we only want first page, break now
            if not get_all:
                break
            
            # Check if there's a next page
            link_header = response.headers.get('Link', '')
            if 'rel="next"' not in link_header:
                break
                
            page += 1
            # Small delay to respect rate limits (50 requests per 10 seconds)
            time.sleep(0.2)
            
        return all_projects
    
    def get_project(self, account_id, project_id):
        """Get a specific project"""
        url = f"{self.base_url}/{account_id}/projects/{project_id}.json"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()
    
    def get_todoset(self, account_id, project_id):
        """Get the todoset ID for a project"""
        project = self.get_project(account_id, project_id)
        
        # Look for todoset in the dock
        if 'dock' in project:
            for dock_item in project['dock']:
                if dock_item.get('name') == 'todoset':
                    return dock_item.get('id')
        
        raise ValueError("No todoset found for this project")
    
    def get_todolists(self, account_id, project_id, todoset_id):
        """Get all todo lists in a todoset"""
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/todosets/{todoset_id}/todolists.json"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()
    
    def create_todolist(self, account_id, project_id, todoset_id, name, description=""):
        """Create a new todo list"""
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/todosets/{todoset_id}/todolists.json"
        data = {
            'name': name,
            'description': f"<div>{description}</div>" if description else ""
        }
        response = requests.post(url, headers=self.headers, json=data)
        response.raise_for_status()
        return response.json()
    
    def get_todos(self, account_id, project_id, todolist_id):
        """Get todos from a todo list"""
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/todolists/{todolist_id}/todos.json"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()
    
    def create_todo(self, account_id, project_id, todolist_id, content, description=""):
        """Create a new todo"""
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/todolists/{todolist_id}/todos.json"
        data = {
            'content': content,
            'description': description if description else ""
        }
        response = requests.post(url, headers=self.headers, json=data)
        response.raise_for_status()
        return response.json()
    
    def post_message(self, account_id, project_id, title, content):
        """Post a message to the message board"""
        # First, get the message board ID
        project = self.get_project(account_id, project_id)
        message_board_id = None
        
        if 'dock' in project:
            for dock_item in project['dock']:
                if dock_item.get('name') == 'message_board':
                    message_board_id = dock_item.get('id')
                    break
        
        if not message_board_id:
            raise ValueError("No message board found for this project")
        
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/message_boards/{message_board_id}/messages.json"
        data = {
            'subject': title,
            'content': f"<div>{content}</div>"
        }
        response = requests.post(url, headers=self.headers, json=data)
        response.raise_for_status()
        return response.json()
    
    def get_campfire(self, account_id, project_id):
        """Get the campfire ID for a project"""
        project = self.get_project(account_id, project_id)
        
        if 'dock' in project:
            for dock_item in project['dock']:
                if dock_item.get('name') == 'chat':
                    return dock_item.get('id')
        
        raise ValueError("No campfire found for this project")
    
    def post_to_campfire(self, account_id, project_id, content):
        """Post a message to the project's campfire"""
        campfire_id = self.get_campfire(account_id, project_id)
        
        url = f"{self.base_url}/{account_id}/buckets/{project_id}/chats/{campfire_id}/lines.json"
        data = {
            'content': content
        }
        response = requests.post(url, headers=self.headers, json=data)
        response.raise_for_status()
        return response.json()

class BasecampAuth:
    """Handle Basecamp OAuth authentication"""
    
    def __init__(self, client_id=DEFAULT_CLIENT_ID, client_secret=DEFAULT_CLIENT_SECRET, redirect_uri=DEFAULT_REDIRECT_URI):
        if not client_id or not client_secret:
            raise ValueError(
                "Basecamp OAuth credentials not found.\n"
                "Please set BASECAMP_CLIENT_ID and BASECAMP_CLIENT_SECRET environment variables.\n"
                "\n"
                "To get these credentials:\n"
                "1. Go to https://launchpad.37signals.com/integrations\n"
                "2. Register a new integration\n"
                "3. Set the redirect URI to: http://localhost\n"
                "4. Copy the Client ID and Client Secret\n"
                "\n"
                "Then set the environment variables:\n"
                "  export BASECAMP_CLIENT_ID='your_client_id'\n"
                "  export BASECAMP_CLIENT_SECRET='your_client_secret'\n"
            )
        self.client_id = client_id
        self.client_secret = client_secret
        self.redirect_uri = redirect_uri
    
    def login(self):
        """Perform OAuth login"""
        # Step 1: Direct user to authorization page
        auth_params = {
            'client_id': self.client_id,
            'redirect_uri': self.redirect_uri,
            'response_type': 'code',
            'type': 'web_server'
        }
        auth_url = f"https://launchpad.37signals.com/authorization/new?{urlencode(auth_params)}"
        
        print("\n=== Basecamp OAuth2 Authorization ===")
        print(f"Opening browser to authorize: {auth_url}")
        print(f"If the browser doesn't open automatically, copy and paste the URL above.")
        webbrowser.open(auth_url)
        
        # Step 2: Get authorization code
        print("\nAfter authorizing, you'll be redirected to a URL like:")
        print(f"{self.redirect_uri}?code=AUTHORIZATION_CODE")
        code_input = input("\nPaste the authorization code or the entire redirect URL: ").strip()
        
        # Extract code if user pasted the whole URL
        auth_code = code_input
        if "?code=" in code_input:
            auth_code = code_input.split("?code=")[1].split("&")[0]
        
        if not auth_code:
            raise ValueError("No authorization code entered")
        
        # Step 3: Exchange code for access token
        token_params = {
            'type': 'web_server',
            'client_id': self.client_id,
            'client_secret': self.client_secret,
            'code': auth_code,
            'redirect_uri': self.redirect_uri,
            'grant_type': 'authorization_code'
        }
        
        print("Exchanging authorization code for access token...")
        token_response = requests.post('https://launchpad.37signals.com/authorization/token', data=token_params)
        token_response.raise_for_status()
        
        token_data = token_response.json()
        token_data['obtained_at'] = int(time.time())
        
        # Save token
        self.save_token(token_data)
        print("✓ Authentication successful!")
        
        return token_data
    
    def refresh_token(self, token_data):
        """Refresh an expired token"""
        if 'refresh_token' not in token_data:
            raise ValueError("No refresh token available")
        
        token_params = {
            'type': 'refresh',
            'client_id': self.client_id,
            'client_secret': self.client_secret,
            'refresh_token': token_data['refresh_token'],
            'grant_type': 'refresh_token'
        }
        
        token_response = requests.post('https://launchpad.37signals.com/authorization/token', data=token_params)
        token_response.raise_for_status()
        
        new_token_data = token_response.json()
        new_token_data['obtained_at'] = int(time.time())
        
        # Preserve refresh token if not included in response
        if 'refresh_token' not in new_token_data and 'refresh_token' in token_data:
            new_token_data['refresh_token'] = token_data['refresh_token']
        
        self.save_token(new_token_data)
        return new_token_data
    
    def load_token(self):
        """Load token from file"""
        if not os.path.exists(TOKEN_FILE):
            return None
        
        try:
            with open(TOKEN_FILE, 'r') as f:
                return json.load(f)
        except Exception:
            return None
    
    def save_token(self, token_data):
        """Save token to file"""
        os.makedirs(os.path.dirname(TOKEN_FILE), exist_ok=True)
        with open(TOKEN_FILE, 'w') as f:
            json.dump(token_data, f, indent=2)
        os.chmod(TOKEN_FILE, 0o600)  # Restrict access to owner only
    
    def get_valid_token(self):
        """Get a valid token, refreshing if necessary"""
        token_data = self.load_token()
        
        if not token_data:
            return None
        
        # Check if token is expired
        current_time = int(time.time())
        obtained_at = token_data.get('obtained_at', 0)
        expires_in = token_data.get('expires_in', 7200)
        
        if (current_time - obtained_at) > (expires_in - 300):  # Refresh 5 min before expiry
            print("Token expired, refreshing...")
            try:
                token_data = self.refresh_token(token_data)
            except Exception as e:
                print(f"Failed to refresh token: {e}")
                return None
        
        return token_data

class BasecampCLI:
    """Main CLI application"""
    
    def __init__(self):
        self.auth = BasecampAuth()
        self.api = None
        self.config = self.load_config()
    
    def load_config(self):
        """Load configuration file"""
        if os.path.exists(CONFIG_FILE):
            try:
                with open(CONFIG_FILE, 'r') as f:
                    return json.load(f)
            except Exception:
                pass
        return {}
    
    def save_config(self):
        """Save configuration file"""
        os.makedirs(os.path.dirname(CONFIG_FILE), exist_ok=True)
        with open(CONFIG_FILE, 'w') as f:
            json.dump(self.config, f, indent=2)
    
    def ensure_auth(self):
        """Ensure we have valid authentication"""
        token_data = self.auth.get_valid_token()
        
        if not token_data:
            print("No valid authentication found. Please log in.")
            token_data = self.auth.login()
        
        self.api = BasecampAPI(token_data)
        return token_data
    
    def cmd_auth_login(self, args):
        """Handle auth login command"""
        token_data = self.auth.login()
        self.api = BasecampAPI(token_data)
        
        # Get and display user info
        identity = self.api.get_identity()
        user = identity.get('identity', {})
        print(f"\nLogged in as: {user.get('first_name')} {user.get('last_name')} ({user.get('email_address')})")
        
        # Save default account if there's only one
        accounts = [a for a in identity.get('accounts', []) if a.get('product') in ['bc3', 'bc4']]
        if len(accounts) == 1:
            self.config['default_account_id'] = str(accounts[0]['id'])
            self.save_config()
            print(f"Set default account: {accounts[0]['name']}")
        elif len(accounts) > 1:
            print(f"\nFound {len(accounts)} accounts:")
            for account in accounts:
                default = " (default)" if str(account['id']) == self.config.get('default_account_id') else ""
                print(f"  - {account['name']}{default}")
            if not self.config.get('default_account_id'):
                print("\nRun 'bc4 account select' to set a default account.")
    
    def cmd_auth_status(self, args):
        """Handle auth status command"""
        token_data = self.auth.load_token()
        
        if not token_data:
            print("Not authenticated. Run 'bc4 auth login' to authenticate.")
            return
        
        # Check if token is valid
        current_time = int(time.time())
        obtained_at = token_data.get('obtained_at', 0)
        expires_in = token_data.get('expires_in', 7200)
        remaining = (obtained_at + expires_in) - current_time
        
        if remaining <= 0:
            print("Authentication expired. Run 'bc4 auth login' to re-authenticate.")
            return
        
        # Get user info
        try:
            self.api = BasecampAPI(token_data)
            identity = self.api.get_identity()
            user = identity.get('identity', {})
            
            print(f"Authenticated as: {user.get('first_name')} {user.get('last_name')} ({user.get('email_address')})")
            print(f"Token expires in: {remaining // 60} minutes")
            
            # Show accounts
            accounts = [a for a in identity.get('accounts', []) if a.get('product') in ['bc3', 'bc4']]
            if accounts:
                print(f"\nAccounts ({len(accounts)}):")
                for account in accounts:
                    default = " (default)" if str(account['id']) == self.config.get('default_account_id') else ""
                    print(f"  - {account['name']} (ID: {account['id']}){default}")
        except Exception as e:
            print(f"Error checking authentication: {e}")
    
    def cmd_account_list(self, args):
        """Handle account list command"""
        self.ensure_auth()
        
        try:
            identity = self.api.get_identity()
            accounts = [a for a in identity.get('accounts', []) if a.get('product') in ['bc3', 'bc4']]
            
            if not accounts:
                print("No Basecamp accounts found.")
                return
            
            print(f"\nAccounts ({len(accounts)}):")
            for account in accounts:
                default = " (default)" if str(account['id']) == self.config.get('default_account_id') else ""
                print(f"  {account['name']} (ID: {account['id']}){default}")
                if account.get('href'):
                    print(f"    URL: {account['href']}")
        except Exception as e:
            print(f"Error listing accounts: {e}")
    
    def cmd_account_select(self, args):
        """Handle account select command"""
        self.ensure_auth()
        
        try:
            identity = self.api.get_identity()
            accounts = [a for a in identity.get('accounts', []) if a.get('product') in ['bc3', 'bc4']]
            
            if not accounts:
                print("No Basecamp accounts found.")
                return
            
            if len(accounts) == 1:
                # Only one account, set it as default
                self.config['default_account_id'] = str(accounts[0]['id'])
                self.save_config()
                print(f"✓ Set default account: {accounts[0]['name']}")
                return
            
            # Multiple accounts, let user choose
            print("\nSelect default account:")
            for i, account in enumerate(accounts, 1):
                current = " (current default)" if str(account['id']) == self.config.get('default_account_id') else ""
                print(f"{i}. {account['name']}{current}")
            
            while True:
                choice = input("\nChoice: ").strip()
                if choice.isdigit():
                    idx = int(choice) - 1
                    if 0 <= idx < len(accounts):
                        self.config['default_account_id'] = str(accounts[idx]['id'])
                        self.save_config()
                        print(f"✓ Set default account: {accounts[idx]['name']}")
                        return
                print("Invalid choice")
                
        except Exception as e:
            print(f"Error selecting account: {e}")
    
    def cmd_project_list(self, args):
        """Handle project list command"""
        self.ensure_auth()
        
        # Get account ID
        account_id = args.account or self.config.get('default_account_id')
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        # Get projects
        try:
            projects = self.api.get_projects(account_id, get_all=True)
            
            if not projects:
                print("No projects found.")
                return
            
            print(f"\nProjects ({len(projects)}):")
            for project in projects:
                created = datetime.fromisoformat(project['created_at'].replace('Z', '+00:00')).strftime('%Y-%m-%d')
                print(f"  {project['name']} (ID: {project['id']}, created: {created})")
                if project.get('description'):
                    print(f"    {project['description']}")
        except Exception as e:
            print(f"Error listing projects: {e}")
    
    def cmd_project_show(self, args):
        """Handle project pattern matching - show project details"""
        self.ensure_auth()
        
        # Get account ID
        account_id = getattr(args, 'account', None) or self.config.get('default_account_id')
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        # Get all projects
        try:
            projects = self.api.get_projects(account_id, get_all=True)
            
            # Find matching project
            pattern = args.pattern.lower()
            matches = [p for p in projects if pattern in p['name'].lower()]
            
            if not matches:
                print(f"No project found matching '{args.pattern}'")
                return
            
            if len(matches) > 1:
                print(f"Multiple projects match '{args.pattern}':")
                for p in matches[:5]:  # Show first 5 matches
                    print(f"  - {p['name']}")
                if len(matches) > 5:
                    print(f"  ... and {len(matches) - 5} more")
                print("\nPlease be more specific.")
                return
            
            # Show project details
            project = matches[0]
            print(f"\nProject: {project['name']}")
            print(f"ID: {project['id']}")
            print(f"Created: {datetime.fromisoformat(project['created_at'].replace('Z', '+00:00')).strftime('%Y-%m-%d')}")
            if project.get('description'):
                print(f"Description: {project['description']}")
            
            # Get project details to show available tools
            full_project = self.api.get_project(account_id, project['id'])
            if 'dock' in full_project:
                tools = []
                for item in full_project['dock']:
                    if item.get('enabled'):
                        tools.append(item.get('title', item.get('name', 'Unknown')))
                if tools:
                    print(f"Tools: {', '.join(tools)}")
            
            # Set as default project
            self.config['default_project_id'] = str(project['id'])
            self.config['default_account_id'] = str(account_id)
            self.save_config()
            print(f"\n✓ Set as default project")
            
        except Exception as e:
            print(f"Error: {e}")
    
    def cmd_project_select(self, args):
        """Handle project select command"""
        self.ensure_auth()
        
        # Get account ID
        account_id = args.account or self.config.get('default_account_id')
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        # Select project
        project_id = self.select_project(account_id)
        if project_id:
            self.config['default_project_id'] = str(project_id)
            self.config['default_account_id'] = str(account_id)
            self.save_config()
            print(f"✓ Set as default project")
    
    def cmd_todo_create(self, args):
        """Handle todo create command"""
        self.ensure_auth()
        
        # Get account and project
        account_id = args.account or self.config.get('default_account_id')
        project_id = args.project or self.config.get('default_project_id')
        
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        if not project_id:
            project_id = self.select_project(account_id)
            if not project_id:
                return
        
        try:
            # Get todoset
            todoset_id = self.api.get_todoset(account_id, project_id)
            
            # Get or create todolist
            todolists = self.api.get_todolists(account_id, project_id, todoset_id)
            
            print("\nSelect a todo list or create a new one:")
            for i, tl in enumerate(todolists, 1):
                print(f"{i}. {tl['name']}")
            print(f"{len(todolists) + 1}. Create new list")
            
            choice = input("\nChoice: ").strip()
            
            if choice.isdigit():
                idx = int(choice) - 1
                if 0 <= idx < len(todolists):
                    todolist_id = todolists[idx]['id']
                elif idx == len(todolists):
                    # Create new list
                    name = input("List name: ").strip()
                    desc = input("Description (optional): ").strip()
                    new_list = self.api.create_todolist(account_id, project_id, todoset_id, name, desc)
                    todolist_id = new_list['id']
                    print(f"✓ Created list: {name}")
                else:
                    print("Invalid choice")
                    return
            else:
                print("Invalid choice")
                return
            
            # Add todos
            print("\nAdd todos (empty line to finish):")
            count = 0
            while True:
                content = input(f"Todo {count + 1}: ").strip()
                if not content:
                    break
                
                desc = input("  Description (optional): ").strip()
                
                try:
                    self.api.create_todo(account_id, project_id, todolist_id, content, desc)
                    print(f"  ✓ Added: {content}")
                    count += 1
                except Exception as e:
                    print(f"  ✗ Failed: {e}")
            
            if count > 0:
                print(f"\n✓ Added {count} todo(s)")
            
        except Exception as e:
            print(f"Error creating todos: {e}")
    
    def cmd_todo_list(self, args):
        """Handle todo list command"""
        self.ensure_auth()
        
        # Get account and project
        account_id = args.account or self.config.get('default_account_id')
        project_id = args.project or self.config.get('default_project_id')
        
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        if not project_id:
            # Try to find project by name if provided
            if args.project_name:
                projects = self.api.get_projects(account_id, get_all=False)
                matches = [p for p in projects if args.project_name.lower() in p['name'].lower()]
                if len(matches) == 1:
                    project_id = matches[0]['id']
                elif len(matches) > 1:
                    print(f"Multiple projects match '{args.project_name}':")
                    for p in matches:
                        print(f"  - {p['name']} (ID: {p['id']})")
                    return
                else:
                    print(f"No project found matching '{args.project_name}'")
                    return
            else:
                project_id = self.select_project(account_id)
                if not project_id:
                    return
        
        try:
            # Get project name
            project = self.api.get_project(account_id, project_id)
            print(f"\nTodo lists in '{project['name']}':")
            
            # Get todoset and lists
            todoset_id = self.api.get_todoset(account_id, project_id)
            todolists = self.api.get_todolists(account_id, project_id, todoset_id)
            
            if not todolists:
                print("  No todo lists found")
                return
            
            for tl in todolists:
                print(f"\n  {tl['name']}:")
                
                # Get todos for this list
                todos = self.api.get_todos(account_id, project_id, tl['id'])
                if not todos:
                    print("    (empty)")
                else:
                    for todo in todos[:10]:  # Show first 10
                        status = "✓" if todo.get('completed') else "□"
                        print(f"    {status} {todo['content']}")
                    
                    if len(todos) > 10:
                        print(f"    ... and {len(todos) - 10} more")
        
        except Exception as e:
            print(f"Error listing todos: {e}")
    
    def cmd_message_post(self, args):
        """Handle message post command"""
        self.ensure_auth()
        
        # Get account and project
        account_id = args.account or self.config.get('default_account_id')
        project_id = args.project or self.config.get('default_project_id')
        
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        if not project_id:
            project_id = self.select_project(account_id)
            if not project_id:
                return
        
        # Get message details
        title = input("Message title: ").strip()
        if not title:
            print("Title is required")
            return
        
        print("Message content (enter blank line to finish):")
        lines = []
        while True:
            line = input()
            if not line:
                break
            lines.append(line)
        
        content = '\n'.join(lines)
        if not content:
            print("Content is required")
            return
        
        try:
            self.api.post_message(account_id, project_id, title, content)
            print(f"✓ Posted message: {title}")
        except Exception as e:
            print(f"Error posting message: {e}")
    
    def cmd_campfire_post(self, args):
        """Handle campfire post command"""
        self.ensure_auth()
        
        # Get account and project
        account_id = args.account or self.config.get('default_account_id')
        project_id = args.project or self.config.get('default_project_id')
        
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        if not project_id:
            # Try to find project by name if provided
            if args.project_name:
                projects = self.api.get_projects(account_id, get_all=False)
                matches = [p for p in projects if args.project_name.lower() in p['name'].lower()]
                if len(matches) == 1:
                    project_id = matches[0]['id']
                    project_name = matches[0]['name']
                elif len(matches) > 1:
                    print(f"Multiple projects match '{args.project_name}':")
                    for p in matches:
                        print(f"  - {p['name']} (ID: {p['id']})")
                    return
                else:
                    print(f"No project found matching '{args.project_name}'")
                    return
            else:
                project_id = self.select_project(account_id)
                if not project_id:
                    return
                # Get project name
                project = self.api.get_project(account_id, project_id)
                project_name = project['name']
        else:
            # Get project name
            project = self.api.get_project(account_id, project_id)
            project_name = project['name']
        
        # Get message
        if args.message:
            # Use provided message
            content = ' '.join(args.message)
        else:
            # Interactive mode
            print(f"Post to Campfire in '{project_name}'")
            content = input("Message: ").strip()
            if not content:
                print("Message is required")
                return
        
        try:
            self.api.post_to_campfire(account_id, project_id, content)
            print(f"✓ Posted to Campfire in '{project_name}': {content}")
        except Exception as e:
            if "No campfire found" in str(e):
                print(f"Error: This project doesn't have a Campfire chat enabled")
            else:
                print(f"Error posting to Campfire: {e}")
    
    def cmd_campfire_update(self, args):
        """Handle campfire update command - posts a formatted update"""
        self.ensure_auth()
        
        # Get account and project
        account_id = args.account or self.config.get('default_account_id')
        project_id = args.project or self.config.get('default_project_id')
        
        if not account_id:
            account_id = self.select_account()
            if not account_id:
                return
        
        if not project_id:
            project_id = self.select_project(account_id)
            if not project_id:
                return
        
        # Get project name
        project = self.api.get_project(account_id, project_id)
        project_name = project['name']
        
        print(f"\nPost update to Campfire in '{project_name}'")
        
        # Get update type
        print("\nUpdate type:")
        print("1. Progress update")
        print("2. Completed task")
        print("3. Blocker/Issue")
        print("4. Question")
        print("5. Custom")
        
        update_type = input("\nChoice (1-5): ").strip()
        
        # Format message based on type
        if update_type == "1":
            prefix = "📊 Progress Update:"
            what = input("What's the update? ").strip()
            content = f"{prefix} {what}"
        elif update_type == "2":
            prefix = "✅ Completed:"
            what = input("What was completed? ").strip()
            content = f"{prefix} {what}"
        elif update_type == "3":
            prefix = "🚨 Blocker:"
            what = input("What's blocking progress? ").strip()
            content = f"{prefix} {what}"
        elif update_type == "4":
            prefix = "❓ Question:"
            what = input("What's your question? ").strip()
            content = f"{prefix} {what}"
        elif update_type == "5":
            content = input("Enter your message: ").strip()
        else:
            print("Invalid choice")
            return
        
        if not content:
            print("Message is required")
            return
        
        # Optional: Add more context
        add_context = input("\nAdd more context? (y/N): ").strip().lower()
        if add_context == 'y':
            print("Additional context (blank line to finish):")
            lines = [content, ""]  # Add blank line after main message
            while True:
                line = input()
                if not line:
                    break
                lines.append(line)
            content = '\n'.join(lines)
        
        try:
            self.api.post_to_campfire(account_id, project_id, content)
            print(f"\n✓ Posted update to Campfire")
        except Exception as e:
            if "No campfire found" in str(e):
                print(f"Error: This project doesn't have a Campfire chat enabled")
            else:
                print(f"Error posting to Campfire: {e}")
    
    def select_account(self):
        """Interactive account selection"""
        identity = self.api.get_identity()
        accounts = [a for a in identity.get('accounts', []) if a.get('product') in ['bc3', 'bc4']]
        
        if not accounts:
            print("No Basecamp accounts found")
            return None
        
        if len(accounts) == 1:
            return str(accounts[0]['id'])
        
        print("\nSelect account:")
        for i, account in enumerate(accounts, 1):
            default = " (current default)" if str(account['id']) == self.config.get('default_account_id') else ""
            print(f"{i}. {account['name']} (ID: {account['id']}){default}")
        
        while True:
            choice = input("\nChoice: ").strip()
            if choice.isdigit():
                idx = int(choice) - 1
                if 0 <= idx < len(accounts):
                    return str(accounts[idx]['id'])
            print("Invalid choice")
    
    def select_project(self, account_id):
        """Interactive project selection"""
        projects = self.api.get_projects(account_id, get_all=True)
        
        if not projects:
            print("No projects found")
            return None
        
        print("\nSelect project:")
        for i, project in enumerate(projects, 1):
            print(f"{i}. {project['name']}")
            if project.get('description'):
                print(f"   {project['description']}")
        
        while True:
            choice = input("\nChoice: ").strip()
            if choice.isdigit():
                idx = int(choice) - 1
                if 0 <= idx < len(projects):
                    return str(projects[idx]['id'])
            print("Invalid choice")

def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        prog='bc4',
        description='Basecamp 4 CLI - Manage Basecamp from the command line'
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Auth commands
    auth_parser = subparsers.add_parser('auth', help='Authentication commands')
    auth_sub = auth_parser.add_subparsers(dest='subcommand')
    
    auth_login = auth_sub.add_parser('login', help='Log in to Basecamp')
    auth_status = auth_sub.add_parser('status', help='Show authentication status')
    
    # Account commands
    account_parser = subparsers.add_parser('account', help='Account management commands')
    account_sub = account_parser.add_subparsers(dest='subcommand')
    
    account_list = account_sub.add_parser('list', help='List all accounts')
    account_select = account_sub.add_parser('select', help='Select default account')
    
    # Project commands
    project_parser = subparsers.add_parser('project', help='Project commands')
    project_sub = project_parser.add_subparsers(dest='subcommand', required=False)
    
    project_list = project_sub.add_parser('list', help='List all projects')
    project_list.add_argument('--account', help='Account ID (uses default if not specified)')
    
    project_select = project_sub.add_parser('select', help='Select default project')
    project_select.add_argument('--account', help='Account ID')
    
    # Todo commands
    todo_parser = subparsers.add_parser('todo', help='Todo commands')
    todo_sub = todo_parser.add_subparsers(dest='subcommand')
    
    todo_create = todo_sub.add_parser('create', help='Create todos interactively')
    todo_create.add_argument('--account', help='Account ID')
    todo_create.add_argument('--project', help='Project ID')
    
    todo_list = todo_sub.add_parser('list', help='List todos in a project')
    todo_list.add_argument('project_name', nargs='?', help='Project name (partial match)')
    todo_list.add_argument('--account', help='Account ID')
    todo_list.add_argument('--project', help='Project ID')
    
    # Message commands
    message_parser = subparsers.add_parser('message', help='Message board commands')
    message_sub = message_parser.add_subparsers(dest='subcommand')
    
    message_post = message_sub.add_parser('post', help='Post a message')
    message_post.add_argument('--account', help='Account ID')
    message_post.add_argument('--project', help='Project ID')
    
    # Campfire commands
    campfire_parser = subparsers.add_parser('campfire', help='Campfire chat commands')
    campfire_sub = campfire_parser.add_subparsers(dest='subcommand')
    
    campfire_post = campfire_sub.add_parser('post', help='Post to campfire')
    campfire_post.add_argument('project_name', nargs='?', help='Project name (partial match)')
    campfire_post.add_argument('message', nargs='*', help='Message to post (or leave empty for interactive)')
    campfire_post.add_argument('--account', help='Account ID')
    campfire_post.add_argument('--project', help='Project ID')
    
    campfire_update = campfire_sub.add_parser('update', help='Post a formatted update to campfire')
    campfire_update.add_argument('--account', help='Account ID')
    campfire_update.add_argument('--project', help='Project ID')
    
    # Special handling for project pattern matching
    if len(sys.argv) > 2 and sys.argv[1] == 'project' and sys.argv[2] not in ['list', 'select', '--help', '-h']:
        # Remove the pattern from argv temporarily
        pattern = sys.argv.pop(2)
        args = parser.parse_args()
        args.pattern = pattern
        args.subcommand = None
    else:
        args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    # Create CLI instance
    cli = BasecampCLI()
    
    # Route commands
    try:
        if args.command == 'auth':
            if args.subcommand == 'login':
                cli.cmd_auth_login(args)
            elif args.subcommand == 'status':
                cli.cmd_auth_status(args)
            else:
                auth_parser.print_help()
        
        elif args.command == 'account':
            if args.subcommand == 'list':
                cli.cmd_account_list(args)
            elif args.subcommand == 'select':
                cli.cmd_account_select(args)
            else:
                account_parser.print_help()
        
        elif args.command == 'project':
            # Check if we have a pattern instead of a subcommand
            if hasattr(args, 'pattern') and args.pattern and not args.subcommand:
                # Pattern matching mode
                cli.cmd_project_show(args)
            elif args.subcommand == 'list':
                cli.cmd_project_list(args)
            elif args.subcommand == 'select':
                cli.cmd_project_select(args)
            else:
                project_parser.print_help()
        
        elif args.command == 'todo':
            if args.subcommand == 'create':
                cli.cmd_todo_create(args)
            elif args.subcommand == 'list':
                cli.cmd_todo_list(args)
            else:
                todo_parser.print_help()
        
        elif args.command == 'message':
            if args.subcommand == 'post':
                cli.cmd_message_post(args)
            else:
                message_parser.print_help()
        
        elif args.command == 'campfire':
            if args.subcommand == 'post':
                cli.cmd_campfire_post(args)
            elif args.subcommand == 'update':
                cli.cmd_campfire_update(args)
            else:
                campfire_parser.print_help()
        
        else:
            parser.print_help()
    
    except KeyboardInterrupt:
        print("\nCancelled")
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()