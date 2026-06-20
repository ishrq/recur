package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/filter"
	"github.com/ishrq/recur/internal/models"
)

func Remove(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("rm")
		return nil
	}

	filters, remaining := extractFilterFlags(args)

	var ids []int
	var removeAll bool
	var removeDone bool
	var removeTrash bool
	var purge bool
	var undo bool

	for _, arg := range remaining {
		switch arg {
		case "--all":
			removeAll = true
		case "--done":
			removeDone = true
		case "--trash":
			removeTrash = true
		case "--purge":
			purge = true
		case "--undo":
			undo = true
		default:
			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", arg)
			}
			ids = append(ids, id)
		}
	}

	// Check for conflicting flags
	if undo {
		if removeAll || removeDone || removeTrash || purge {
			return fmt.Errorf("--undo cannot be combined with --all, --done, --trash, or --purge")
		}
	}

	specialFlags := []bool{removeAll, removeDone, removeTrash, purge}
	specialFlagCount := 0
	for _, flag := range specialFlags {
		if flag {
			specialFlagCount++
		}
	}

	if specialFlagCount > 1 {
		return fmt.Errorf("cannot combine --all, --done, --trash, and --purge flags")
	}

	if purge {
		count, err := db.GetAllTasksCount(database)
		if err != nil {
			return fmt.Errorf("failed to count tasks: %w", err)
		}

		if count == 0 {
			fmt.Println("No tasks in database.")
			return nil
		}

		fmt.Println("\n⚠️  WARNING: This will PERMANENTLY delete ALL tasks from the database.")
		fmt.Println("⚠️  This action CANNOT be undone!")
		fmt.Printf("\nTotal tasks in database: %d\n\n", count)
		fmt.Print("Type 'DELETE ALL' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(response)
		if response != "DELETE ALL" {
			fmt.Println("Purge cancelled.")
			return nil
		}

		if err := db.PurgeAllTasks(database); err != nil {
			return fmt.Errorf("failed to purge tasks: %w", err)
		}

		fmt.Printf("\n✓ All tasks permanently deleted\n")
		return nil
	}

	if undo {
		var tasksToRestore []*models.Task
		var initialTasks []models.Task
		var err error

		if len(ids) > 0 {
			// Get tasks by IDs and check if deleted
			for _, id := range ids {
				task, err := db.GetTaskByID(database, id)
				if err != nil {
					fmt.Printf("Warning: Task #%d not found\n", id)
					continue
				}
				// Only include deleted tasks for undo
				if task.Deleted {
					initialTasks = append(initialTasks, *task)
				} else {
					fmt.Printf("Warning: Task #%d is not deleted\n", id)
				}
			}
		} else {
			// Get all deleted tasks for filtering
			initialTasks, err = db.GetTasks(database, true, false)
			if err != nil {
				return fmt.Errorf("failed to get deleted tasks: %w", err)
			}

			// Apply filters
			initialTasks, err = filter.ApplyFilters(initialTasks, filters)
			if err != nil {
				return err
			}
		}

		// Convert to tasksToRestore
		for i := range initialTasks {
			tasksToRestore = append(tasksToRestore, &initialTasks[i])
		}

		if len(tasksToRestore) == 0 {
			return fmt.Errorf("no deleted tasks found matching criteria")
		}

		// Display tasks to be restored
		fmt.Printf("\nFound %d deleted task(s) to restore:\n", len(tasksToRestore))
		for _, t := range tasksToRestore {
			fmt.Printf("#%-4d %s\n", t.ID, t.Name)
		}
		fmt.Println()

		// Ask for confirmation
		fmt.Printf("Restore these %d task(s)? (y/n): ", len(tasksToRestore))
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restore cancelled.")
			return nil
		}

		// Restore tasks
		restored := 0
		for _, t := range tasksToRestore {
			if err := db.RestoreTask(database, t.ID); err != nil {
				fmt.Printf("Warning: Failed to restore task #%d: %v\n", t.ID, err)
				continue
			}

			fmt.Printf("↺ Restored #%d: %s\n", t.ID, t.Name)
			restored++
		}

		if restored > 0 {
			fmt.Printf("\n%d task(s) restored\n", restored)
		}

		return nil
	}

	// Collect initial task set
	var tasksToDelete []*models.Task
	permanentDelete := false
	var initialTasks []models.Task
	var err error

	if removeTrash {
		permanentDelete = true
		initialTasks, err = db.GetTasks(database, true, false)
		if err != nil {
			return fmt.Errorf("failed to get deleted tasks: %w", err)
		}
	} else if removeAll {
		initialTasks, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else if removeDone {
		initialTasks, err = db.GetTasks(database, false, true)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else if len(ids) > 0 {
		// Get tasks by IDs
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil {
				fmt.Printf("Warning: Task #%d not found\n", id)
				continue
			}
			// Skip deleted tasks for normal operations
			if !task.Deleted {
				initialTasks = append(initialTasks, *task)
			} else {
				fmt.Printf("Warning: Task #%d is already deleted\n", id)
			}
		}
	} else {
		// Get all incomplete tasks for filtering
		initialTasks, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	}

	// Apply filters
	initialTasks, err = filter.ApplyFilters(initialTasks, filters)
	if err != nil {
		return err
	}

	// Convert to tasksToDelete
	for i := range initialTasks {
		tasksToDelete = append(tasksToDelete, &initialTasks[i])
	}

	if len(tasksToDelete) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// Display tasks to be deleted
	fmt.Println()
	if permanentDelete {
		fmt.Println("⚠️  WARNING: This will PERMANENTLY delete tasks from the database.")
		fmt.Println("⚠️  This action CANNOT be undone!")
		fmt.Println()
	}
	fmt.Printf("Found %d task(s) to delete:\n", len(tasksToDelete))
	for _, t := range tasksToDelete {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	// Ask for confirmation
	if permanentDelete {
		fmt.Printf("Permanently delete these %d task(s)? (y/n): ", len(tasksToDelete))
	} else {
		fmt.Printf("Delete these %d task(s)? (y/n): ", len(tasksToDelete))
	}

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
		var err error
		if permanentDelete {
			err = db.PermanentlyDeleteTask(database, t.ID)
		} else {
			err = db.DeleteTask(database, t.ID)
		}

		if err != nil {
			fmt.Printf("Warning: Failed to delete task #%d: %v\n", t.ID, err)
			continue
		}

		fmt.Printf("✗ Deleted #%d: %s\n", t.ID, t.Name)
		deleted++
	}

	if deleted > 0 {
		if permanentDelete {
			fmt.Printf("\n%d task(s) permanently deleted\n", deleted)
		} else {
			fmt.Printf("\n%d task(s) deleted\n", deleted)
		}
	}

	return nil
}
