package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/blackraven/todo-tui/internal/styles"
	"github.com/blackraven/todo-tui/internal/themes"
)

// View renders the current application state
func (m Model) View() string {
	currentTheme := m.CurrentTheme()

	switch m.State {
	case StateLogin, StateRegister:
		return m.viewAuth(currentTheme)
	case StateHelp:
		return m.viewHelp(currentTheme)
	case StateCategorySelect:
		return m.viewCategorySelect(currentTheme)
	case StateCategoryCreate:
		return m.viewCategoryCreate(currentTheme)
	case StateViewTask:
		return m.viewTaskDetail(currentTheme)
	case StateConfirmDelete:
		return m.viewConfirmDelete(currentTheme)
	default:
		return m.viewMain(currentTheme)
	}
}

// viewAuth renders the login/register screen
func (m Model) viewAuth(t themes.Theme) string {
	var title string
	if m.State == StateLogin {
		title = "// LOGIN"
	} else {
		title = "// REGISTER"
	}

	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render(title))

	// Form fields
	emailLabel := styles.InputLabelStyle.Render("Email:")
	passwordLabel := styles.InputLabelStyle.Render("Password:")

	emailField := m.EmailInput.View()
	passwordField := m.PasswordInput.View()

	// Highlight focused field
	if m.FocusedField == FieldEmail {
		emailLabel = lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("Email:")
	} else {
		passwordLabel = lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("Password:")
	}

	form := lipgloss.JoinVertical(lipgloss.Left,
		emailLabel,
		emailField,
		"",
		passwordLabel,
		passwordField,
	)

	// Error/success messages
	var msg string
	if m.ErrorMsg != "" {
		msg = styles.ErrorStyle.Render(m.ErrorMsg)
	} else if m.SuccessMsg != "" {
		msg = styles.SuccessStyle.Render(m.SuccessMsg)
	}

	if msg != "" {
		form = lipgloss.JoinVertical(lipgloss.Left, form, "", msg)
	}

	// Loading indicator
	if m.Loading {
		form = lipgloss.JoinVertical(lipgloss.Left, form, "",
			lipgloss.NewStyle().Foreground(t.Accent).Render("Loading..."))
	}

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Padding(2).
		Render(form)

	var help string
	if m.State == StateLogin {
		help = "Tab: Switch fields | Enter: Login | Ctrl+R: Register | q: Quit"
	} else {
		help = "Tab: Switch fields | Enter: Register | Ctrl+R: Login | q: Quit"
	}
	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// viewMain renders the main task list view
func (m Model) viewMain(t themes.Theme) string {
	content := m.viewList(t)

	// Build header with view mode tabs
	tabs := m.renderTabs(t)
	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top, tabs)

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Render(content)

	// Sort mode string
	sortStr := m.sortModeString()

	// Build help text
	fullHelp := fmt.Sprintf("Theme: %s (t) | Sort: %s (s) | New (n) | Edit (e) | Done (Space) | Del (d) | Category (c) | Help (?)",
		t.Name, sortStr)
	shortHelp := fmt.Sprintf("%s (t) | %s (s) | n/e/Space/d/c | ? Help", t.Name, sortStr)

	help := fullHelp
	if m.Width < 100 {
		help = shortHelp
	}

	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// renderTabs renders the view mode tabs
