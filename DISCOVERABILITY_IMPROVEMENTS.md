# bc4 CLI Discoverability & Usability Improvements

## Executive Summary

This document analyzes the bc4 CLI tool's current structure and provides recommendations to improve discoverability for both AI assistants (like Claude Code) and human users. The goal is to make bc4's features more easily discoverable and its usage patterns more intuitive, following GitHub CLI best practices.

## Current State Assessment

### Strengths ✅

1. **Well-structured help system** - Comprehensive help text at all levels
2. **GitHub CLI-inspired** - Follows proven command structure patterns
3. **Excellent examples** - Most commands include practical usage examples
4. **URL support** - Accepts Basecamp URLs directly, reducing cognitive load
5. **Aliases** - Provides shortcuts (e.g., `todo`/`todos`/`t`, `message`/`msg`)
6. **Interactive modes** - Offers TUI for complex operations
7. **Consistent flags** - Global flags work across all commands
8. **Good documentation** - README and SPEC.md are comprehensive

### Areas for Improvement 🔧

1. **Command organization in help output** - Flat list makes scanning difficult
2. **Basecamp-specific jargon** - Terms like "recording" may confuse new users
3. **Inconsistent example depth** - Some commands have great examples, others minimal
4. **Missing command suggestions** - No "did you mean?" or "next steps" guidance
5. **Error message clarity** - Could better guide users to solutions
6. **Shell completion prominence** - Available but not highlighted
7. **Resource relationships** - Not always clear how resources relate (e.g., comments on todos)

---

## Recommended Improvements

### Priority 1: High Impact, Low Effort

#### 1.1 Group Commands in Help Output

**Current:**
```
Available Commands:
  account     Manage Basecamp accounts
  auth        Manage authentication
  campfire    Manage campfire chats
  card        Manage card tables and cards
  comment     Work with Basecamp comments
  ...
```

**Recommended:**
```
Core Commands:
  project     Manage Basecamp projects
  todo        Work with Basecamp todos
  message     Work with Basecamp messages
  card        Manage card tables and cards

Chat & Collaboration:
  campfire    Manage campfire chats
  comment     Work with comments on resources

Account Management:
  auth        Manage authentication
  account     Manage Basecamp accounts

Other Commands:
  completion  Generate shell completions
  help        Help about any command
```

**Implementation:**
- Modify `cmd/root.go` to use Cobra's command annotations
- Add `Annotations: map[string]string{"group": "core"}` to each command
- Implement custom help template with grouped output

**Files to modify:**
- `cmd/root.go` - Add custom help template
- `cmd/project/project.go`, `cmd/todo/todo.go`, etc. - Add group annotations

---

#### 1.2 Add More Intuitive Command Aliases

**Current:** Limited aliases, some resources hard to discover

**Recommended additions:**

```go
// In cmd/comment/comment.go
Aliases: []string{"comments", "comm"}

// In cmd/campfire/campfire.go
Aliases: []string{"chat", "cf"}

// In cmd/document/document.go
Aliases: []string{"doc", "docs"}

// In cmd/card/card.go
Aliases: []string{"cards", "kanban"}
```

**Rationale:**
- Users might think "chat" before "campfire"
- "comments" (plural) feels more natural than "comment"
- Shorter aliases speed up common operations

**Files to modify:**
- `cmd/comment/comment.go`
- `cmd/campfire/campfire.go`
- `cmd/document/document.go`
- `cmd/card/card.go`

---

#### 1.3 Improve Error Messages with Next Steps

**Current:**
```
Error: no todo list specified. Use --list flag or run 'bc4 todo set' to set a default
```

**Enhanced:**
```
Error: No default todo list configured

To fix this, you can either:
  1. Set a default: bc4 todo set
  2. Specify a list: bc4 todo add "Task" --list "My List"
  3. View available lists: bc4 todo lists

Learn more: bc4 todo --help
```

**Implementation:**
- Create helper functions in `internal/errors/` for common error scenarios
- Include contextual suggestions based on the operation
- Add "Learn more" links to relevant help

**Files to modify:**
- Create `internal/errors/suggestions.go`
- Update error returns in `cmd/todo/add.go`, `cmd/card/add.go`, etc.

---

#### 1.4 Add "Quick Start" Section to Root Help

**Addition to `cmd/root.go`:**

