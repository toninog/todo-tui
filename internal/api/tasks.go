package api

import (
	"fmt"
	"net/url"
	"time"
)

// Task represents a task from the API
type Task struct {
	ID                 int        `json:"id"`
	Title              string     `json:"title"`
	Notes              *string    `json:"notes"`
	Status             string     `json:"status"`
	DueAt              *time.Time `json:"due_at"`
	Priority           int        `json:"priority"`
	EffortMin          int        `json:"effort_min"`
	NotificationsEnabled bool     `json:"notifications_enabled"`
	ScheduledStart     *time.Time `json:"scheduled_start"`
	ScheduledEnd       *time.Time `json:"scheduled_end"`
	AutoScheduled      bool       `json:"auto_scheduled"`
	PreferredTimeOfDay *string    `json:"preferred_time_of_day"`
	CategoryID         *int       `json:"category_id"`
	Category           *Category  `json:"category"`
	Subtasks           []Subtask  `json:"subtasks"`
	Tags               []string   `json:"tags"`
	IsOwner            *bool      `json:"is_owner"`
	CanComplete        *bool      `json:"can_complete"`
	CanDelete          *bool      `json:"can_delete"`
	CanShare           *bool      `json:"can_share"`
	SharedWith         []ShareInfo `json:"shared_with"`
	OwnerEmail         *string    `json:"owner_email"`
	CommentCount       int        `json:"comment_count"`
}

// Subtask represents a subtask
type Subtask struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Sort   int    `json:"sort"`
}

// ShareInfo represents sharing information
type ShareInfo struct {
	UserID  *int   `json:"user_id"`
	Email   string `json:"email"`
	Pending bool   `json:"pending"`
}

// TaskCreateRequest represents a create task request
type TaskCreateRequest struct {
	Title             string     `json:"title"`
	Notes             *string    `json:"notes,omitempty"`
	DueAt             *time.Time `json:"due_at,omitempty"`
	Priority          *int       `json:"priority,omitempty"`
	EffortMin         *int       `json:"effort_min,omitempty"`
	CategoryID        *int       `json:"category_id,omitempty"`
	GenerateSubtasks  *bool      `json:"generate_subtasks,omitempty"`
}

// TaskUpdateRequest represents an update task request
type TaskUpdateRequest struct {
	Title              *string    `json:"title,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
	Status             *string    `json:"status,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	Priority           *int       `json:"priority,omitempty"`
	EffortMin          *int       `json:"effort_min,omitempty"`
	CategoryID         *int       `json:"category_id,omitempty"`
	NotificationsEnabled *bool    `json:"notifications_enabled,omitempty"`
}

// SubtaskCreateRequest represents a create subtask request
type SubtaskCreateRequest struct {
	Title string `json:"title"`
}

// SubtaskUpdateRequest represents an update subtask request
type SubtaskUpdateRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty"`
	Sort   *int    `json:"sort,omitempty"`
}

// TaskListParams holds parameters for listing tasks
type TaskListParams struct {
	Status     string // "open" or "done"
	Scope      string // "all", "mine", or "shared"
	CategoryID *int
	Search     string
}

// ListTasks fetches tasks with optional filters
func (c *Client) ListTasks(params TaskListParams) ([]Task, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Scope != "" {
		query.Set("scope", params.Scope)
	}
	if params.CategoryID != nil {
		query.Set("category_id", fmt.Sprintf("%d", *params.CategoryID))
	}
	if params.Search != "" {
		query.Set("search", params.Search)
	}

	path := "/tasks"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var tasks []Task
	if err := c.Get(path, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask fetches a single task by ID
func (c *Client) GetTask(id int) (*Task, error) {
	var task Task
	if err := c.Get(fmt.Sprintf("/tasks/%d", id), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask creates a new task
func (c *Client) CreateTask(req TaskCreateRequest) (*Task, error) {
	var task Task
	if err := c.Post("/tasks", req, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(id int, req TaskUpdateRequest) (*Task, error) {
	var task Task
	if err := c.Patch(fmt.Sprintf("/tasks/%d", id), req, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(id int) error {
	return c.Delete(fmt.Sprintf("/tasks/%d", id), nil)
}

// CompleteTask marks a task as complete
func (c *Client) CompleteTask(id int) error {
	return c.Post(fmt.Sprintf("/tasks/%d/complete", id), nil, nil)
}

// BreakdownTask triggers AI breakdown of a task
func (c *Client) BreakdownTask(id int) error {
	return c.Post(fmt.Sprintf("/tasks/%d/breakdown", id), nil, nil)
}

// CreateSubtask creates a new subtask
func (c *Client) CreateSubtask(taskID int, title string) error {
	req := SubtaskCreateRequest{Title: title}
	return c.Post(fmt.Sprintf("/tasks/%d/subtasks", taskID), req, nil)
}

// UpdateSubtask updates an existing subtask
func (c *Client) UpdateSubtask(subtaskID int, req SubtaskUpdateRequest) error {
	return c.Patch(fmt.Sprintf("/tasks/subtasks/%d", subtaskID), req, nil)
}
