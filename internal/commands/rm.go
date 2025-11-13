package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	var tasksToDelete []struct {
		id   int
		name string
	}

	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil {
			fmt.Printf("Warning: Task #%d not found\n", id)
			continue
		}
		tasksToDelete = append(tasksToDelete, struct {
			id   int
			name string
		}{id: id, name: task.Name})
	}

	if len(tasksToDelete) == 0 {
		return fmt.Errorf("no valid tasks to delete")
	}

	// Display tasks to be deleted
	fmt.Printf("\nFound %d task(s) to delete:\n", len(tasksToDelete))
	for _, t := range tasksToDelete {
		fmt.Printf("#%-4d %s\n", t.id, t.name)
	}
	fmt.Println()

	// Ask for confirmation
	fmt.Printf("Delete these %d task(s)? (y/n): ", len(tasksToDelete))
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete tasks
	deleted := 0
	for _, t := range tasksToDelete {
		if err := db.DeleteTask(database, t.id); err != nil {
			fmt.Printf("Warning: Failed to delete task #%d: %v\n", t.id, err)
			continue
		}

		fmt.Printf("✗ Deleted #%d: %s\n", t.id, t.name)
		deleted++
	}

	if deleted > 0 {
		fmt.Printf("\n%d task(s) deleted\n", deleted)
	}

	return nil
}
