package models

import (
	"math/rand"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/blackraven/todo-tui/internal/api"
	"github.com/blackraven/todo-tui/internal/styles"
	"github.com/blackraven/todo-tui/internal/themes"
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.TitleInput.Width = msg.Width - 20
		m.NotesInput.Width = msg.Width - 20
		m.EmailInput.Width = min(40, msg.Width-20)
		m.PasswordInput.Width = min(40, msg.Width-20)

	case TasksLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
			// Check if unauthorized
			if apiErr, ok := msg.Err.(*api.APIError); ok && apiErr.IsUnauthorized() {
				m.State = StateLogin
				m.Client.ClearToken()
			}
		} else {
			m.ErrorMsg = ""
			m.Tasks = make([]Task, len(msg.Tasks))
			for i, t := range msg.Tasks {
				m.Tasks[i] = Task{Task: t}
			}
			m.ApplySort()
		}
		m.ValidateCursor()
		m.EnsureCursorVisible()

	case CategoriesLoadedMsg:
		if msg.Err == nil {
			m.Categories = msg.Categories
		}

	case TaskCreatedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
		} else if msg.Task != nil {
			m.SuccessMsg = "Task created"
			newTask := Task{Task: *msg.Task}
			m.Tasks = append(m.Tasks, newTask)
			m.ApplySort()
			m.Cursor = len(m.Tasks) - 1
			m.ValidateCursor()
		}
		m.State = StateBrowse
		m.TitleInput.Blur()
		m.NotesInput.Blur()

	case TaskUpdatedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
		} else if msg.Task != nil {
			// Update the task in our list
			for i := range m.Tasks {
				if m.Tasks[i].ID == msg.Task.ID {
					m.Tasks[i].Task = *msg.Task
					break
				}
			}
			m.ApplySort()
		}
		if m.State == StateEditing || m.State == StateEditingNotes {
			m.State = StateBrowse
			m.TitleInput.Blur()
			m.NotesInput.Blur()
		}

	case TaskDeletedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
		} else {
			// Remove the task from our list
			if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
				m.Tasks = append(m.Tasks[:m.SelectedTaskIdx], m.Tasks[m.SelectedTaskIdx+1:]...)
			}
		}
		m.State = StateBrowse
		m.ValidateCursor()

	case LoginMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
		} else {
			m.ErrorMsg = ""
			m.SuccessMsg = "Login successful"
			m.State = StateBrowse
			// Load user data
			cmds = append(cmds, m.loadTasks(), m.loadCategories())
		}

	case RegisterMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMsg = msg.Err.Error()
		} else {
			m.ErrorMsg = ""
			m.SuccessMsg = "Registration successful"
			m.State = StateBrowse
			// Load user data
			cmds = append(cmds, m.loadTasks(), m.loadCategories())
		}

	case TickMsg:
		needsTick := false
		for i := range m.Tasks {
			t := &m.Tasks[i]

			if t.IsDeleting {
				if time.Since(t.AnimStart) > DeleteAnimDuration {
					// Delete from API
					m.SelectedTaskIdx = i
					cmds = append(cmds, m.deleteTask(t.ID))
				} else {
					needsTick = true
				}
			}

			if t.IsAnimatingCheck {
				if time.Since(t.AnimStart) > CheckAnimDuration {
					t.IsAnimatingCheck = false
				} else {
					needsTick = true
				}
			}
		}
		if needsTick {
			cmds = append(cmds, TickCmd())
		}

	case tea.KeyMsg:
		// Clear status messages on key press
		m.ErrorMsg = ""
		m.SuccessMsg = ""

		// Handle different states
		switch m.State {
		case StateLogin, StateRegister:
			return m.updateAuth(msg)
		case StateEditing, StateCreating, StateEditingNotes, StateCreatingNotes:
			return m.updateEditing(msg)
		case StateCategorySelect:
			return m.updateCategorySelect(msg)
		case StateCategoryCreate:
			return m.updateCategoryCreate(msg)
		case StateViewTask:
			return m.updateTaskDetail(msg)
		case StateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case StateHelp:
			return m.updateHelp(msg)
		default:
			return m.updateBrowse(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateAuth handles input in login/register state
func (m Model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.FocusedField == FieldEmail {
			m.FocusedField = FieldPassword
			m.EmailInput.Blur()
			m.PasswordInput.Focus()
		} else {
			m.FocusedField = FieldEmail
			m.PasswordInput.Blur()
			m.EmailInput.Focus()
		}
		return m, textinput.Blink

	case "ctrl+r":
		// Toggle between login and register
		if m.State == StateLogin {
			m.State = StateRegister
		} else {
			m.State = StateLogin
		}
		m.ErrorMsg = ""
		return m, nil

	case "enter":
		email := m.EmailInput.Value()
		password := m.PasswordInput.Value()

		if email == "" || password == "" {
			m.ErrorMsg = "Email and password are required"
			return m, nil
		}

		m.Loading = true
		if m.State == StateLogin {
			return m, m.login(email, password)
		} else {
			return m, m.register(email, password)
		}
	}

	// Update focused input
	if m.FocusedField == FieldEmail {
		m.EmailInput, cmd = m.EmailInput.Update(msg)
	} else {
		m.PasswordInput, cmd = m.PasswordInput.Update(msg)
	}

	return m, cmd
}

// updateEditing handles input during task editing/creation
func (m Model) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.State = StateBrowse
		m.TitleInput.Blur()
		m.NotesInput.Blur()
		m.TempTitle = ""
		m.TempNotes = ""
		m.ValidateCursor()
		return m, nil

	case "enter":
		if m.State == StateCreating {
			val := m.TitleInput.Value()
			if val == "" {
				m.State = StateBrowse
				m.TitleInput.Blur()
				return m, nil
			}
			// Store title and move to notes
			m.TempTitle = val
			m.State = StateCreatingNotes
			m.NotesInput.SetValue("")
			m.NotesInput.Focus()
			m.TitleInput.Blur()
			return m, textinput.Blink
		}

		if m.State == StateCreatingNotes {
			// Create the task
			notes := m.NotesInput.Value()
			m.Loading = true
			return m, m.createTask(m.TempTitle, notes)
		}

		if m.State == StateEditing {
			val := m.TitleInput.Value()
			if val == "" {
				m.State = StateBrowse
				m.TitleInput.Blur()
				return m, nil
			}
			if m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
				m.Loading = true
				return m, m.updateTaskTitle(m.Tasks[m.Cursor].ID, val)
			}
		}

		if m.State == StateEditingNotes {
			val := m.NotesInput.Value()
			if m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
				m.Loading = true
				return m, m.updateTaskNotes(m.Tasks[m.Cursor].ID, val)
			}
		}
	}

	// Update the active input
	if m.State == StateCreating || m.State == StateEditing {
		m.TitleInput, cmd = m.TitleInput.Update(msg)
	} else {
		m.NotesInput, cmd = m.NotesInput.Update(msg)
	}

	return m, cmd
}

