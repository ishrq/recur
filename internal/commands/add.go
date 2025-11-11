package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/parser"
)

func Add(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task name required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("add")
		return nil
	}

	taskString := strings.Join(args, " ")

	// Parse the task string
	task, err := parser.ParseTaskString(taskString)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	// Insert task
	id, err := db.InsertTask(database, task)
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	// Display confirmation
	fmt.Printf("Added task #%d: %s\n", id, task.Name)

	if task.DueDate != nil {
		fmt.Printf("  Due: %s\n", task.DueDate.Format("Mon Jan 2, 2006 15:04"))
	}
	if task.RecurFrequency != "" {
		fmt.Printf("  Recurs: %s\n", task.RecurFrequency)
	}
	if task.RecurEndDate != nil {
		fmt.Printf("  Until: %s\n", task.RecurEndDate.Format("Mon Jan 2, 2006"))
	}
	if task.Project != "" {
		fmt.Printf("  Project: %s\n", task.Project)
	}
	if task.Tag != "" {
		fmt.Printf("  Tag: %s\n", task.Tag)
	}
	if task.Priority != "" {
		fmt.Printf("  Priority: %s\n", task.Priority)
	}
	if task.Note != "" {
		fmt.Printf("  Note: %s\n", task.Note)
	}

	return nil
}
