# Basecamp 3/4 API Implementation Status

This document tracks which Basecamp API features have been implemented in bc4 and which are still pending.

**API Documentation:** https://github.com/basecamp/bc3-api

## Summary

| Status | Count |
|--------|-------|
| Fully Implemented | 16 |
| Partially Implemented | 3 |
| Not Implemented | 20 |
| **Total API Resources** | **39** |

---

## Fully Implemented

These resources have complete or near-complete API coverage.

| Resource | Operations | Notes |
|----------|------------|-------|
| **To-dos** | list, get, create, update, check, uncheck, move, reposition | Complete with attachments |
| **To-do lists** | list, get, create, update | Full support |
| **To-do list groups** | get, create, reposition | Full support |
| **To-do sets** | Accessed via todo lists | Implicit support |
| **Comments** | list, get, create, edit, delete | Works on todos, messages, documents, cards |
| **Messages** | list, get, create, edit, delete, pin, unpin | Full support |
| **Message boards** | list messages | Via message commands |
| **Campfires** | list, get, post, delete lines | Full chat support |
| **Card tables** | get | Board level operations |
| **Card table cards** | list, get, create, edit, move, archive, assign/unassign | Full kanban support |
| **Card table columns** | list, create, edit, move, color, on-hold | Full column management |
| **Card table steps** | list, add, edit, delete, move, check, uncheck, assign | Card checklist items |
| **Documents** | list, get, create, edit | Full CRUD (API limits delete) |
| **Vaults** | list documents | Document containers |
| **Uploads/Attachments** | upload, attach to resources | Full file support |
| **People** | my profile, get by ID, list in project | User info |

---

## Partially Implemented

These resources have some functionality but are missing operations.

| Resource | What Works | What's Missing |
|----------|------------|----------------|
| **Projects** | list, get, search, select default | create, update, delete, archive, trash |
| **Message types** | list categories | create, update, delete categories |
| **People** | view/list people | invite, update roles, remove from project |

---

## Not Implemented

These resources have no implementation yet.

