package commands

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ishrq/recur/internal/db"
)

func Remove(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("rm")
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

	deleted := 0
	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil {
			fmt.Printf("Warning: Task #%d not found\n", id)
			continue
		}

		if err := db.DeleteTask(database, id); err != nil {
			fmt.Printf("Warning: Failed to delete task #%d: %v\n", id, err)
			continue
		}

		fmt.Printf("✗ Deleted #%d: %s\n", id, task.Name)
		deleted++
	}

	if deleted > 0 {
		fmt.Printf("\n%d task(s) deleted\n", deleted)
	}

	return nil
}
