package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/blackraven/todo-tui/internal/api"
	"github.com/blackraven/todo-tui/internal/config"
	"github.com/blackraven/todo-tui/internal/models"
	"github.com/blackraven/todo-tui/internal/styles"
)

func main() {
	// Parse command line flags
	newTask := flag.String("n", "", "Create a new task with the given title")
	listTasks := flag.Bool("l", false, "List all open tasks")
	deleteTask := flag.Int("d", 0, "Delete a task by ID")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Ensure data directory exists
	if err := config.EnsureDataDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(cfg)

	// Handle CLI modes
	if *newTask != "" {
		createTaskFromCLI(client, *newTask)
		return
	}

	if *listTasks {
		listTasksFromCLI(client)
		return
	}

	if *deleteTask > 0 {
		deleteTaskFromCLI(client, *deleteTask)
		return
	}

	// Initialize styles
	styles.Init()

	// Create initial model
	model := models.NewModel(client)

	// Create Bubble Tea program with alt screen
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// createTaskFromCLI creates a task directly from command line
func createTaskFromCLI(client *api.Client, title string) {
	// Ensure we're authenticated
	if !ensureAuth(client) {
		return
	}

	// Create the task
	req := api.TaskCreateRequest{Title: title}
	task, err := client.CreateTask(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created task #%d: %s\n", task.ID, task.Title)
}

// listTasksFromCLI lists all open tasks
func listTasksFromCLI(client *api.Client) {
	// Ensure we're authenticated
	if !ensureAuth(client) {
		return
	}

	// Fetch tasks
	params := api.TaskListParams{
		Status: "open",
		Scope:  "all",
	}
	tasks, err := client.ListTasks(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No open tasks.")
		return
	}

	fmt.Printf("Open tasks (%d):\n", len(tasks))
	for _, task := range tasks {
		status := "[ ]"
		if task.Status == "done" {
			status = "[x]"
		}
		fmt.Printf("  #%-4d %s %s\n", task.ID, status, task.Title)
	}
}

// deleteTaskFromCLI deletes a task by ID
func deleteTaskFromCLI(client *api.Client, id int) {
	// Ensure we're authenticated
	if !ensureAuth(client) {
		return
	}

	// Delete the task
	err := client.DeleteTask(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted task #%d\n", id)
}

// ensureAuth ensures the client is authenticated
func ensureAuth(client *api.Client) bool {
	if !client.HasToken() || !client.ValidateToken() {
		// Try auto-login
		if client.HasCredentials() {
			if err := client.AutoLogin(); err != nil {
				fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Please run the TUI to login first.\n")
				os.Exit(1)
				return false
			}
		} else {
			fmt.Fprintf(os.Stderr, "Not authenticated. Please run the TUI to login first.\n")
			os.Exit(1)
			return false
		}
	}
	return true
}