```
Quick Start:
  bc4                          # Run first-time setup wizard
  bc4 auth status              # Check if authenticated
  bc4 project list             # See your projects
  bc4 project select           # Pick a default project
  bc4 todo lists               # View todo lists
  bc4 todo list "Tasks"        # See todos in a list
  bc4 todo add "New task"      # Create a todo

Common Workflows:
  bc4 campfire post "Update"   # Quick team update
  bc4 message post             # Announce to project
  bc4 card table "Bugs"        # View kanban board
```

---

### Priority 2: Medium Impact, Medium Effort

#### 2.1 Add Resource-Scoped Comment Commands

**Problem:** Users must know the abstract concept of "recording" to use comments

**Current:**
```bash
bc4 comment list 12345           # Requires knowing recording ID
bc4 comment create 12345         # Not intuitive
```

**Enhanced - Add convenience commands:**
```bash
# Keep existing generic commands
bc4 comment list <id>
bc4 comment create <id>

# Add resource-specific shortcuts as examples in help
bc4 todo view 123 --comments     # Show todo with comments
bc4 message view 456 --comments  # Show message with comments
bc4 card view 789 --comments     # Show card with comments

# Or consider alternative command structure
bc4 todo comment <todo-id> "Great work!"    # Direct comment on todo
bc4 message comment <msg-id> "Thanks!"      # Direct comment on message
```

**Note:** This may require architectural changes. Alternative: improve documentation to explain the "recording" concept better.

**Files to modify:**
- `cmd/todo/view.go` - Add `--comments` flag
- `cmd/message/view.go` - Add `--comments` flag
- `cmd/card/view.go` - Add `--comments` flag
- Update help text to clarify "recording" means "any commentable resource"

---

#### 2.2 Add Command Usage Examples to README

**Recommended:** Create a "Common Workflows" section with copy-pasteable commands

```markdown
## Common Workflows

### Daily Standup
```bash
# Check what's assigned to you (future feature)
bc4 todo list --assigned-to-me

# Post standup update
bc4 campfire post "Standup: Finished API work, starting tests today"
```

### Managing a Sprint
```bash
# Create sprint todo list
bc4 todo create-list "Sprint 2025-W45"

# Add tasks
bc4 todo add "Implement login" --list "Sprint 2025-W45" --due 2025-11-15
bc4 todo add "Write tests" --list "Sprint 2025-W45"

# Check progress
bc4 todo list "Sprint 2025-W45"

# Mark complete
bc4 todo check 12345
```

### Bug Triage
```bash
# View bug board
bc4 card table "Bugs"

# Add new bug
bc4 card add "Login fails with special chars" --table "Bugs"

# Move through workflow
bc4 card move 123 --column "In Progress"
bc4 card move 123 --column "Testing"
bc4 card move 123 --column "Done"
```
```

**Files to modify:**
- `README.md` - Add expanded "Common Workflows" section

---

#### 2.3 Enhance Command Descriptions

**Make descriptions more action-oriented and clear:**

**Before:**
```
comment     Work with Basecamp comments
```

**After:**
```
comment     View and add comments on todos, messages, cards, and documents
```

**Before:**
```
card        Manage card tables and cards
```

**After:**
```
card        Manage kanban boards - view cards, move through columns, track work
```

**Files to modify:**
- `cmd/comment/comment.go` - Update `Short` description
- `cmd/card/card.go` - Update `Short` description
- `cmd/campfire/campfire.go` - Consider "Post to team chat rooms"
- `cmd/document/document.go` - Consider "Create and edit project documents"

---

### Priority 3: Lower Impact, Higher Effort

#### 3.1 Add Interactive Discovery Mode

**New command:** `bc4 explore` or `bc4 wizard`

Launch an interactive TUI that:
1. Shows available commands with live preview
2. Lets users navigate by category
3. Shows examples and runs commands in preview mode
4. Helps users discover features they didn't know existed

**Implementation:**
- Create `cmd/explore/explore.go`
- Use Bubbletea for interactive navigation
- Group commands by use case

---

#### 3.2 Add Telemetry for Command Discovery

**Optional, Privacy-Conscious:**
- Track which commands users run most/least (opt-in only)
- Use insights to improve documentation and help text
- Generate "Getting Started" based on popular workflows

---

#### 3.3 Generate Context-Aware Suggestions

**After command completion, suggest next steps:**

