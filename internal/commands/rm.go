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

	var ids []int
	var removeAll bool
	var removeDone bool
	var removeTrash bool
	var purge bool

	filters := filter.Filters{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all":
			removeAll = true
		case "--done":
			removeDone = true
		case "--trash":
			removeTrash = true
		case "--purge":
			purge = true
		case "--today":
			filters.Today = true
		case "--tomorrow":
			filters.Tomorrow = true
		case "--overdue":
			filters.Overdue = true
		case "--upcoming":
			filters.Upcoming = true
		case "--due", "-d":
			if i+1 < len(args) {
				filters.DueDate = args[i+1]
				i++
			}
		case "--from":
			if i+1 < len(args) {
				filters.FromDate = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				filters.ToDate = args[i+1]
				i++
			}
		case "--query", "-q":
			if i+1 < len(args) {
				filters.Query = args[i+1]
				i++
			}
		case "--tag", "-t":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Tags = append(filters.Tags, args[i+1])
				i++
			}
		case "--project", "-p":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Projects = append(filters.Projects, args[i+1])
				i++
			}
		case "--priority", "-P":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Priorities = append(filters.Priorities, args[i+1])
				i++
			}
		default:
			// Try to parse as ID
			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", arg)
			}
			ids = append(ids, id)
		}
	}

	// Check for conflicting flags
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
		fmt.Print("Type 'DELETE' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(response)
		if response != "DELETE" {
			fmt.Println("Purge cancelled.")
			return nil
		}

		if err := db.PurgeAllTasks(database); err != nil {
			return fmt.Errorf("failed to purge tasks: %w", err)
		}

		fmt.Printf("\n✓ All tasks permanently deleted\n")
		return nil
	}

	// Collect initial task set
	var tasksToDelete []*models.Task
	permanentDelete := false
	var initialTasks []models.Task
	var err error

	if removeTrash {
		permanentDelete = true
		initialTasks, err = db.GetDeletedTasks(database)
		if err != nil {
			return fmt.Errorf("failed to get deleted tasks: %w", err)
		}
	} else if removeAll {
		initialTasks, err = db.GetTasks(database, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else if removeDone {
		allTasks, err := db.GetTasks(database, true)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
		for _, task := range allTasks {
			if task.CompletedDate != nil {
				initialTasks = append(initialTasks, task)
			}
		}
	} else if len(ids) > 0 {
		// Get tasks by IDs
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil {
				fmt.Printf("Warning: Task #%d not found\n", id)
				continue
			}
			initialTasks = append(initialTasks, *task)
		}
	} else {
		// Get all incomplete tasks for filtering
		initialTasks, err = db.GetTasks(database, false)
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
