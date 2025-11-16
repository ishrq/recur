package integration

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
)

// TestDone marks tasks as done without confirmation prompt
func TestDone(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return nil
	}

	var ids []int
	var undo bool

	// Parse args (simplified - only handling IDs for tests)
	for _, arg := range args {
		if arg == "--undo" {
			undo = true
			continue
		}

		// Try to parse as ID
		var id int
		if _, err := fmt.Sscanf(arg, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}

	if undo {
		// Undo logic
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil || task.CompletedDate == nil || task.Deleted {
				continue
			}
			db.UndoCompleteTask(database, id)
		}
		return nil
	}

	// Mark as done
	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil || task.CompletedDate != nil || task.Deleted {
			continue
		}

		// Handle recurring task
		if task.RecurFrequency != "" && task.DueDate != nil {
			nextDueDate, err := parser.CalculateNextOccurrence(*task.DueDate, task.RecurFrequency)
			if err == nil {
				if task.RecurEndDate == nil || !nextDueDate.After(*task.RecurEndDate) {
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
					db.InsertTask(database, nextTask)
				}
			}
		}

		db.MarkTaskDone(database, id)
	}

	return nil
}

// TestRemove deletes tasks without confirmation prompt
func TestRemove(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return nil
	}

	var ids []int
	var undo bool
	var permanentDelete bool

	for _, arg := range args {
		if arg == "--undo" {
			undo = true
			continue
		}
		if arg == "--trash" {
			permanentDelete = true
			continue
		}

		// Try to parse as ID
		var id int
		if _, err := fmt.Sscanf(arg, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}

	if undo {
		// Restore deleted tasks
		for _, id := range ids {
			db.RestoreTask(database, id)
		}
		return nil
	}

	for _, id := range ids {
		if permanentDelete {
			db.PermanentlyDeleteTask(database, id)
		} else {
			db.DeleteTask(database, id)
		}
	}

	return nil
}

// TestCopy copies tasks without confirmation prompt
func TestCopy(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return nil
	}

	var ids []int
	var modifyStr string
	foundModifyStr := false

	for _, arg := range args {
		var id int
		if _, err := fmt.Sscanf(arg, "%d", &id); err == nil {
			ids = append(ids, id)
		} else {
			modifyStr = arg
			foundModifyStr = true
			break
		}
	}

	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil || task.Deleted {
			continue
		}

		newTask := &models.Task{
			Name:        task.Name,
			DueDate:     task.DueDate,
			CreatedDate: time.Now(),
			Tag:         task.Tag,
			Project:     task.Project,
			Priority:    task.Priority,
			Note:        task.Note,
		}

		// Apply modifications if provided
		if foundModifyStr && modifyStr != "" {
			parsed, err := parser.ParseTaskString(modifyStr)
			if err == nil {
				newTask.Name = parsed.Name
				if parsed.DueDate != nil {
					newTask.DueDate = parsed.DueDate
				}
				if parsed.Tag != "" {
					newTask.Tag = parsed.Tag
				}
				if parsed.Project != "" {
					newTask.Project = parsed.Project
				}
				if parsed.Priority != "" {
					newTask.Priority = parsed.Priority
				}
				if parsed.Note != "" {
					newTask.Note = parsed.Note
				}
			}
		}

		db.InsertTask(database, newTask)
	}

	return nil
}

// TestMove edits tasks without confirmation prompt
func TestMove(database *sql.DB, args []string) error {
	if len(args) < 2 {
		return nil
	}

	var ids []int
	var modifyStr string

	for _, arg := range args {
		var id int
		if _, err := fmt.Sscanf(arg, "%d", &id); err == nil {
			ids = append(ids, id)
		} else {
			modifyStr = arg
			break
		}
	}

	if modifyStr == "" {
		return nil
	}

	parsed, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return err
	}

	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil || task.Deleted {
			continue
		}

		// Merge changes
		task.Name = parsed.Name
		if parsed.DueDate != nil {
			task.DueDate = parsed.DueDate
		}
		if parsed.Tag != "" {
			task.Tag = parsed.Tag
		}
		if parsed.Project != "" {
			task.Project = parsed.Project
		}
		if parsed.Priority != "" {
			task.Priority = parsed.Priority
		}
		if parsed.Note != "" {
			task.Note = parsed.Note
		}

		db.UpdateTask(database, task)
	}

	return nil
}