func (m Model) renderTabs(t themes.Theme) string {
	tabs := []string{"Open", "Completed", "Shared"}
	var rendered []string

	for i, tab := range tabs {
		var style lipgloss.Style
		if ViewMode(i) == m.ViewMode {
			style = styles.TabActiveStyle
		} else {
			style = styles.TabInactiveStyle
		}
		rendered = append(rendered, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// viewList renders the task list
func (m Model) viewList(t themes.Theme) string {
	if len(m.Tasks) == 0 && m.State != StateCreating && m.State != StateCreatingNotes {
		emptyMsg := "No tasks yet. Press 'n' to add a new task."
		if m.ViewMode == ViewCompleted {
			emptyMsg = "No completed tasks."
		} else if m.ViewMode == ViewShared {
			emptyMsg = "No shared tasks."
		}
		return styles.HelpStyle.Padding(2).Render(emptyMsg)
	}

	var s strings.Builder

	// Get paginated tasks
	pageTasks := m.PageTasks()
	pageStart := m.PageStart()

	// Layout calculations
	availableWidth := m.Width - 4
	textWidth := availableWidth - 30 // Room for number, checkbox, category, priority
	if textWidth < 10 {
		textWidth = 10
	}

	// Determine if we're creating a new task on this page
	creatingOnThisPage := false
	if m.State == StateCreating || m.State == StateCreatingNotes {
		// New task appears at the top - show on first page
		if m.Page == 0 {
			creatingOnThisPage = true
		}
	}

	// Render new task input at the top
	if creatingOnThisPage {
		var newTaskRow string
		checkIcon := lipgloss.NewStyle().Foreground(t.Accent).Render(">")

		leftBlock := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(t.Dim).Width(4).Align(lipgloss.Right).Render("*"),
			" ",
			lipgloss.NewStyle().Width(3).Align(lipgloss.Center).Render(checkIcon),
			" ",
		)

		if m.State == StateCreating {
			titleContent := styles.InlineInputStyle.Render(m.TitleInput.View())
			newTaskRow = lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, titleContent)
		} else if m.State == StateCreatingNotes {
			titleContent := lipgloss.NewStyle().Foreground(t.Fg).Render(m.TempTitle)
			titleRow := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, titleContent)
			notesIcon := lipgloss.NewStyle().Foreground(t.Dim).Render("+-")
			notesContent := styles.InlineInputStyle.Render(m.NotesInput.View())
			notesRow := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(7).Render(""),
				notesIcon,
				" ",
				notesContent,
			)
			newTaskRow = lipgloss.JoinVertical(lipgloss.Left, titleRow, notesRow)
		}

		s.WriteString(styles.ListSelectedStyle.Render(newTaskRow))
		s.WriteString("\n")
	}

	// Render tasks on this page
	for pageIdx, task := range pageTasks {
		globalIdx := pageStart + pageIdx
		selected := m.Cursor == globalIdx

		numberStr := fmt.Sprintf("%d.", globalIdx+1)
		var checkIcon string
		var titleContent string
		var categoryBadge string
		var priorityBadge string
		var dueBadge string
		var subtaskBadge string
		var notesContent string

		isEditingThis := (m.State == StateEditing && globalIdx == m.Cursor)
		isEditingNotes := (m.State == StateEditingNotes && globalIdx == m.Cursor)

		if isEditingThis {
			checkIcon = lipgloss.NewStyle().Foreground(t.Accent).Render(">")
			titleContent = styles.InlineInputStyle.Render(m.TitleInput.View())
		} else if isEditingNotes {
			checkIcon = lipgloss.NewStyle().Foreground(t.Accent).Render(">")
			titleContent = lipgloss.NewStyle().Foreground(t.Fg).Render(task.Title)
			notesContent = styles.InlineInputStyle.Render(m.NotesInput.View())
		} else {
			// Checkbox
			if task.Status == "done" {
				checkIcon = lipgloss.NewStyle().Foreground(t.Success).Render("[x]")
			} else {
				checkIcon = lipgloss.NewStyle().Foreground(t.Accent).Render("[ ]")
			}

			// Title with animations
			var rawTitle string
			if task.IsDeleting {
				rawTitle = RenderDeleteAnim(task.Title, t)
			} else if task.IsAnimatingCheck {
				rawTitle = RenderCheckAnim(task, t)
			} else if task.Status == "done" {
				rawTitle = styles.StrikeStyle.Render(task.Title)
			} else {
				rawTitle = lipgloss.NewStyle().Foreground(t.Fg).Render(task.Title)
			}

			// Expansion indicator
			displayTitle := rawTitle
			if (task.Notes != nil && *task.Notes != "") || len(task.Subtasks) > 0 {
				arrow := " >"
				if task.Expanded {
					arrow = " v"
				}
				displayTitle += lipgloss.NewStyle().Foreground(t.Accent).Render(arrow)
			}

			titleContent = lipgloss.NewStyle().MaxWidth(textWidth).Render(displayTitle)

			// Category badge
			if task.Category != nil {
				catColor := lipgloss.Color(task.Category.Color)
				categoryBadge = lipgloss.NewStyle().
					Foreground(t.Bg).
					Background(catColor).
					Padding(0, 1).
					Render(task.Category.Name)
			}

			// Priority badge
			if task.Priority > 0 {
				var prioStyle lipgloss.Style
				if task.Priority >= 8 {
					prioStyle = styles.PriorityHighStyle
				} else if task.Priority >= 5 {
					prioStyle = styles.PriorityMedStyle
				} else {
					prioStyle = styles.PriorityLowStyle
				}
				priorityBadge = prioStyle.Render(fmt.Sprintf("P%d", task.Priority))
			}

			// Due date badge
			if task.DueAt != nil {
				dueStr := formatDue(*task.DueAt)
				if task.DueAt.Before(time.Now()) && task.Status != "done" {
					dueBadge = styles.OverdueStyle.Render(dueStr)
				} else {
					dueBadge = styles.DueStyle.Render(dueStr)
				}
			}

			// Subtask progress
			if len(task.Subtasks) > 0 {
				done := 0
				for _, st := range task.Subtasks {
					if st.Status == "done" {
						done++
					}
				}
				subtaskBadge = lipgloss.NewStyle().Foreground(t.Dim).
					Render(fmt.Sprintf("[%d/%d]", done, len(task.Subtasks)))
			}
		}

		// Build left block (number + checkbox)
		leftBlock := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(t.Dim).Width(4).Align(lipgloss.Right).Render(numberStr),
			" ",
			lipgloss.NewStyle().Width(3).Align(lipgloss.Center).Render(checkIcon),
			" ",
		)

		// Build right block (badges)
		var badges []string
		if categoryBadge != "" {
			badges = append(badges, categoryBadge)
		}
		if priorityBadge != "" {
			badges = append(badges, priorityBadge)
		}
		if dueBadge != "" {
			badges = append(badges, dueBadge)
		}
		if subtaskBadge != "" {
			badges = append(badges, subtaskBadge)
		}
		rightBlock := strings.Join(badges, " ")

		var row string
		if isEditingNotes {
			titleRow := lipgloss.JoinHorizontal(lipgloss.Top,
				leftBlock,
				titleContent,
			)
			notesIcon := lipgloss.NewStyle().Foreground(t.Dim).Render("+-")
			notesRow := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(7).Render(""),
				notesIcon,
				" ",
				notesContent,
			)
			row = lipgloss.JoinVertical(lipgloss.Left, titleRow, notesRow)
		} else {
			row = lipgloss.JoinHorizontal(lipgloss.Top,
				leftBlock,
				titleContent,
				"  ",
				rightBlock,
			)
		}

		if selected {
			s.WriteString(styles.ListSelectedStyle.Render(row))
		} else {
			s.WriteString(styles.ListItemStyle.Render(row))
		}
		s.WriteString("\n")

		// Render expanded content (notes and subtasks)
		if !isEditingThis && task.Expanded {
			// Notes
			if task.Notes != nil && *task.Notes != "" {
				notesIcon := lipgloss.NewStyle().Foreground(t.Dim).Render("+-")
				notesText := *task.Notes
				if task.IsDeleting {
					notesText = RenderDeleteAnim(notesText, t)
				} else if task.Status == "done" {
					notesText = styles.StrikeStyle.Italic(true).Render(notesText)
				} else {
					notesText = lipgloss.NewStyle().Foreground(t.Fg).Italic(true).Render(notesText)
				}

				notesRow := lipgloss.JoinHorizontal(lipgloss.Top,
					lipgloss.NewStyle().Width(7).Render(""),
					notesIcon,
					" ",
					lipgloss.NewStyle().Width(textWidth-3).Render(notesText),
				)

				if selected {
					s.WriteString(styles.ListSelectedStyle.Render(notesRow))
				} else {
					s.WriteString(styles.ListItemStyle.Render(notesRow))
				}
				s.WriteString("\n")
			}

			// Subtasks
			for j, subtask := range task.Subtasks {
				var subIcon string
				if subtask.Status == "done" {
					subIcon = lipgloss.NewStyle().Foreground(t.Success).Render("[x]")
				} else {
					subIcon = lipgloss.NewStyle().Foreground(t.Dim).Render("[ ]")
				}

				connector := "|-"
				if j == len(task.Subtasks)-1 {
					connector = "+-"
				}

				subTitle := subtask.Title
				if subtask.Status == "done" {
					subTitle = styles.StrikeStyle.Render(subTitle)
				}

				subRow := lipgloss.JoinHorizontal(lipgloss.Top,
					lipgloss.NewStyle().Width(7).Render(""),
					lipgloss.NewStyle().Foreground(t.Dim).Render(connector),
					" ",
					subIcon,
					" ",
					subTitle,
				)

				if selected {
					s.WriteString(styles.ListSelectedStyle.Render(subRow))
				} else {
					s.WriteString(styles.ListItemStyle.Render(subRow))
				}
				s.WriteString("\n")
			}
		}
	}

	// Add pagination info if there are multiple pages
	totalPages := m.TotalPages()
	if totalPages > 1 {
		s.WriteString("\n")
		pageInfo := fmt.Sprintf("Page %d/%d", m.Page+1, totalPages)
		navHint := ""
		if m.Page > 0 && m.Page < totalPages-1 {
			navHint = " | < prev | next >"
		} else if m.Page > 0 {
			navHint = " | < prev"
		} else if m.Page < totalPages-1 {
			navHint = " | next >"
		}
		s.WriteString(styles.HelpStyle.Render(pageInfo + navHint))
	}

	return s.String()
}

