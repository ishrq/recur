package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/editor"
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

	filters, remaining := extractFilterFlags(args)

	var ids []int
	var modifyStr string
	foundModifyStr := false

	for i := 0; i < len(remaining); i++ {
		arg := remaining[i]
		switch arg {
		default:
			id, err := strconv.Atoi(arg)
			if err != nil {
				modifyStr = strings.Join(remaining[i:], " ")
				foundModifyStr = true
				break
			}
			ids = append(ids, id)
		}
		if foundModifyStr {
			break
		}
	}

	var tasksToEdit []models.Task
	var err error

	if len(ids) > 0 {
		tasksToEdit = collectTasksByID(database, ids, func(t *models.Task) (bool, string) {
			if t.Deleted {
				return false, fmt.Sprintf("Warning: Task #%d is deleted", t.ID)
			}
			return true, ""
		})
	} else {
		tasksToEdit, err = collectTasksByFilter(database, filters, false, false)
		if err != nil {
			return err
		}
	}

	if len(tasksToEdit) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// If no modification string provided, use $EDITOR
	if modifyStr == "" {
		if len(tasksToEdit) == 1 {
			// Single task editing
			return editSingleTaskInEditor(database, &tasksToEdit[0])
		} else {
			// Multiple task editing
			return editMultipleTasksInEditor(database, tasksToEdit)
		}
	}

	ok, err := confirmTasks(tasksToEdit, fmt.Sprintf("Found %d task(s) to edit:", len(tasksToEdit)), fmt.Sprintf("Edit these %d task(s)? (y/n): ", len(tasksToEdit)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	parsedChanges, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return fmt.Errorf("failed to parse changes: %w", err)
	}

	// Edit tasks
	updated := 0
	for _, task := range tasksToEdit {
		oldName := task.Name

		// Merge changes (new values override old ones)
		task.Merge(parsedChanges)

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

func editSingleTaskInEditor(database *sql.DB, task *models.Task) error {
	editedTask, err := openTaskInEditor(task)
	if err != nil {
		return err
	}

	if !editor.HasChanges(task, editedTask) {
		fmt.Println("No changes detected. Edit cancelled.")
		return nil
	}

	editedTask.ID = task.ID
	editedTask.CreatedDate = task.CreatedDate
	editedTask.CompletedDate = task.CompletedDate
	editedTask.LastTaskID = task.LastTaskID
	editedTask.Deleted = task.Deleted

	if err := db.UpdateTask(database, editedTask); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✓ Updated #%d: %s\n", task.ID, editedTask.Name)
	return nil
}

func editMultipleTasksInEditor(database *sql.DB, tasks []models.Task) error {
	editedTasks, err := openTasksInEditor(tasks)
	if err != nil {
		return err
	}

	hasAnyChanges := false
	for id, editedTask := range editedTasks {
		for _, original := range tasks {
			if original.ID == id {
				if editor.HasChanges(&original, editedTask) {
					hasAnyChanges = true
					break
				}
			}
		}
	}

	if !hasAnyChanges {
		fmt.Println("No changes detected. Edit cancelled.")
		return nil
	}

	updated := 0
	for id, editedTask := range editedTasks {
		var original *models.Task
		for i := range tasks {
			if tasks[i].ID == id {
				original = &tasks[i]
				break
			}
		}

		if original == nil {
			continue
		}

		if !editor.HasChanges(original, editedTask) {
			continue
		}

		editedTask.CreatedDate = original.CreatedDate
		editedTask.CompletedDate = original.CompletedDate
		editedTask.LastTaskID = original.LastTaskID
		editedTask.Deleted = original.Deleted

		if err := db.UpdateTask(database, editedTask); err != nil {
			fmt.Printf("Warning: Failed to update task #%d: %v\n", id, err)
			continue
		}

		fmt.Printf("✓ Updated #%d: %s\n", id, editedTask.Name)
		updated++
	}

	if updated > 0 {
		fmt.Printf("\n%d task(s) updated\n", updated)
	}

	return nil
}
