# Implement Campfire Chat Features

## Overview
Add comprehensive Campfire chat functionality to bc4, enabling users to interact with Basecamp campfires (chat rooms) directly from the command line.

## Background
Campfires are Basecamp's chat feature, similar to Slack channels. Projects can have multiple campfires, and users need the ability to:
- List available campfires
- View recent messages
- Post new messages
- Set a default campfire per project

## Implementation Status
✅ **Completed** - Basic implementation is done and functional

### Completed Features
- [x] API client methods for campfire operations (`internal/api/campfire.go`)
- [x] Command structure (`cmd/campfire/`)
- [x] `campfire list` - List all campfires in project with table output
- [x] `campfire set [ID|name]` - Set default campfire
- [x] `campfire view [ID|name]` - View recent messages  
- [x] `campfire post <message>` - Post messages to campfire
- [x] Default campfire management per project
- [x] Integration with existing auth and config systems

### Code Structure
```
cmd/campfire/
├── campfire.go    # Root campfire command
├── list.go        # List campfires (with default indicator)
├── set.go         # Set default campfire
├── view.go        # View campfire messages
└── post.go        # Post messages

internal/api/
└── campfire.go    # API client methods
```

## Future Enhancements

### Phase 2 - Enhanced Features
- [ ] `campfire select` - Interactive campfire selection with TUI
- [ ] Interactive message composition (multi-line editor)
- [ ] `--follow` flag for live message updates
- [ ] Message pagination for viewing history
- [ ] Rich message formatting support

### Phase 3 - Advanced Features  
- [ ] File attachments
- [ ] @ mentions with autocomplete
- [ ] Emoji reactions
- [ ] Search within campfire history
- [ ] Thread/reply support
- [ ] Webhook integrations for CI/CD

## Technical Details

### API Endpoints Used
- `GET /buckets/{project_id}/chats.json` - List campfires
- `GET /buckets/{project_id}/chats/{id}.json` - Get campfire details
- `GET /buckets/{project_id}/chats/{id}/lines.json` - Get messages
- `POST /buckets/{project_id}/chats/{id}/lines.json` - Post message

### Configuration
Default campfire is stored per-project in config:
```json
{
  "accounts": {
    "123": {
      "project_defaults": {
        "456": {
          "default_campfire": "789"
        }
      }
    }
  }
}
```

## Testing Notes
- Test with projects that have multiple campfires
- Verify default campfire persistence
- Test message formatting and special characters
- Ensure proper error handling for non-existent campfires

## Related Issues
- Part of the bc4 MVP feature set
- Follows patterns established in todo commands
- Uses GitHub CLI-style table rendering system