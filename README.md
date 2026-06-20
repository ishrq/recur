# Recur

A powerful CLI todo application with support for recurring tasks, natural language date parsing, and advanced filtering. Written in Go.

## Features

- 📅 **Natural language date parsing** - Use `tomorrow`, `next friday`, `+3d`, or ISO dates
- 🔄 **Recurring tasks** - Support for hourly, daily, weekly, monthly, and yearly recurrence
- 🏷️ **Rich task attributes** - Tags, projects, priorities, and notes
- 🔍 **Powerful filtering** - Filter by date, tag, project, priority, or keyword
- 📤 **CSV export** - Export filtered tasks to CSV
- ✏️ **Editor integration** - Bulk edit tasks in your `$EDITOR`
- 🗑️ **Undo operations** - Restore deleted or completed tasks
- 📊 **Dashboard view** - See overdue, today's, and upcoming tasks at a glance

## Installation

### From Source

**Prerequisites:**
- Go 1.21 or later
- SQLite with FTS5 support (usually included by default)

```bash
# Clone the repository
git clone https://github.com/yourusername/recur.git
cd recur

# Build the binary
make build

# (Optional) Install to $GOPATH/bin
make install
```

### Manual Build

```bash
go build -tags fts5 -o recur cmd/recur/main.go
```

### Binary Distribution