```bash
$ bc4 todo add "Review PR"
#12345

Next steps:
  bc4 todo view 12345              # View todo details
  bc4 todo check 12345             # Mark complete when done
  bc4 comment create 12345         # Add notes or updates
```

**Implementation:**
- Create suggestion engine in `internal/suggestions/`
- Add optional suggestions after successful operations
- Make it opt-in/opt-out via config

---

## Specific Improvements for Claude Code

Claude Code specifically struggles with:
1. **Discovering available commands** - Flat help output is hard to parse
2. **Understanding resource relationships** - How comments relate to todos/messages
3. **Finding the right command for a task** - Needs better semantic mapping

### Recommendations for Claude Code

#### Add Semantic Command Descriptions

```go
// In cmd/todo/todo.go
Long: `Work with Basecamp todos and todo lists.

Basecamp projects can have multiple todo lists, each containing individual todos.
Todo lists can optionally be organized into groups for better organization.
Use these commands to navigate and manage your tasks.

COMMON TASKS:
  - View all todo lists: bc4 todo lists
  - View todos in a list: bc4 todo list "List Name"
  - Create a new todo: bc4 todo add "Task description"
  - Complete a todo: bc4 todo check 123
  - Add a comment: bc4 comment create 123

RESOURCE HIERARCHY:
  Project → Todo Lists → Todos
                      → Comments (on todos)`,
```

#### Add Machine-Readable Command Metadata

```go
// Add to command annotations
Annotations: map[string]string{
    "group": "core",
    "keywords": "tasks,checklist,work,assign,complete",
    "related": "comment,project",
    "resource": "todo",
},
```

This would help Claude Code:
- Search by keywords ("I need to manage tasks" → finds `todo`)
- Understand relationships (todo → comment)
- Group related commands

---

## Implementation Plan

### Phase 1: Quick Wins (1-2 days)
1. Add command groups to help output
2. Add more aliases
3. Enhance error messages
4. Add Quick Start to root help
5. Improve command descriptions

### Phase 2: Documentation (2-3 days)
1. Expand Common Workflows in README
2. Add troubleshooting guide
3. Create cheat sheet
4. Add more examples to each command

### Phase 3: Enhanced Discoverability (3-5 days)
1. Add resource-scoped convenience commands
2. Implement suggestion engine
3. Add `--comments` flags to view commands
4. Create explore/wizard mode

---

## Success Metrics

### Qualitative
- Users can find the right command without consulting docs
- Claude Code successfully discovers and uses commands
- Fewer "how do I...?" questions in issues/support

### Quantitative (if telemetry added)
- Reduced time to first successful command
- Increased usage of advanced features
- Reduced help command usage (means commands are clearer)

---

## Appendix: GitHub CLI Patterns to Adopt

### Pattern 1: Consistent List/View/Create/Edit
- `gh issue list` / `gh pr list`
- `gh issue view` / `gh pr view`
- `gh issue create` / `gh pr create`

✅ bc4 already follows this pattern well

### Pattern 2: Interactive by Default
- `gh pr create` is fully interactive
- Flags allow non-interactive usage

✅ bc4 supports this with some commands
❌ Could be more consistent across all create/edit commands

### Pattern 3: Smart Defaults
- `gh pr create` infers base branch
- `gh issue create` uses current repo

✅ bc4 has default project/todo list
✅ bc4 infers context from URLs

### Pattern 4: Output Modes
- `gh pr list` shows nice table in TTY
- `gh pr list` shows parseable format when piped

✅ bc4 implements this perfectly with tableprinter

### Pattern 5: Helpful Errors
- gh suggests alternatives when command fails
- gh shows next steps after success

❌ bc4 could improve here (see recommendations above)

---

## Questions for Discussion

1. **Breaking changes:** Are we willing to add new command structures (e.g., `bc4 todo comment`)?
2. **Telemetry:** Should we add opt-in usage tracking to guide improvements?
3. **Command aliases:** Which additional aliases would be most helpful?
4. **Interactive discovery:** Is an `explore` command worth the implementation effort?
5. **Documentation location:** Should we create separate "Guides" vs "Reference" docs?

---

## Conclusion

bc4 is already well-structured and follows GitHub CLI best practices. The recommendations above focus on incremental improvements to:
- Make features more discoverable
- Reduce cognitive load for new users
- Help AI assistants understand command relationships
- Provide better guidance through errors and suggestions

Most improvements are low-effort with high impact, making them good candidates for quick wins that significantly improve the user experience.
