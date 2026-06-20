package commands

import (
	"database/sql"
	"fmt"
	"strconv"
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
		return undoCompleteTasks(database, ids, filters)
	}
	return completeTasks(database, ids, filters)
}

func undoCompleteTasks(database *sql.DB, ids []int, filters filter.Filters) error {
	var tasks []models.Task
	var err error

	if len(ids) > 0 {
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
			tasks = append(tasks, *task)
		}
	} else {
		tasks, err = db.GetTasks(database, false, true)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
		tasks, err = filter.ApplyFilters(tasks, filters)
		if err != nil {
			return err
		}
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no completed tasks found matching criteria")
	}

	fmt.Printf("\nFound %d completed task(s) to unmark:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := confirmPrompt(fmt.Sprintf("Unmark these %d task(s) as incomplete? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Operation cancelled.")
		return nil
	}

	unmarked := 0
	for _, task := range tasks {
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

func completeTasks(database *sql.DB, ids []int, filters filter.Filters) error {
	var tasks []models.Task
	var err error

	if len(ids) > 0 {
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
			tasks = append(tasks, *task)
		}
	} else {
		tasks, err = db.GetTasks(database, false, false)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
		tasks, err = filter.ApplyFilters(tasks, filters)
		if err != nil {
			return err
		}
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	fmt.Printf("\nFound %d task(s) to complete:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := confirmPrompt(fmt.Sprintf("Mark these %d task(s) as done? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Operation cancelled.")
		return nil
	}

	completed := 0
	for _, task := range tasks {
		if task.RecurFrequency != "" {
			if err := handleRecurringTask(database, &task); err != nil {
				fmt.Printf("Warning: Failed to create next occurrence for task #%d: %v\n", task.ID, err)
			}
		}
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
	if task.DueDate == nil {
		return fmt.Errorf("recurring task must have a due date")
	}

	nextDueDate, err := parser.CalculateNextOccurrence(*task.DueDate, task.RecurFrequency)
	if err != nil {
		return fmt.Errorf("invalid frequency '%s': %w", task.RecurFrequency, err)
	}

	if task.RecurEndDate != nil && nextDueDate.After(*task.RecurEndDate) {
		fmt.Printf("  → Recurring task ended (past end date)\n")
		return nil
	}

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
