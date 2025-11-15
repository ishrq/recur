package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/filter"
	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
)

func Copy(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("cp")
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

	// Collect tasks to copy
	var tasksToCopy []models.Task
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
			tasksToCopy = append(tasksToCopy, *task)
		}
	} else {
		// Get all incomplete tasks for filtering
		tasksToCopy, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Apply filters
		tasksToCopy, err = filter.ApplyFilters(tasksToCopy, filters)
		if err != nil {
			return err
		}
	}

	if len(tasksToCopy) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// Display tasks to be copied
	fmt.Printf("\nFound %d task(s) to copy:\n", len(tasksToCopy))
	for _, t := range tasksToCopy {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	// Ask for confirmation
	fmt.Printf("Copy these %d task(s)? (y/n): ", len(tasksToCopy))
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

	// Parse modification if provided
	var parsedChanges *models.Task
	if modifyStr != "" {
		parsedChanges, err = parser.ParseTaskString(modifyStr)
		if err != nil {
			return fmt.Errorf("failed to parse modifications: %w", err)
		}
	}

	// Copy tasks
	copied := 0
	for _, task := range tasksToCopy {
		// Create new task (copy of original)
		newTask := &models.Task{
			Name:        task.Name,
			DueDate:     task.DueDate,
			CreatedDate: time.Now(),
			Tag:         task.Tag,
			Project:     task.Project,
			Priority:    task.Priority,
			Note:        task.Note,
		}

		// If modification string provided, merge changes
		if parsedChanges != nil {
			newTask.Name = parsedChanges.Name
			if parsedChanges.DueDate != nil {
				newTask.DueDate = parsedChanges.DueDate
			}
			if parsedChanges.Tag != "" {
				newTask.Tag = parsedChanges.Tag
			}
			if parsedChanges.Project != "" {
				newTask.Project = parsedChanges.Project
			}
			if parsedChanges.Priority != "" {
				newTask.Priority = parsedChanges.Priority
			}
			if parsedChanges.Note != "" {
				newTask.Note = parsedChanges.Note
			}
		}

		newID, err := CreateTask(database, newTask)
		if err != nil {
			fmt.Printf("Warning: Failed to copy task #%d: %v\n", task.ID, err)
			continue
		}

		fmt.Printf("✓ Copied #%d → #%d: %s\n", task.ID, newID, newTask.Name)
		copied++
	}

	if copied > 0 {
		fmt.Printf("\n%d task(s) copied\n", copied)
	}

	return nil
}
