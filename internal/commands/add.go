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
		PrintHelp("add")
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
