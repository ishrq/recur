package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ishrq/recur/internal/editor"
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

	var tasksToCopy []models.Task
	var err error

	if len(ids) > 0 {
		tasksToCopy = collectTasksByID(database, ids, func(t *models.Task) (bool, string) {
			if t.Deleted {
				return false, fmt.Sprintf("Warning: Task #%d is deleted", t.ID)
			}
			return true, ""
		})
	} else {
		tasksToCopy, err = collectTasksByFilter(database, filters, false, false)
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
	ok, err := confirmTasks(tasks, fmt.Sprintf("Found %d task(s) to copy:", len(tasks)), fmt.Sprintf("Duplicate these %d task(s)? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	copied := 0
	for _, task := range tasks {
		newTask := task.Clone()

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
	ok, err := confirmTasks(tasks, fmt.Sprintf("Found %d task(s) to copy:", len(tasks)), fmt.Sprintf("Copy these %d task(s) with modifications? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	parsedChanges, err := parser.ParseTaskString(modifyStr)
	if err != nil {
		return fmt.Errorf("failed to parse modifications: %w", err)
	}

	// Copy tasks with modifications
	copied := 0
	for _, task := range tasks {
		// Create new task (copy of original) with modifications
		newTask := task.Clone()
		newTask.Merge(parsedChanges)

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

		newTask := editedTask.Clone()

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

			newTask := editedTask.Clone()

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