// viewCategorySelect renders the category selection overlay
func (m Model) viewCategorySelect(t themes.Theme) string {
	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render("// SELECT CATEGORY"))

	var s strings.Builder
	s.WriteString("\n")

	// "None" option
	noneRow := "  [ ] None (remove category)"
	if m.CategoryCursor == -1 {
		s.WriteString(styles.ListSelectedStyle.Render(noneRow))
	} else {
		s.WriteString(styles.ListItemStyle.Render(noneRow))
	}
	s.WriteString("\n")

	for i, cat := range m.Categories {
		catColor := lipgloss.Color(cat.Color)
		colorSwatch := lipgloss.NewStyle().Background(catColor).Render("  ")
		catName := lipgloss.NewStyle().Foreground(t.Fg).Render(cat.Name)

		row := fmt.Sprintf("  %s %s", colorSwatch, catName)

		if m.CategoryCursor == i {
			s.WriteString(styles.ListSelectedStyle.Render(row))
		} else {
			s.WriteString(styles.ListItemStyle.Render(row))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(styles.HelpStyle.Render("  Press 'C' to create new category"))

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Render(s.String())

	help := "Enter: Select | C: Create New | Esc: Cancel"
	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// viewCategoryCreate renders the category creation form
func (m Model) viewCategoryCreate(t themes.Theme) string {
	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render("// CREATE CATEGORY"))

	nameLabel := styles.InputLabelStyle.Render("Name:")
	nameField := m.CategoryInput.View()

	form := lipgloss.JoinVertical(lipgloss.Left,
		nameLabel,
		nameField,
		"",
		styles.HelpStyle.Render("Color will be assigned automatically"),
	)

	if m.ErrorMsg != "" {
		form = lipgloss.JoinVertical(lipgloss.Left, form, "",
			styles.ErrorStyle.Render(m.ErrorMsg))
	}

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Padding(2).
		Render(form)

	help := "Enter: Create | Esc: Cancel"
	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// viewTaskDetail renders the task detail view
func (m Model) viewTaskDetail(t themes.Theme) string {
	if m.SelectedTaskIdx < 0 || m.SelectedTaskIdx >= len(m.Tasks) {
		return m.viewMain(t)
	}

	task := m.Tasks[m.SelectedTaskIdx]

	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render("// TASK DETAILS"))

	var s strings.Builder

	// Title
	titleLabel := styles.InputLabelStyle.Render("Title:")
	s.WriteString(titleLabel + "\n")
	s.WriteString(lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(task.Title))
	s.WriteString("\n\n")

	// Status and priority
	statusStr := "Open"
	if task.Status == "done" {
		statusStr = "Completed"
	}
	s.WriteString(fmt.Sprintf("Status: %s | Priority: %d\n",
		lipgloss.NewStyle().Foreground(t.Accent).Render(statusStr),
		task.Priority))

	// Category
	if task.Category != nil {
		catColor := lipgloss.Color(task.Category.Color)
		catBadge := lipgloss.NewStyle().Background(catColor).Foreground(t.Bg).Padding(0, 1).Render(task.Category.Name)
		s.WriteString(fmt.Sprintf("Category: %s\n", catBadge))
	}

	// Due date
	if task.DueAt != nil {
		s.WriteString(fmt.Sprintf("Due: %s\n", task.DueAt.Format("Jan 2, 2006 3:04 PM")))
	}

	s.WriteString("\n")

	// Notes
	if task.Notes != nil && *task.Notes != "" {
		notesLabel := styles.InputLabelStyle.Render("Notes:")
		s.WriteString(notesLabel + "\n")
		s.WriteString(lipgloss.NewStyle().Foreground(t.Fg).Italic(true).Render(*task.Notes))
		s.WriteString("\n\n")
	}

	// Subtasks
	if len(task.Subtasks) > 0 {
		subtasksLabel := styles.InputLabelStyle.Render("Subtasks:")
		s.WriteString(subtasksLabel + "\n")
		for _, st := range task.Subtasks {
			var icon string
			if st.Status == "done" {
				icon = "[x]"
			} else {
				icon = "[ ]"
			}
			stTitle := st.Title
			if st.Status == "done" {
				stTitle = styles.StrikeStyle.Render(stTitle)
			}
			s.WriteString(fmt.Sprintf("  %s %s\n", icon, stTitle))
		}
	}

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Padding(1).
		Render(s.String())

	help := "e: Edit | Space: Toggle Done | b: Breakdown | c: Category | Esc: Back"
	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// viewConfirmDelete renders the delete confirmation dialog
func (m Model) viewConfirmDelete(t themes.Theme) string {
	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render("// CONFIRM DELETE"))

	var taskTitle string
	if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
		taskTitle = m.Tasks[m.SelectedTaskIdx].Title
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		styles.ErrorStyle.Render("Are you sure you want to delete this task?"),
		"",
		lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(taskTitle),
		"",
		"",
		styles.HelpStyle.Render("Press 'y' to confirm, 'n' or Esc to cancel"),
	)

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Warning).
		Width(m.Width - 4).
		Height(containerHeight).
		Padding(2).
		Render(content)

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// viewHelp renders the help overlay
func (m Model) viewHelp(t themes.Theme) string {
	header := lipgloss.Place(m.Width, 1, lipgloss.Center, lipgloss.Top,
		styles.HeaderStyle.Render("// HELP"))

	var s strings.Builder

	s.WriteString(styles.InputLabelStyle.Render("Navigation:") + "\n")
	s.WriteString("  Up/Down, k/j    Move cursor\n")
	s.WriteString("  Left/Right, h/l Navigate pages\n")
	s.WriteString("  PgUp/PgDown     Jump pages\n")
	s.WriteString("  Tab             Cycle views (Open/Completed/Shared)\n")
	s.WriteString("  Enter           Open task details\n")
	s.WriteString("  v               Expand/collapse task\n\n")

	s.WriteString(styles.InputLabelStyle.Render("Task Management:") + "\n")
	s.WriteString("  n               New task\n")
	s.WriteString("  e               Edit task title\n")
	s.WriteString("  E               Edit task notes\n")
	s.WriteString("  Space           Toggle task done/open\n")
	s.WriteString("  d               Delete task\n")
	s.WriteString("  c               Change category\n")
	s.WriteString("  C               Create new category\n")
	s.WriteString("  b               AI breakdown (create subtasks)\n\n")

	s.WriteString(styles.InputLabelStyle.Render("Display:") + "\n")
	s.WriteString("  t               Cycle themes\n")
	s.WriteString("  s               Cycle sort modes\n\n")

	s.WriteString(styles.InputLabelStyle.Render("Other:") + "\n")
	s.WriteString("  ?               Toggle this help\n")
	s.WriteString("  L               Logout\n")
	s.WriteString("  Esc             Cancel/back\n")
	s.WriteString("  Ctrl+C, q       Quit\n")

	containerHeight := m.Height - 7
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Width(m.Width - 4).
		Height(containerHeight).
		Padding(1).
		Render(s.String())

	help := "Press ? to close"
	status := lipgloss.NewStyle().Width(m.Width).Align(lipgloss.Center).
		Render(styles.HelpStyle.Render(help))

	ui := lipgloss.JoinVertical(lipgloss.Center, header, container, status)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, ui)
}

// sortModeString returns a string representation of the sort mode
func (m Model) sortModeString() string {
	switch m.SortMode {
	case SortPriority:
		return "Priority"
	case SortDueDate:
		return "Due Date"
	case SortAlphabetical:
		return "A-Z"
	case SortCreated:
		return "Created"
	}
	return "Off"
}

// formatDue formats a due date relative to now
func formatDue(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		// Overdue
		diff = -diff
		if diff < 24*time.Hour {
			return "Overdue"
		}
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd overdue", days)
	}

	if diff < time.Hour {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "Tomorrow"
	}
	if days < 7 {
		return fmt.Sprintf("%dd", days)
	}
	return t.Format("Jan 2")
}
