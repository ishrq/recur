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
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("rm")
		return nil
	}

	var ids []int
	var tags []string
	var projects []string
	var priorities []string
	var removeAll bool
	var removeDone bool
	var removeTrash bool
	var purge bool

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
		case "--tag", "-t":
			// Collect all following non-flag arguments as tags
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				tags = append(tags, args[i+1])
				i++
			}
		case "--project", "-p":
			// Collect all following non-flag arguments as projects
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				projects = append(projects, args[i+1])
				i++
			}
		case "--priority", "-P":
			// Collect all following non-flag arguments as priorities
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				priorities = append(priorities, args[i+1])
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

	// Collect valid tasks to delete
	var tasksToDelete []struct {
		id   int
		name string
	}
	permanentDelete := false

	if removeTrash {
		permanentDelete = true
		deletedTasks, err := db.GetDeletedTasks(database)
		if err != nil {
			return fmt.Errorf("failed to get deleted tasks: %w", err)
		}

		for _, task := range deletedTasks {
			tasksToDelete = append(tasksToDelete, struct {
				id   int
				name string
			}{id: task.ID, name: task.Name})
		}
	} else if removeAll {
		allTasks, err := db.GetTasks(database, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		for _, task := range allTasks {
			tasksToDelete = append(tasksToDelete, struct {
				id   int
				name string
			}{id: task.ID, name: task.Name})
		}
	} else if removeDone {
		allTasks, err := db.GetTasks(database, true)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		for _, task := range allTasks {
			if task.CompletedDate != nil {
				tasksToDelete = append(tasksToDelete, struct {
					id   int
					name string
				}{id: task.ID, name: task.Name})
			}
		}
	} else if len(tags) > 0 || len(projects) > 0 || len(priorities) > 0 {
		// Handle filters
		allTasks, err := db.GetTasks(database, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Apply filters
		for _, task := range allTasks {
			matched := false

			if len(tags) > 0 {
				for _, tag := range tags {
					if strings.EqualFold(task.Tag, tag) {
						matched = true
						break
					}
				}
			}

			if len(projects) > 0 && !matched {
				for _, project := range projects {
					if strings.EqualFold(task.Project, project) {
						matched = true
						break
					}
				}
			}

			if len(priorities) > 0 && !matched {
				for _, priority := range priorities {
					if strings.EqualFold(task.Priority, priority) {
						matched = true
						break
					}
				}
			}

			if matched {
				tasksToDelete = append(tasksToDelete, struct {
					id   int
					name string
				}{id: task.ID, name: task.Name})
			}
		}
	} else {
		// Handle IDs
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
		fmt.Printf("#%-4d %s\n", t.id, t.name)
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
			err = db.PermanentlyDeleteTask(database, t.id)
		} else {
			err = db.DeleteTask(database, t.id)
		}

		if err != nil {
			fmt.Printf("Warning: Failed to delete task #%d: %v\n", t.id, err)
			continue
		}

		fmt.Printf("✗ Deleted #%d: %s\n", t.id, t.name)
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
