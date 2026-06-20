package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/editor"
	"github.com/ishrq/recur/internal/filter"
	"github.com/ishrq/recur/internal/models"
)

func collectTasksByID(database *sql.DB, ids []int, guardFn func(*models.Task) (bool, string)) []models.Task {
	var tasks []models.Task
	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil {
			fmt.Printf("Warning: Task #%d not found\n", id)
			continue
		}
		if ok, warn := guardFn(task); ok {
			tasks = append(tasks, *task)
		} else if warn != "" {
			fmt.Println(warn)
		}
	}
	return tasks
}

func collectTasksByFilter(database *sql.DB, filters filter.Filters, deleted, completed bool) ([]models.Task, error) {
	tasks, err := db.GetTasks(database, deleted, completed)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	return filter.ApplyFilters(tasks, filters)
}

func openTaskInEditor(task *models.Task) (*models.Task, error) {
	content := editor.FormatTaskForEditor(task)
	for {
		edited, err := editor.OpenEditor(content)
		if err != nil {
			return nil, err
		}
		editedTask, err := editor.ParseEditedTask(edited)
		if err != nil {
			fmt.Printf("\nError parsing edited task: %v\n", err)
			fmt.Print("Press Enter to re-edit, or Ctrl+C to cancel: ")
			bufio.NewReader(os.Stdin).ReadString('\n')
			content = edited
			continue
		}
		return editedTask, nil
	}
}

func openTasksInEditor(tasks []models.Task) (map[int]*models.Task, error) {
	content := editor.FormatTasksForEditor(tasks)
	for {
		edited, err := editor.OpenEditor(content)
		if err != nil {
			return nil, err
		}
		editedTasks, errors := editor.ParseEditedTasks(edited, tasks)
		if len(errors) > 0 {
			fmt.Println("\nErrors found:")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
			fmt.Print("\nPress Enter to re-edit, or Ctrl+C to cancel: ")
			bufio.NewReader(os.Stdin).ReadString('\n')
			content = edited
			continue
		}
		return editedTasks, nil
	}
}