Download pre-built binaries from the [releases page](https://github.com/yourusername/recur/releases).

**Linux/macOS:**
```bash
# Extract and install
tar -xzf recur-*-linux-amd64.tar.gz
sudo mv recur /usr/local/bin/
```

**Verify installation:**
```bash
recur help
```

## Quick Start

```bash
# Add a simple task
recur add "Buy groceries"

# Add a task with a due date
recur add "Team meeting @(tomorrow 3pm)"

# Add a recurring task
recur add "Take vitamins @(today 8am, daily)"

# List today's tasks
recur ls --today

# Mark a task as done
recur done 1

# View all tasks
recur ls --all
```

## Task Syntax

Tasks support inline attributes using a simple syntax:

```
recur add "Task name @(date time, frequency, end) #tag +project !priority *note"
```

### Date and Time (`@(...)`)

**Semantic dates:**
```bash
recur add "Call mom @(tomorrow)"
recur add "Meeting @(friday 2pm)"
recur add "Standup @(today 9am)"
```

**Relative dates:**
```bash
recur add "Follow up @(+3d)"        # 3 days from now
recur add "Review notes @(-1w)"      # 1 week ago
recur add "Plan trip @(+2m)"         # 2 months from now
```

**Standard formats:**
```bash
recur add "Deadline @(2025-12-31)"
recur add "Appointment @(2025-11-20 14:30)"
recur add "Call @(Jan 15 3pm)"
```

**Recurring tasks:**
```bash
recur add "Backup files @(monday 6pm, weekly)"
recur add "Pay rent @(2025-12-01, monthly, 2026-12-01)"
recur add "Quarterly review @(2025-12-31, 3m)"
```

### Tags (`#tag`)

```bash
recur add "Fix bug #work #urgent"
recur add "Read book #leisure #personal"
```

### Projects (`+project`)

```bash
recur add "Write report +Work"
recur add "Plan vacation +Personal +Travel"
```

### Priorities (`!priority`)

```bash
recur add "Submit proposal !urgent"
recur add "Review PR !high"
recur add "Clean desk !low"
```

### Notes (`*note`)

```bash
recur add "Dentist appointment *'Bring insurance card'"
recur add "Meeting *'Prepare slides on Q4 results'"
```

## Commands

### add - Add tasks

```bash
# Simple task
recur add "Task name"

# Task with attributes
recur add "Team meeting @(friday 2pm) #work +Office !high *'Prepare agenda'"

# Open editor for bulk adding
recur add --edit

# Open editor with prefilled content
recur add --edit "Task name #tag"
```

### ls - List tasks

```bash
# Dashboard view (default)
recur ls

# View specific groups
recur ls --today
recur ls --tomorrow
recur ls --overdue
recur ls --upcoming              # Next 7 days

# View all tasks
recur ls --all
recur ls --done
recur ls --trash

# Filter by attributes
recur ls --tag work
recur ls --project Home
recur ls --priority urgent high

# Filter by date
recur ls --due 2025-11-15
recur ls --due tomorrow
recur ls --from today --to 2025-11-20

# Search by keyword
recur ls --query meeting

# List metadata
recur ls --tags                  # Show all tags with counts
recur ls --projects              # Show all projects with counts
recur ls --priorities            # Show all priorities with counts

# Export to CSV
recur ls --export                           # Auto-generated filename
recur ls --today --export                   # recur_today.csv
recur ls --tag work --export work_tasks.csv # Custom filename
```

### done - Complete tasks

```bash
# Complete by ID
recur done 1
recur done 1 2 3

# Complete by filter
recur done --today
recur done --tag work
recur done --project Home
recur done --due today

# Undo completion
recur done --undo 1 2 3
recur done --undo --tag work
```

### rm - Delete tasks

```bash
# Delete by ID
recur rm 1
recur rm 1 2 3

# Delete by filter
recur rm --tag work
recur rm --project Home
recur rm --due today

# Delete groups
recur rm --all                   # All incomplete tasks
recur rm --done                  # All completed tasks
recur rm --trash                 # Permanently delete trashed tasks

# Restore deleted tasks
recur rm --undo 1 2 3
recur rm --undo --tag work
```

⚠️ **Warning:** `recur rm --trash` and `recur rm --purge` permanently delete tasks and cannot be undone.

### cp - Copy tasks

```bash
# Duplicate a task
recur cp 1

# Duplicate with modifications
recur cp 5 "Modified copy @(tomorrow) !urgent"

# Edit before copying
recur cp --edit 1 2 3

# Copy by filter
recur cp --tag work
recur cp --project Home
recur cp --due today
```

### mv - Edit tasks

```bash
# Edit with new content
recur mv 1 "Updated task name"
recur mv 3 "@(tomorrow 3pm)"
recur mv 5 "New name +NewProject"

# Edit in $EDITOR
recur mv 1
recur mv 1 2 3

# Edit by filter
recur mv --tag work
recur mv --tag work "@(tomorrow)"
recur mv --due today "!urgent"
```

## Date Parsing Reference

Run `recur help dates` for a comprehensive guide to all supported date formats.

**Quick reference:**

| Format | Example | Result |
|--------|---------|--------|
| Semantic | `tomorrow`, `friday` | Next occurrence at 12:00 PM |
| Relative | `+3d`, `-1w`, `+2m` | Relative to current time |
| ISO | `2025-11-20` | Specified date at 12:00 PM |
| ISO with time | `2025-11-20 15:04` | Specified date and time |
| Month/day | `Jan 20`, `January 20 3pm` | Current year assumed |
| Frequency | `1d`, `2w`, `1m`, `daily`, `weekly` | Recurring interval |

## Configuration

Recur stores its database at:
- **Linux/macOS:** `~/.local/share/recur/recur.db`
- **Windows:** `%LOCALAPPDATA%\recur\recur.db`

Set `$EDITOR` environment variable for editor integration:
```bash
export EDITOR=vim
# or
export EDITOR=nano
# or
export EDITOR=code --wait
```

## Examples

### Daily workflow
```bash
# Morning: check today's tasks
recur ls --today

# Add a new task
recur add "Reply to emails @(today 10am) #work !high"

# Mark tasks as done throughout the day
recur done 5 8 12

# Evening: review what's coming up
recur ls --tomorrow
recur ls --upcoming
```

### Weekly planning
```bash
# Review overdue tasks
recur ls --overdue

# Set up recurring tasks
recur add "Weekly team sync @(monday 10am, weekly) +Work"
recur add "Grocery shopping @(saturday 2pm, weekly) #errands"
recur add "Review goals @(sunday 8pm, weekly) +Personal"

# Plan the week ahead
recur ls --from today --to +7d
```

### Project management
```bash
# Add project tasks
recur add "Design mockups @(+2d) +Website #design !high"
recur add "Write copy @(+3d) +Website #content !medium"
recur add "Set up hosting @(+5d) +Website #dev !high"

# View project tasks
recur ls --project Website

# Export project for reporting
recur ls --project Website --export website_tasks.csv
```

### Batch operations
```bash
# Complete all work tasks due today
recur done --tag work --due today

# Delete all low priority tasks
recur rm --priority low

# Move all personal tasks to tomorrow
recur mv --tag personal "@(tomorrow)"

# Duplicate work template for next week
recur cp --tag template --project Work
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Generate coverage report
make test-coverage
```

### Building

```bash
# Build binary
make build

# Install to $GOPATH/bin
make install

# Create distribution packages
make package-all
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [SQLite](https://www.sqlite.org/) with FTS5 for full-text search
- Inspired by [todo.txt](http://todotxt.org/) and [taskwarrior](https://taskwarrior.org/)
