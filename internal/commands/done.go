package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
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

		if task.RecurFrequency != "" {
			if err := handleRecurringTask(database, task); err != nil {
				fmt.Printf("Warning: Failed to create next occurrence for task #%d: %v\n", id, err)
			}
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

func handleRecurringTask(database *sql.DB, task *models.Task) error {
	// Parse the frequency
	duration, err := parser.ParseFrequency(task.RecurFrequency)
	if err != nil {
		return fmt.Errorf("invalid frequency '%s': %w", task.RecurFrequency, err)
	}

	// Calculate next due date
	if task.DueDate == nil {
		return fmt.Errorf("recurring task must have a due date")
	}

	nextDueDate := task.DueDate.Add(duration)

	// Check if end date is passed
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
