package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/editor"
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

	// Display tasks to be edited
	fmt.Printf("\nFound %d task(s) to edit:\n", len(tasksToEdit))
	for _, t := range tasksToEdit {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := ConfirmPrompt(fmt.Sprintf("Edit these %d task(s)? (y/n): ", len(tasksToEdit)))
	if err != nil {
		return err
	}
	if !ok {
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

func editSingleTaskInEditor(database *sql.DB, task *models.Task) error {
	content := editor.FormatTaskForEditor(task)

	for {
		edited, err := editor.OpenEditor(content)
		if err != nil {
			return err
		}

		editedTask, err := editor.ParseEditedTask(edited)
		if err != nil {
			fmt.Printf("\nError parsing edited task: %v\n", err)
			fmt.Print("Press Enter to re-edit, or Ctrl+C to cancel: ")
			bufio.NewReader(os.Stdin).ReadString('\n')
			content = edited // Use the invalid content so user can fix it
			continue
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
}

func editMultipleTasksInEditor(database *sql.DB, tasks []models.Task) error {
	content := editor.FormatTasksForEditor(tasks)

	for {
		edited, err := editor.OpenEditor(content)
		if err != nil {
			return err
		}

		editedTasks, errors := editor.ParseEditedTasks(edited, tasks)

		if len(errors) > 0 {
			fmt.Println("\nErrors found:")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
			fmt.Print("\nPress Enter to re-edit, or Ctrl+C to cancel: ")
			bufio.NewReader(os.Stdin).ReadString('\n')
			content = edited // Use the invalid content so user can fix it
			continue
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
			// Find original task
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

			// Skip if no changes
			if !editor.HasChanges(original, editedTask) {
				continue
			}

			// Merge changes (preserve non-editable fields)
			editedTask.CreatedDate = original.CreatedDate
			editedTask.CompletedDate = original.CompletedDate
			editedTask.LastTaskID = original.LastTaskID
			editedTask.Deleted = original.Deleted

			// Update task
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
}
