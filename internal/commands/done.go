package commands

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ishrq/recur/internal/db"
)

func Done(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("done")
		return nil
	}

	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", arg)
		}
		ids = append(ids, id)
	}

	completed := 0
	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil {
			fmt.Printf("Warning: Task #%d not found\n", id)
			continue
		}

		if task.CompletedDate != nil {
			fmt.Printf("Task #%d already completed\n", id)
			continue
		}

		if err := db.MarkTaskDone(database, id); err != nil {
			fmt.Printf("Warning: Failed to complete task #%d: %v\n", id, err)
			continue
		}

		fmt.Printf("✓ Completed #%d: %s\n", id, task.Name)
		completed++
	}

	if completed > 0 {
		fmt.Printf("\n%d task(s) completed\n", completed)
	}

	return nil
}
