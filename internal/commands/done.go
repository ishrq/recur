package commands

import (
	"database/sql"
	"fmt"
	"strconv"

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

	return completeTasks(database, ids, filters, undo)
}

func completeTasks(database *sql.DB, ids []int, filters filter.Filters, undo bool) error {
	var err error
	var tasks []models.Task

	if len(ids) > 0 {
		tasks = collectTasksByID(database, ids, func(t *models.Task) (bool, string) {
			if undo && t.CompletedDate == nil {
				return false, fmt.Sprintf("Warning: Task #%d is not completed", t.ID)
			}
			if !undo && t.CompletedDate != nil {
				return false, fmt.Sprintf("Task #%d already completed", t.ID)
			}
			if t.Deleted {
				return false, fmt.Sprintf("Warning: Task #%d is deleted", t.ID)
			}
			return true, ""
		})
	} else {
		tasks, err = collectTasksByFilter(database, filters, false, undo)
		if err != nil {
			return err
		}
	}

	if len(tasks) == 0 {
		if undo {
			return fmt.Errorf("no completed tasks found matching criteria")
		}
		return fmt.Errorf("no tasks found matching criteria")
	}

	var title, prompt string
	if undo {
		title = fmt.Sprintf("Found %d completed task(s) to unmark:", len(tasks))
		prompt = fmt.Sprintf("Unmark these %d task(s) as incomplete? (y/n): ", len(tasks))
	} else {
		title = fmt.Sprintf("Found %d task(s) to complete:", len(tasks))
		prompt = fmt.Sprintf("Mark these %d task(s) as done? (y/n): ", len(tasks))
	}
	ok, err := confirmTasks(tasks, title, prompt)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var count int
	for _, task := range tasks {
		if undo {
			if err := db.UndoCompleteTask(database, task.ID); err != nil {
				fmt.Printf("Warning: Failed to unmark task #%d: %v\n", task.ID, err)
				continue
			}
			fmt.Printf("↺ Unmarked #%d: %s\n", task.ID, task.Name)
			count++
		} else {
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
			count++
		}
	}

	if count > 0 {
		if undo {
			fmt.Printf("\n%d task(s) unmarked as incomplete\n", count)
		} else {
			fmt.Printf("\n%d task(s) completed\n", count)
		}
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

	nextTask := task.Clone()
	nextTask.DueDate = &nextDueDate
	nextTask.LastTaskID = &task.ID

	newID, err := CreateTask(database, nextTask)
	if err != nil {
		return fmt.Errorf("failed to insert next task: %w", err)
	}

	fmt.Printf("  → Created next occurrence #%d (due %s)\n", newID, nextDueDate.Format("Mon Jan 2, 15:04"))

	return nil
}
