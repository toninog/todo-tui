package models

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/blackraven/todo-tui/internal/api"
	"github.com/blackraven/todo-tui/internal/themes"
)

// AppState represents the current state of the application
type AppState int

const (
	StateLogin AppState = iota
	StateRegister
	StateBrowse
	StateEditing
	StateCreating
	StateEditingNotes
	StateCreatingNotes
	StateViewTask
	StateCategorySelect
	StateCategoryCreate
	StateHelp
	StateConfirmDelete
)

// ViewMode represents which list view is active
type ViewMode int

const (
	ViewOpen ViewMode = iota
	ViewCompleted
	ViewShared
)

// SortMode represents the current sort order
type SortMode int

const (
	SortPriority SortMode = iota
	SortDueDate
	SortAlphabetical
	SortCreated
)

// InputField represents which input field is focused
type InputField int

const (
	FieldEmail InputField = iota
	FieldPassword
)

// Animation type constants
const (
	AnimSparkle = iota
	AnimMatrix
	AnimWipeRight
	AnimWipeLeft
	AnimRainbow
	AnimWave
	AnimBinary
	AnimDissolve
	AnimFlip
	AnimPulse
	AnimTypewriter
	AnimParticle
	AnimRedact
	AnimChaos
	AnimConverge
	AnimBounce
	AnimSpin
	AnimZipper
	AnimEraser
	AnimGlitch
	AnimMoons
	AnimBraille
	AnimHex
	AnimReverse
	AnimCaseFlip
	AnimWide
	AnimTraffic
	AnimCenterStrike
	AnimLoading
	AnimSlider

	AnimCount = 30
)

// Animation duration constants
const (
	CheckAnimDuration  = 290 * time.Millisecond
	DeleteAnimDuration = 200 * time.Millisecond
	FPS                = 60
)

// Task wraps the API task with UI state
type Task struct {
	api.Task

	// UI States
	Expanded bool

	// Animation States
	IsAnimatingCheck bool
	IsDeleting       bool
	AnimType         int
	AnimStart        time.Time
}

// Category wraps the API category
type Category = api.Category

// TickMsg is sent on animation tick
type TickMsg struct{}

// TasksLoadedMsg is sent when tasks are loaded from API
type TasksLoadedMsg struct {
	Tasks []api.Task
	Err   error
}

// CategoriesLoadedMsg is sent when categories are loaded from API
type CategoriesLoadedMsg struct {
	Categories []api.Category
	Err        error
}

// TaskCreatedMsg is sent when a task is created
type TaskCreatedMsg struct {
	Task *api.Task
	Err  error
}

// TaskUpdatedMsg is sent when a task is updated
type TaskUpdatedMsg struct {
	Task *api.Task
	Err  error
}

// TaskDeletedMsg is sent when a task is deleted
type TaskDeletedMsg struct {
	Err error
}

// LoginMsg is sent after login attempt
type LoginMsg struct {
	Err error
}

// RegisterMsg is sent after register attempt
type RegisterMsg struct {
	Err error
}

// Model is the main application model
type Model struct {
	// API client
	Client *api.Client

	// Application state
	State         AppState
	PreviousState AppState
	ViewMode      ViewMode
	SortMode      SortMode
	ThemeIndex    int
	LastAnim      int

	// Data
	Tasks      []Task
	Categories []Category
	User       *api.UserInfo

	// UI state
	Cursor         int
	CategoryCursor int
	Width          int
	Height         int

	// Pagination
	Page     int
	PageSize int

	// Input fields
	EmailInput    textinput.Model
	PasswordInput textinput.Model
	TitleInput    textinput.Model
	NotesInput    textinput.Model
	CategoryInput textinput.Model
	FocusedField  InputField

	// Temporary storage
	TempTitle       string
	TempNotes       string
	EditingTaskID   int
	SelectedTaskIdx int

	// Status messages
	ErrorMsg   string
	SuccessMsg string

	// Loading state
	Loading bool
}

