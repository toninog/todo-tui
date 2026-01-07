# todo-tui

A terminal user interface (TUI) application for managing tasks, built with Go and Bubble Tea.

## Features

- Interactive TUI with keyboard navigation
- CLI mode for quick task operations
- Multiple view modes: Open, Completed, Shared tasks
- Task categories with color coding
- Priority and due date display
- Subtask support with progress indicators
- 10 color themes (Catppuccin, Nord, Gruvbox, Dracula, Tokyo Night, Rose Pine, Everforest, One Dark, Solarized, Kanagawa)
- 30 task completion animations
- Pagination for large task lists
- Auto-authentication with stored credentials
- AI-powered task breakdown

## Installation

### Prerequisites

- Go 1.21 or later

### Build from source

```bash
git clone https://github.com/blackraven/todo-tui.git
cd todo-tui
go build -o todo-tui ./cmd/todo-tui
```

## Usage

### Interactive TUI

```bash
./todo-tui
```

### CLI Commands

```bash
# Create a new task
./todo-tui -n "Task title"

# List all open tasks
./todo-tui -l

# Delete a task by ID
./todo-tui -d 123

# Show help
./todo-tui -h
```

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `Up/Down` or `k/j` | Move cursor |
| `Left/Right` or `h/l` | Navigate pages |
| `PgUp/PgDown` | Jump pages |
| `Tab` | Cycle views (Open/Completed/Shared) |
| `Enter` | Open task details |
| `v` | Expand/collapse task |

### Task Management

| Key | Action |
|-----|--------|
| `n` | New task |
| `e` | Edit task title |
| `E` | Edit task notes |
| `Space` | Toggle task done/open |
| `d` | Delete task |
| `c` | Change category |
| `C` | Create new category |
| `b` | AI breakdown (create subtasks) |

### Display

| Key | Action |
|-----|--------|
| `t` | Cycle themes |
| `s` | Cycle sort modes |

### Other

| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `L` | Logout |
| `r` | Refresh tasks |
| `Esc` | Cancel/back |
| `q` or `Ctrl+C` | Quit |

## Configuration

Configuration and credentials are stored in `~/.config/todo-tui/`:

- `token` - JWT authentication token
- `credentials` - Stored login credentials for auto-login

## API

The application connects to the TODO API at `https://todo.blackraven.org/api`.

## Project Structure

```
todo-tui/
  cmd/
    todo-tui/
      main.go              # Entry point
  internal/
    api/
      client.go            # HTTP client
      auth.go              # Authentication
      tasks.go             # Task operations
      categories.go        # Category operations
    config/
      config.go            # Configuration
    models/
      models.go            # App state and types
      animations.go        # Completion animations
      view.go              # View rendering
      update.go            # Event handling
    styles/
      styles.go            # Lipgloss styles
    themes/
      themes.go            # Color themes
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT
