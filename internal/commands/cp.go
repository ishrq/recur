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
	"github.com/ishrq/recur/internal/editor"
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

	filters, remaining := extractFilterFlags(args)

	var ids []int
	var modifyStr string
	var useEditor bool
	foundModifyStr := false

	for i := 0; i < len(remaining); i++ {
		arg := remaining[i]
		switch arg {
		case "--edit", "-e":
			useEditor = true
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

	// --edit flag is set
	if useEditor {
		if len(tasksToCopy) == 1 {
			// Single task copying with editor
			return copySingleTaskInEditor(database, &tasksToCopy[0])
		} else {
			// Multiple task copying with editor
			return copyMultipleTasksInEditor(database, tasksToCopy)
		}
	}

	// inline modifications
	if modifyStr != "" {
		return copyWithModifications(database, tasksToCopy, modifyStr)
	}

	// Default: duplicate tasks in place
	return duplicateTasksInPlace(database, tasksToCopy)
}

func duplicateTasksInPlace(database *sql.DB, tasks []models.Task) error {
	fmt.Printf("\nFound %d task(s) to copy:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := ConfirmPrompt(fmt.Sprintf("Duplicate these %d task(s)? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Duplicate tasks exactly
	copied := 0
	for _, task := range tasks {
		newTask := &models.Task{
			Name:           task.Name,
			DueDate:        task.DueDate,
			CreatedDate:    time.Now(),
			Tag:            task.Tag,
			Project:        task.Project,
			Priority:       task.Priority,
			Note:           task.Note,
			RecurFrequency: task.RecurFrequency,
			RecurEndDate:   task.RecurEndDate,
		}

		newID, err := CreateTask(database, newTask)
		if err != nil {
			fmt.Printf("Warning: Failed to copy task #%d: %v\n", task.ID, err)
			continue
		}

		fmt.Printf("✓ Duplicated #%d → #%d: %s\n", task.ID, newID, newTask.Name)
		copied++
	}

	if copied > 0 {
		fmt.Printf("\n%d task(s) duplicated\n", copied)
	}

	return nil
}

func copyWithModifications(database *sql.DB, tasks []models.Task, modifyStr string) error {
	fmt.Printf("\nFound %d task(s) to copy:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := ConfirmPrompt(fmt.Sprintf("Copy these %d task(s) with modifications? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Parse modification
	parsedChanges, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return fmt.Errorf("failed to parse modifications: %w", err)
	}

	// Copy tasks with modifications
	copied := 0
	for _, task := range tasks {
		// Create new task (copy of original)
		newTask := &models.Task{
			Name:           task.Name,
			DueDate:        task.DueDate,
			CreatedDate:    time.Now(),
			Tag:            task.Tag,
			Project:        task.Project,
			Priority:       task.Priority,
			Note:           task.Note,
			RecurFrequency: task.RecurFrequency,
			RecurEndDate:   task.RecurEndDate,
		}

		// Merge changes
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
		if parsedChanges.RecurFrequency != "" {
			newTask.RecurFrequency = parsedChanges.RecurFrequency
		}
		if parsedChanges.RecurEndDate != nil {
			newTask.RecurEndDate = parsedChanges.RecurEndDate
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

func copySingleTaskInEditor(database *sql.DB, task *models.Task) error {
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

		newTask := &models.Task{
			Name:           editedTask.Name,
			DueDate:        editedTask.DueDate,
			CreatedDate:    time.Now(),
			Tag:            editedTask.Tag,
			Project:        editedTask.Project,
			Priority:       editedTask.Priority,
			Note:           editedTask.Note,
			RecurFrequency: editedTask.RecurFrequency,
			RecurEndDate:   editedTask.RecurEndDate,
		}

		hasChanges := editor.HasChanges(task, editedTask)

		newID, err := CreateTask(database, newTask)
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		if hasChanges {
			fmt.Printf("✓ Created new task #%d: %s\n", newID, newTask.Name)
		} else {
			fmt.Printf("✓ Duplicated #%d → #%d: %s\n", task.ID, newID, newTask.Name)
		}

		return nil
	}
}

func copyMultipleTasksInEditor(database *sql.DB, tasks []models.Task) error {
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

		copied := 0
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

			newTask := &models.Task{
				Name:           editedTask.Name,
				DueDate:        editedTask.DueDate,
				CreatedDate:    time.Now(),
				Tag:            editedTask.Tag,
				Project:        editedTask.Project,
				Priority:       editedTask.Priority,
				Note:           editedTask.Note,
				RecurFrequency: editedTask.RecurFrequency,
				RecurEndDate:   editedTask.RecurEndDate,
			}

			newID, err := CreateTask(database, newTask)
			if err != nil {
				fmt.Printf("Warning: Failed to copy task #%d: %v\n", id, err)
				continue
			}

			hasChanges := editor.HasChanges(original, editedTask)

			if hasChanges {
				fmt.Printf("✓ Created new task from #%d → #%d: %s\n", id, newID, newTask.Name)
			} else {
				fmt.Printf("✓ Duplicated #%d → #%d: %s\n", id, newID, newTask.Name)
			}

			copied++
		}

		if copied > 0 {
			fmt.Printf("\n%d task(s) copied\n", copied)
		}

		return nil
	}
}
