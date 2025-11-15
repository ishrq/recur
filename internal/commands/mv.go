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
	"github.com/ishrq/recur/internal/parser"
)

func Move(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("mv")
		return nil
	}

	// Parse flags and IDs
	var ids []int
	var modifyStr string
	filters := filter.Filters{}
	foundModifyStr := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
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
				// Not an ID, must be the modification string
				modifyStr = strings.Join(args[i:], " ")
				foundModifyStr = true
				break
			}
			ids = append(ids, id)
		}
		if foundModifyStr {
			break
		}
	}

	// Collect tasks to edit
	var tasksToEdit []models.Task
	var err error

	if len(ids) > 0 {
		// Get tasks by IDs
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil {
				fmt.Printf("Warning: Task #%d not found\n", id)
				continue
			}
			if task.Deleted {
				fmt.Printf("Warning: Task #%d is deleted\n", id)
				continue
			}
			tasksToEdit = append(tasksToEdit, *task)
		}
	} else {
		// Get all incomplete tasks for filtering
		tasksToEdit, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Apply filters
		tasksToEdit, err = filter.ApplyFilters(tasksToEdit, filters)
		if err != nil {
			return err
		}
	}

	if len(tasksToEdit) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// If no modification string provided, return error (TODO: implement $EDITOR support)
	if modifyStr == "" {
		return fmt.Errorf("modification string required (or use $EDITOR - coming soon)")
	}

	// Display tasks to be edited
	fmt.Printf("\nFound %d task(s) to edit:\n", len(tasksToEdit))
	for _, t := range tasksToEdit {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	// Ask for confirmation
	fmt.Printf("Edit these %d task(s)? (y/n): ", len(tasksToEdit))
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Parse changes
	parsedChanges, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return fmt.Errorf("failed to parse changes: %w", err)
	}

	// Edit tasks
	updated := 0
	for _, task := range tasksToEdit {
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

		if err := db.UpdateTask(database, &task); err != nil {
			fmt.Printf("Warning: Failed to update task #%d: %v\n", task.ID, err)
			continue
		}

		fmt.Printf("✓ Updated #%d\n", task.ID)
		fmt.Printf("  Old: %s\n", oldName)
		fmt.Printf("  New: %s\n", task.Name)
		updated++
	}

	if updated > 0 {
		fmt.Printf("\n%d task(s) updated\n", updated)
	}

	return nil
}