| Resource | API Documentation | Description |
|----------|-------------------|-------------|
| **Schedules** | [schedules.md](https://github.com/basecamp/bc3-api/blob/master/sections/schedules.md) | Calendar/schedule containers |
| **Schedule entries** | [schedule_entries.md](https://github.com/basecamp/bc3-api/blob/master/sections/schedule_entries.md) | Calendar events and milestones |
| **Chatbots** | [chatbots.md](https://github.com/basecamp/bc3-api/blob/master/sections/chatbots.md) | Automated chat integration |
| **Client approvals** | [client_approvals.md](https://github.com/basecamp/bc3-api/blob/master/sections/client_approvals.md) | Client sign-off workflows |
| **Client correspondences** | [client_correspondences.md](https://github.com/basecamp/bc3-api/blob/master/sections/client_correspondences.md) | Client email threads |
| **Client replies** | [client_replies.md](https://github.com/basecamp/bc3-api/blob/master/sections/client_replies.md) | Responses to client correspondence |
| **Client visibility** | [client_visibility.md](https://github.com/basecamp/bc3-api/blob/master/sections/client_visibility.md) | Control what clients can see |
| **Events** | [events.md](https://github.com/basecamp/bc3-api/blob/master/sections/events.md) | Activity stream/history |
| **Forwards** | [forwards.md](https://github.com/basecamp/bc3-api/blob/master/sections/forwards.md) | Email forwarding |
| **Inbox replies** | [inbox_replies.md](https://github.com/basecamp/bc3-api/blob/master/sections/inbox_replies.md) | Email thread replies |
| **Inboxes** | [inboxes.md](https://github.com/basecamp/bc3-api/blob/master/sections/inboxes.md) | Email-in functionality |
| **Lineup Markers** | [lineup_markers.md](https://github.com/basecamp/bc3-api/blob/master/sections/lineup_markers.md) | Timeline visual markers |
| **Question answers** | [question_answers.md](https://github.com/basecamp/bc3-api/blob/master/sections/question_answers.md) | Responses to automatic check-ins |
| **Questionnaires** | [questionnaires.md](https://github.com/basecamp/bc3-api/blob/master/sections/questionnaires.md) | Automatic check-ins container |
| **Questions** | [questions.md](https://github.com/basecamp/bc3-api/blob/master/sections/questions.md) | Individual check-in questions |
| **Recordings** | [recordings.md](https://github.com/basecamp/bc3-api/blob/master/sections/recordings.md) | Recording management (archive/trash) |
| **Subscriptions** | [subscriptions.md](https://github.com/basecamp/bc3-api/blob/master/sections/subscriptions.md) | Notification preferences |
| **Templates** | [templates.md](https://github.com/basecamp/bc3-api/blob/master/sections/templates.md) | Project templates |
| **Timesheets** | [timesheets.md](https://github.com/basecamp/bc3-api/blob/master/sections/timesheets.md) | Time tracking |
| **Webhooks** | [webhooks.md](https://github.com/basecamp/bc3-api/blob/master/sections/webhooks.md) | Event subscriptions |

---

## Implementation Priority Recommendations

### High Value (Commonly Used Features)

1. **Schedules & Schedule entries** - Calendar functionality is a core Basecamp feature used by most teams
2. **Events** - Activity feed useful for monitoring project changes
3. **Subscriptions** - Allow users to manage their notification preferences
4. **Recordings** - Bulk archive/trash operations for cleanup

### Medium Value

5. **Questionnaires/Questions/Answers** - Automatic check-ins feature
6. **Webhooks** - Enable integrations and automation
7. **Project CRUD** - Create/update/delete/archive projects

### Lower Priority (Specialized Use Cases)

8. **Client visibility/approvals** - Only relevant for client-facing projects
9. **Chatbots** - Automation/bot integration
10. **Inboxes/Forwards** - Email integration features
11. **Templates** - Project template management
12. **Timesheets** - Time tracking (if enabled on account)
13. **Lineup Markers** - Visual timeline markers

---

## Current Implementation Strengths

The bc4 CLI has excellent coverage for core productivity features:

- **Task Management** - Complete todo support with groups, positioning, assignments, and attachments
- **Kanban Boards** - Full card table support with cards, columns, steps, and workflow states
- **Team Communication** - Messages with pin/unpin and real-time campfire chat
- **Document Management** - Create, edit, and organize documents in vaults
- **Comments** - Universal comment support across all resource types
- **File Attachments** - Upload and attach files to any supported resource

### Main Gaps

- **Scheduling/Calendar** - No support for schedules or calendar events
- **Administrative Features** - Webhooks, templates, client features not implemented
- **Activity Tracking** - No events/activity stream support

---

## Command Reference

### Todos
```
bc4 todo list [todolist-id]
bc4 todo get <todo-id>
bc4 todo create <todolist-id> <content>
bc4 todo edit <todo-id>
bc4 todo check <todo-id>
bc4 todo uncheck <todo-id>
bc4 todo move <todo-id> --to <todolist-id>
bc4 todo reposition <todo-id> --position <n>
bc4 todolist list
bc4 todolist get <todolist-id>
bc4 todolist create <name>
bc4 todogroup create <todolist-id> <name>
```

### Messages
```
bc4 message list
bc4 message get <message-id>
bc4 message create <subject>
bc4 message edit <message-id>
bc4 message delete <message-id>
bc4 message pin <message-id>
bc4 message unpin <message-id>
```

### Campfire
```
bc4 campfire list
bc4 campfire get <campfire-id>
bc4 campfire post <content>
bc4 campfire lines [campfire-id]
bc4 campfire delete <line-id>
```

### Cards
```
bc4 card list [column-id]
bc4 card get <card-id>
bc4 card create <column-id> <title>
bc4 card edit <card-id>
bc4 card move <card-id> --to <column-id>
bc4 card archive <card-id>
bc4 card assign <card-id> <person-id>
bc4 card unassign <card-id> <person-id>
bc4 column list
bc4 column create <name>
bc4 column edit <column-id>
bc4 column move <column-id> --position <n>
bc4 column on-hold <column-id>
bc4 column remove-hold <column-id>
bc4 step list <card-id>
bc4 step add <card-id> <name>
bc4 step check <step-id>
bc4 step uncheck <step-id>
```

### Documents
```
bc4 document list
bc4 document get <document-id>
bc4 document create <title>
bc4 document edit <document-id>
```

### Comments
```
bc4 comment list <resource-type> <resource-id>
bc4 comment get <comment-id>
bc4 comment create <resource-type> <resource-id>
bc4 comment edit <comment-id>
bc4 comment delete <comment-id>
```

### Projects & Accounts
```
bc4 project list
bc4 project get <project-id>
bc4 project search <pattern>
bc4 project select
bc4 account list
bc4 account current
bc4 account select
```

---

*Last updated: 2026-01-07*
