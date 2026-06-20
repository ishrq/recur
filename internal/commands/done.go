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

func Done(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("done")
		return nil
	}

	filters, remaining := extractFilterFlags(args)

	var ids []int
	var undo bool

	for _, arg := range remaining {
		switch arg {
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

	if undo {
		var tasksToUndo []models.Task
		var err error

		if len(ids) > 0 {
			// Get tasks by IDs
			for _, id := range ids {
				task, err := db.GetTaskByID(database, id)
				if err != nil {
					fmt.Printf("Warning: Task #%d not found\n", id)
					continue
				}
				if task.CompletedDate == nil {
					fmt.Printf("Warning: Task #%d is not completed\n", id)
					continue
				}
				if task.Deleted {
					fmt.Printf("Warning: Task #%d is deleted\n", id)
					continue
				}
				tasksToUndo = append(tasksToUndo, *task)
			}
		} else {
			// Get all completed tasks for filtering
			tasksToUndo, err = db.GetTasks(database, false, true)
			if err != nil {
				return fmt.Errorf("failed to get tasks: %w", err)
			}

			// Apply filters
			tasksToUndo, err = filter.ApplyFilters(tasksToUndo, filters)
			if err != nil {
				return err
			}
		}

		if len(tasksToUndo) == 0 {
			return fmt.Errorf("no completed tasks found matching criteria")
		}

		// Display tasks to be unmarked
		fmt.Printf("\nFound %d completed task(s) to unmark:\n", len(tasksToUndo))
		for _, t := range tasksToUndo {
			fmt.Printf("#%-4d %s\n", t.ID, t.Name)
		}
		fmt.Println()

		// Ask for confirmation
		fmt.Printf("Unmark these %d task(s) as incomplete? (y/n): ", len(tasksToUndo))
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

		// Unmark tasks
		unmarked := 0
		for _, task := range tasksToUndo {
			if err := db.UndoCompleteTask(database, task.ID); err != nil {
				fmt.Printf("Warning: Failed to unmark task #%d: %v\n", task.ID, err)
				continue
			}

			fmt.Printf("↺ Unmarked #%d: %s\n", task.ID, task.Name)
			unmarked++
		}

		if unmarked > 0 {
			fmt.Printf("\n%d task(s) unmarked as incomplete\n", unmarked)
		}

		return nil
	}

	// Collect tasks to mark as done
	var tasksToComplete []models.Task
	var err error

	if len(ids) > 0 {
		// Get tasks by IDs
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
			if task.Deleted {
				fmt.Printf("Task #%d is deleted\n", id)
				continue
			}
			tasksToComplete = append(tasksToComplete, *task)
		}
	} else {
		// Get all incomplete tasks for filtering
		tasksToComplete, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Apply filters
		tasksToComplete, err = filter.ApplyFilters(tasksToComplete, filters)
		if err != nil {
			return err
		}
	}

	if len(tasksToComplete) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// Display tasks to be completed
	fmt.Printf("\nFound %d task(s) to complete:\n", len(tasksToComplete))
	for _, t := range tasksToComplete {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	// Ask for confirmation
	fmt.Printf("Mark these %d task(s) as done? (y/n): ", len(tasksToComplete))
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

	// Mark tasks as done
	completed := 0
	for _, task := range tasksToComplete {
		// Check if this is a recurring task
		if task.RecurFrequency != "" {
			if err := handleRecurringTask(database, &task); err != nil {
				fmt.Printf("Warning: Failed to create next occurrence for task #%d: %v\n", task.ID, err)
			}
		}

		// Mark current task as done
		if err := db.MarkTaskDone(database, task.ID); err != nil {
			fmt.Printf("Warning: Failed to complete task #%d: %v\n", task.ID, err)
			continue
		}

		fmt.Printf("✓ Completed #%d: %s\n", task.ID, task.Name)
		completed++
	}

	if completed > 0 {
		fmt.Printf("\n%d task(s) completed\n", completed)
	}

	return nil
}

func handleRecurringTask(database *sql.DB, task *models.Task) error {
	// Validate recurring task has a due date
	if task.DueDate == nil {
		return fmt.Errorf("recurring task must have a due date")
	}

	// Calculate next due date using calendar-aware function
	nextDueDate, err := parser.CalculateNextOccurrence(*task.DueDate, task.RecurFrequency)
	if err != nil {
		return fmt.Errorf("invalid frequency '%s': %w", task.RecurFrequency, err)
	}

	// Check if we've passed the end date
	if task.RecurEndDate != nil && nextDueDate.After(*task.RecurEndDate) {
		fmt.Printf("  → Recurring task ended (past end date)\n")
		return nil
	}

	// Create the next occurrence
	nextTask := &models.Task{
		Name:           task.Name,
		DueDate:        &nextDueDate,
		CreatedDate:    time.Now(),
		Tag:            task.Tag,
		Project:        task.Project,
		Priority:       task.Priority,
		Note:           task.Note,
		RecurFrequency: task.RecurFrequency,
		RecurEndDate:   task.RecurEndDate,
		LastTaskID:     &task.ID,
	}

	newID, err := CreateTask(database, nextTask)
	if err != nil {
		return fmt.Errorf("failed to insert next task: %w", err)
	}

	fmt.Printf("  → Created next occurrence #%d (due %s)\n", newID, nextDueDate.Format("Mon Jan 2, 15:04"))

	return nil
}