// NewModel creates a new application model
func NewModel(client *api.Client) Model {
	emailInput := textinput.New()
	emailInput.Placeholder = "email@example.com"
	emailInput.CharLimit = 100
	emailInput.Width = 40

	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 100
	passwordInput.Width = 40

	titleInput := textinput.New()
	titleInput.Placeholder = "Task title..."
	titleInput.CharLimit = 200
	titleInput.Width = 60

	notesInput := textinput.New()
	notesInput.Placeholder = "Notes..."
	notesInput.CharLimit = 1000
	notesInput.Width = 60

	categoryInput := textinput.New()
	categoryInput.Placeholder = "Category name..."
	categoryInput.CharLimit = 50
	categoryInput.Width = 40

	// Determine initial state based on token
	initialState := StateLogin
	if client.HasToken() && client.ValidateToken() {
		initialState = StateBrowse
	} else if client.HasCredentials() {
		// Try auto-login with stored credentials
		if err := client.AutoLogin(); err == nil {
			initialState = StateBrowse
		}
	}

	m := Model{
		Client:        client,
		State:         initialState,
		ViewMode:      ViewOpen,
		SortMode:      SortCreated,
		ThemeIndex:    0,
		PageSize:      10,
		Page:          0,
		EmailInput:    emailInput,
		PasswordInput: passwordInput,
		TitleInput:    titleInput,
		NotesInput:    notesInput,
		CategoryInput: categoryInput,
		FocusedField:  FieldEmail,
	}

	if initialState == StateLogin {
		m.EmailInput.Focus()
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.State == StateBrowse {
		return tea.Batch(
			m.loadTasks(),
			m.loadCategories(),
		)
	}
	return textinput.Blink
}

// loadTasks creates a command to load tasks from the API
func (m Model) loadTasks() tea.Cmd {
	return func() tea.Msg {
		params := api.TaskListParams{
			Status: "open",
			Scope:  "all",
		}
		switch m.ViewMode {
		case ViewCompleted:
			params.Status = "done"
		case ViewShared:
			params.Scope = "shared"
			params.Status = ""
		}

		tasks, err := m.Client.ListTasks(params)
		return TasksLoadedMsg{Tasks: tasks, Err: err}
	}
}

// loadCategories creates a command to load categories from the API
func (m Model) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.Client.ListCategories()
		return CategoriesLoadedMsg{Categories: categories, Err: err}
	}
}

// CurrentTheme returns the current theme
func (m Model) CurrentTheme() themes.Theme {
	if m.ThemeIndex >= 0 && m.ThemeIndex < len(themes.All) {
		return themes.All[m.ThemeIndex]
	}
	return themes.All[0]
}

// ValidateCursor ensures the cursor is within valid bounds
func (m *Model) ValidateCursor() {
	switch m.State {
	case StateCategorySelect:
		if len(m.Categories) == 0 {
			m.CategoryCursor = 0
		} else if m.CategoryCursor >= len(m.Categories) {
			m.CategoryCursor = len(m.Categories) - 1
		} else if m.CategoryCursor < 0 {
			m.CategoryCursor = 0
		}
	default:
		if len(m.Tasks) == 0 {
			m.Cursor = 0
		} else if m.Cursor >= len(m.Tasks) {
			m.Cursor = len(m.Tasks) - 1
		} else if m.Cursor < 0 {
			m.Cursor = 0
		}
	}
}

// CurrentTask returns the currently selected task, or nil if none
func (m *Model) CurrentTask() *Task {
	if m.Cursor >= 0 && m.Cursor < len(m.Tasks) {
		return &m.Tasks[m.Cursor]
	}
	return nil
}

// ViewModeString returns a string representation of the current view mode
func (m Model) ViewModeString() string {
	switch m.ViewMode {
	case ViewOpen:
		return "Open"
	case ViewCompleted:
		return "Completed"
	case ViewShared:
		return "Shared"
	}
	return "Unknown"
}

// TickCmd returns a command that sends tick messages for animations
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second/FPS, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// TotalPages returns the total number of pages
func (m Model) TotalPages() int {
	if len(m.Tasks) == 0 {
		return 1
	}
	pages := len(m.Tasks) / m.PageSize
	if len(m.Tasks)%m.PageSize != 0 {
		pages++
	}
	return pages
}

// PageStart returns the starting index for the current page
func (m Model) PageStart() int {
	return m.Page * m.PageSize
}

// PageEnd returns the ending index (exclusive) for the current page
func (m Model) PageEnd() int {
	end := (m.Page + 1) * m.PageSize
	if end > len(m.Tasks) {
		end = len(m.Tasks)
	}
	return end
}

// PageTasks returns the tasks for the current page
func (m Model) PageTasks() []Task {
	if len(m.Tasks) == 0 {
		return nil
	}
	start := m.PageStart()
	end := m.PageEnd()
	if start >= len(m.Tasks) {
		return nil
	}
	return m.Tasks[start:end]
}

// CursorInPage returns the cursor position relative to the current page
func (m Model) CursorInPage() int {
	return m.Cursor - m.PageStart()
}

// ValidatePage ensures the page is within valid bounds
func (m *Model) ValidatePage() {
	totalPages := m.TotalPages()
	if m.Page < 0 {
		m.Page = 0
	}
	if m.Page >= totalPages {
		m.Page = totalPages - 1
	}
	if m.Page < 0 {
		m.Page = 0
	}
}

// EnsureCursorVisible adjusts the page to make the cursor visible
func (m *Model) EnsureCursorVisible() {
	if len(m.Tasks) == 0 {
		m.Page = 0
		return
	}
	// Calculate which page the cursor should be on
	m.Page = m.Cursor / m.PageSize
	m.ValidatePage()
}
