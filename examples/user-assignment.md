# User Assignment Examples

The bc4 CLI now supports flexible user assignment for todos and cards. You can assign users by:

## 1. Email Address
```bash
# Assign todo to user by email
bc4 todo add "Review documentation" --assign john@example.com

# Multiple assignees
bc4 todo add "Team meeting prep" --assign john@example.com,jane@example.com

# Card assignment
bc4 card add "Update website" --assign webmaster@company.com
```

## 2. @Mentions
```bash
# Assign by @mention (matches name)
bc4 todo add "Fix login bug" --assign @john

# Case-insensitive matching
bc4 todo add "Deploy to staging" --assign @John

# Partial name matching (first or last name)
bc4 todo add "Security audit" --assign @smith

# Multiple assignees with @mentions
bc4 todo add "Design review" --assign @john,@jane
```

## 3. Mixed Identifiers
```bash
# Mix emails and @mentions
bc4 todo add "Project kickoff" --assign @john,jane@example.com,@bob
```

## Error Handling
If a user cannot be found, you'll get a clear error message:
```bash
$ bc4 todo add "Task" --assign @unknown
Error: failed to resolve assignees: could not find users: @unknown
```

## How It Works
1. The CLI fetches all people in the project
2. For email addresses: Exact match (case-insensitive)
3. For @mentions: Tries exact name match, then first/last name, then partial match
4. Shows clear errors if users cannot be found
5. Avoids duplicate assignments