// updateBrowse handles input in browse state
func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			m.EnsureCursorVisible()
		}

	case "down", "j":
		if m.Cursor < len(m.Tasks)-1 {
			m.Cursor++
			m.EnsureCursorVisible()
		}

	case "left", "h":
		// Previous page
		if m.Page > 0 {
			m.Page--
			m.Cursor = m.PageStart()
		}

	case "right", "l":
		// Next page
		if m.Page < m.TotalPages()-1 {
			m.Page++
			m.Cursor = m.PageStart()
		}

	case "pgup":
		// Jump to previous page
		if m.Page > 0 {
			m.Page--
			m.Cursor = m.PageStart()
		}

	case "pgdown":
		// Jump to next page
		if m.Page < m.TotalPages()-1 {
			m.Page++
			m.Cursor = m.PageStart()
		}

	case "tab":
		// Cycle view modes
		m.ViewMode = (m.ViewMode + 1) % 3
		m.Cursor = 0
		m.Page = 0
		m.Loading = true
		cmds = append(cmds, m.loadTasks())

	case "t":
		// Cycle themes
		m.ThemeIndex = (m.ThemeIndex + 1) % len(themes.All)
		styles.Update(themes.All[m.ThemeIndex])

	case "s":
		// Cycle sort modes
		m.SortMode = (m.SortMode + 1) % 4
		m.ApplySort()

	case "n":
		// New task - appears at top of first page
		m.State = StateCreating
		m.TitleInput.SetValue("")
		m.TitleInput.Focus()
		m.Cursor = 0
		m.Page = 0
		return m, textinput.Blink

	case "e":
		// Edit task title
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.State = StateEditing
			m.EditingTaskID = m.Tasks[m.Cursor].ID
			m.TitleInput.SetValue(m.Tasks[m.Cursor].Title)
			m.TitleInput.Focus()
			m.TitleInput.SetCursor(len(m.TitleInput.Value()))
			return m, textinput.Blink
		}

	case "E":
		// Edit task notes
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.State = StateEditingNotes
			m.EditingTaskID = m.Tasks[m.Cursor].ID
			notes := ""
			if m.Tasks[m.Cursor].Notes != nil {
				notes = *m.Tasks[m.Cursor].Notes
			}
			m.NotesInput.SetValue(notes)
			m.NotesInput.Focus()
			m.NotesInput.SetCursor(len(m.NotesInput.Value()))
			return m, textinput.Blink
		}

	case "v":
		// Expand/collapse task
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.Tasks[m.Cursor].Expanded = !m.Tasks[m.Cursor].Expanded
		}

	case "enter":
		// Open task detail view
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.SelectedTaskIdx = m.Cursor
			m.State = StateViewTask
		}

	case " ":
		// Toggle task done/open
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			t := &m.Tasks[m.Cursor]
			newStatus := "done"
			if t.Status == "done" {
				newStatus = "open"
			}

			// Start animation
			if newStatus == "done" {
				t.IsAnimatingCheck = true
				t.AnimStart = time.Now()
				t.AnimType = RandomAnimType(m.LastAnim)
				m.LastAnim = t.AnimType
				cmds = append(cmds, TickCmd())
			}

			// Update via API
			cmds = append(cmds, m.updateTaskStatus(t.ID, newStatus))
		}

	case "d":
		// Delete task (with confirmation)
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.SelectedTaskIdx = m.Cursor
			m.State = StateConfirmDelete
		}

	case "c":
		// Open category picker
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			m.SelectedTaskIdx = m.Cursor
			m.CategoryCursor = -1 // Start at "None"
			// If task has a category, select it
			if m.Tasks[m.Cursor].CategoryID != nil {
				for i, cat := range m.Categories {
					if cat.ID == *m.Tasks[m.Cursor].CategoryID {
						m.CategoryCursor = i
						break
					}
				}
			}
			m.State = StateCategorySelect
		}

	case "C":
		// Create new category
		m.State = StateCategoryCreate
		m.CategoryInput.SetValue("")
		m.CategoryInput.Focus()
		return m, textinput.Blink

	case "b":
		// Trigger AI breakdown
		if len(m.Tasks) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
			cmds = append(cmds, m.breakdownTask(m.Tasks[m.Cursor].ID))
		}

	case "?":
		// Show help
		m.PreviousState = m.State
		m.State = StateHelp

	case "L":
		// Logout
		m.Client.Logout()
		m.State = StateLogin
		m.Tasks = nil
		m.Categories = nil
		m.User = nil
		m.EmailInput.SetValue("")
		m.PasswordInput.SetValue("")
		m.EmailInput.Focus()
		return m, textinput.Blink

	case "r", "R":
		// Refresh tasks
		m.Loading = true
		cmds = append(cmds, m.loadTasks(), m.loadCategories())
	}

	return m, tea.Batch(cmds...)
}

