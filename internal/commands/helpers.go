package commands

import (
	"database/sql"
	"fmt"

	"github.com/ishrq/recur/internal/db"
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
