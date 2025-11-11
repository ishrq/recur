package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/db"
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

	// NOTE: For now, just update the name
	// Full parsing for @() #tag !project etc. will be implemented later
	oldName := task.Name
	task.Name = modifyStr

	if err := db.UpdateTask(database, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✓ Updated #%d\n", id)
	fmt.Printf("  Old: %s\n", oldName)
	fmt.Printf("  New: %s\n", task.Name)

	return nil
}