// updateCategorySelect handles input in category selection
func (m Model) updateCategorySelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.State = StateBrowse
		return m, nil

	case "up", "k":
		if m.CategoryCursor > -1 {
			m.CategoryCursor--
		}

	case "down", "j":
		if m.CategoryCursor < len(m.Categories)-1 {
			m.CategoryCursor++
		}

	case "enter":
		// Apply selected category
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			var categoryID *int
			if m.CategoryCursor >= 0 && m.CategoryCursor < len(m.Categories) {
				id := m.Categories[m.CategoryCursor].ID
				categoryID = &id
			}
			m.State = StateBrowse
			return m, m.updateTaskCategory(m.Tasks[m.SelectedTaskIdx].ID, categoryID)
		}

	case "C":
		// Create new category
		m.State = StateCategoryCreate
		m.CategoryInput.SetValue("")
		m.CategoryInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// updateCategoryCreate handles input during category creation
func (m Model) updateCategoryCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.State = StateBrowse
		m.CategoryInput.Blur()
		return m, nil

	case "enter":
		name := m.CategoryInput.Value()
		if name == "" {
			m.ErrorMsg = "Category name is required"
			return m, nil
		}
		// Create category with a random color
		colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#DDA0DD", "#98D8C8", "#F7DC6F"}
		color := colors[rand.Intn(len(colors))]
		m.CategoryInput.Blur()
		return m, m.createCategory(name, color)
	}

	m.CategoryInput, cmd = m.CategoryInput.Update(msg)
	return m, cmd
}

