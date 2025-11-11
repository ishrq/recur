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
		printDoneHelp()
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

func printDoneHelp() {
	help := `Mark tasks as complete

	Usage:
	recur done <id>              Complete a task
	recur done <id1> <id2> ...   Complete multiple tasks

	Options:
	-h, --help                   Show this help message

	More options coming soon:
	--tag, --project, --priority
	--due, --query
	--undo

	Examples:
	recur done 1
	recur done 1 2 3
	recur done 5`

	fmt.Println(help)
}
