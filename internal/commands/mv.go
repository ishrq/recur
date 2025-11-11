package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/parser"
)

func Move(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("mv")
		return nil
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Check if modification string provided
	if len(args) < 2 {
		return fmt.Errorf("modification string required (or use $EDITOR - coming soon)")
	}

	modifyStr := strings.Join(args[1:], " ")

	// Get existing task
	task, err := db.GetTaskByID(database, id)
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	parsedChanges, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return fmt.Errorf("failed to parse changes: %w", err)
	}

	// Store old values for display
	oldName := task.Name

	// Merge changes (new values override old ones)
	task.Name = parsedChanges.Name
	if parsedChanges.DueDate != nil {
		task.DueDate = parsedChanges.DueDate
	}
	if parsedChanges.Tag != "" {
		task.Tag = parsedChanges.Tag
	}
	if parsedChanges.Project != "" {
		task.Project = parsedChanges.Project
	}
	if parsedChanges.Priority != "" {
		task.Priority = parsedChanges.Priority
	}
	if parsedChanges.Note != "" {
		task.Note = parsedChanges.Note
	}

	if err := db.UpdateTask(database, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✓ Updated #%d\n", id)
	fmt.Printf("  Old: %s\n", oldName)
	fmt.Printf("  New: %s\n", task.Name)

	return nil
}