// updateTaskDetail handles input in task detail view
func (m Model) updateTaskDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.State = StateBrowse
		return m, nil

	case "e":
		// Edit title
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			m.State = StateEditing
			m.Cursor = m.SelectedTaskIdx
			m.EditingTaskID = m.Tasks[m.SelectedTaskIdx].ID
			m.TitleInput.SetValue(m.Tasks[m.SelectedTaskIdx].Title)
			m.TitleInput.Focus()
			return m, textinput.Blink
		}

	case " ":
		// Toggle done
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			t := &m.Tasks[m.SelectedTaskIdx]
			newStatus := "done"
			if t.Status == "done" {
				newStatus = "open"
			}
			return m, m.updateTaskStatus(t.ID, newStatus)
		}

	case "c":
		// Change category
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			m.CategoryCursor = -1
			if m.Tasks[m.SelectedTaskIdx].CategoryID != nil {
				for i, cat := range m.Categories {
					if cat.ID == *m.Tasks[m.SelectedTaskIdx].CategoryID {
						m.CategoryCursor = i
						break
					}
				}
			}
			m.State = StateCategorySelect
		}

	case "b":
		// Breakdown
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			return m, m.breakdownTask(m.Tasks[m.SelectedTaskIdx].ID)
		}
	}

	return m, nil
}

// updateConfirmDelete handles input in delete confirmation
func (m Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm delete - start animation
		if m.SelectedTaskIdx >= 0 && m.SelectedTaskIdx < len(m.Tasks) {
			m.Tasks[m.SelectedTaskIdx].IsDeleting = true
			m.Tasks[m.SelectedTaskIdx].AnimStart = time.Now()
			m.State = StateBrowse
			return m, TickCmd()
		}

	case "n", "N", "esc":
		m.State = StateBrowse
	}

	return m, nil
}

// updateHelp handles input in help view
func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q":
		m.State = m.PreviousState
		if m.State == StateHelp {
			m.State = StateBrowse
		}
	}
	return m, nil
}

// API command helpers

func (m Model) login(email, password string) tea.Cmd {
	return func() tea.Msg {
		err := m.Client.Login(email, password)
		return LoginMsg{Err: err}
	}
}

