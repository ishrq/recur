package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

func Add(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task name required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		printAddHelp()
		return nil
	}

	// NOTE: For now, join all args as the task name
	// Add parsing for @() #tag !project etc. later
	taskName := strings.Join(args, " ")

	task := &models.Task{
		Name:        taskName,
		CreatedDate: time.Now(),
	}

	id, err := db.InsertTask(database, task)
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	fmt.Printf("Added task #%d: %s\n", id, taskName)
	return nil
}

func printAddHelp() {
	help := `Add a new task

	Usage:
	recur add "Task name"
	recur add "Task name @(date time, frequency, end) #tag !project !priority *note"

	Examples:
	recur add "Buy groceries"
	recur add "Team meeting @(2025-11-12 15:00) !Work"
	recur add "Water plants @(today 9am, 1d) #chores"

	Syntax (coming soon):
	@(date time, frequency, end)  - Due date and recurrence
	#tag                          - Add a tag
	!project                      - Add to project
	!priority                     - Set priority
	*note                         - Add a note

	For detailed syntax information, visit the documentation.`

	fmt.Println(help)
}