func (m Model) register(email, password string) tea.Cmd {
	return func() tea.Msg {
		err := m.Client.Register(email, password)
		return RegisterMsg{Err: err}
	}
}

func (m Model) createTask(title, notes string) tea.Cmd {
	return func() tea.Msg {
		req := api.TaskCreateRequest{Title: title}
		if notes != "" {
			req.Notes = &notes
		}
		task, err := m.Client.CreateTask(req)
		return TaskCreatedMsg{Task: task, Err: err}
	}
}

func (m Model) updateTaskTitle(id int, title string) tea.Cmd {
	return func() tea.Msg {
		req := api.TaskUpdateRequest{Title: &title}
		task, err := m.Client.UpdateTask(id, req)
		return TaskUpdatedMsg{Task: task, Err: err}
	}
}

func (m Model) updateTaskNotes(id int, notes string) tea.Cmd {
	return func() tea.Msg {
		req := api.TaskUpdateRequest{Notes: &notes}
		task, err := m.Client.UpdateTask(id, req)
		return TaskUpdatedMsg{Task: task, Err: err}
	}
}

func (m Model) updateTaskStatus(id int, status string) tea.Cmd {
	return func() tea.Msg {
		req := api.TaskUpdateRequest{Status: &status}
		task, err := m.Client.UpdateTask(id, req)
		return TaskUpdatedMsg{Task: task, Err: err}
	}
}

func (m Model) updateTaskCategory(id int, categoryID *int) tea.Cmd {
	return func() tea.Msg {
		req := api.TaskUpdateRequest{CategoryID: categoryID}
		task, err := m.Client.UpdateTask(id, req)
		return TaskUpdatedMsg{Task: task, Err: err}
	}
}

func (m Model) deleteTask(id int) tea.Cmd {
	return func() tea.Msg {
		err := m.Client.DeleteTask(id)
		return TaskDeletedMsg{Err: err}
	}
}

func (m Model) breakdownTask(id int) tea.Cmd {
	return func() tea.Msg {
		err := m.Client.BreakdownTask(id)
		if err != nil {
			return TaskUpdatedMsg{Err: err}
		}
		// Refetch the task to get new subtasks
		task, err := m.Client.GetTask(id)
		return TaskUpdatedMsg{Task: task, Err: err}
	}
}

func (m Model) createCategory(name, color string) tea.Cmd {
	return func() tea.Msg {
		cat, err := m.Client.CreateCategory(name, color)
		if err != nil {
			return CategoriesLoadedMsg{Err: err}
		}
		// Refetch all categories
		categories, err := m.Client.ListCategories()
		if err == nil && cat != nil {
			// If we were selecting a category, auto-select the new one
			return CategoriesLoadedMsg{Categories: categories}
		}
		return CategoriesLoadedMsg{Categories: categories, Err: err}
	}
}

// ApplySort sorts the tasks based on current sort mode
func (m *Model) ApplySort() {
	switch m.SortMode {
	case SortPriority:
		sort.Slice(m.Tasks, func(i, j int) bool {
			return m.Tasks[i].Priority > m.Tasks[j].Priority
		})
	case SortDueDate:
		sort.Slice(m.Tasks, func(i, j int) bool {
			if m.Tasks[i].DueAt == nil && m.Tasks[j].DueAt == nil {
				return false
			}
			if m.Tasks[i].DueAt == nil {
				return false
			}
			if m.Tasks[j].DueAt == nil {
				return true
			}
			return m.Tasks[i].DueAt.Before(*m.Tasks[j].DueAt)
		})
	case SortAlphabetical:
		sort.Slice(m.Tasks, func(i, j int) bool {
			return m.Tasks[i].Title < m.Tasks[j].Title
		})
	case SortCreated:
		// Newest first (descending by ID)
		sort.Slice(m.Tasks, func(i, j int) bool {
			return m.Tasks[i].ID > m.Tasks[j].ID
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
