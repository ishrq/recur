package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

func getFilteredTasks(database *sql.DB, showAll, showDone, showTrash, showToday, showTomorrow, showOverdue, showUpcoming bool,
	dueDate, fromDate, toDate, query string, tags, projects, priorities []string) ([]models.Task, error) {

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextWeek := today.AddDate(0, 0, 7)

	// Determine which tasks to fetch
	var allTasks []models.Task
	var err error

	if showTrash {
		// Get only deleted
		allTasks, err = db.GetDeletedTasks(database)
		if err != nil {
			return nil, err
		}
	} else if showDone {
		// Get only completed
		allTasks, err = db.GetTasks(database, true) // Get all including completed
		if err != nil {
			return nil, err
		}
		// Filter to only completed
		var completedTasks []models.Task
		for _, task := range allTasks {
			if task.CompletedDate != nil {
				completedTasks = append(completedTasks, task)
			}
		}
		allTasks = completedTasks
	} else {
		// Get incomplete tasks (or all if showAll is true)
		allTasks, err = db.GetTasks(database, showAll)
		if err != nil {
			return nil, err
		}
	}

	if dueDate != "" {
		allTasks, err = filterByDate(allTasks, dueDate)
		if err != nil {
			return nil, err
		}
	}

	if fromDate != "" || toDate != "" {
		allTasks, err = filterByDateRange(allTasks, fromDate, toDate)
		if err != nil {
			return nil, err
		}
	}

	if showOverdue || showToday || showTomorrow || showUpcoming {
		var dateFiltered []models.Task
		for _, task := range allTasks {
			if task.DueDate == nil {
				continue
			}

			taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, now.Location())

			if showOverdue && taskDate.Before(today) {
				dateFiltered = append(dateFiltered, task)
			} else if showToday && taskDate.Equal(today) {
				dateFiltered = append(dateFiltered, task)
			} else if showTomorrow && taskDate.Equal(today.AddDate(0, 0, 1)) {
				dateFiltered = append(dateFiltered, task)
			} else if showUpcoming && taskDate.After(today) && taskDate.Before(nextWeek.AddDate(0, 0, 1)) {
				dateFiltered = append(dateFiltered, task)
			}
		}
		allTasks = dateFiltered
	}

	if query != "" {
		allTasks = filterByQuery(allTasks, query)
	}

	if len(tags) > 0 {
		allTasks = filterByTags(allTasks, tags)
	}

	if len(projects) > 0 {
		allTasks = filterByProjects(allTasks, projects)
	}

	if len(priorities) > 0 {
		allTasks = filterByPriorities(allTasks, priorities)
	}

	return allTasks, nil
}

func filterByDate(tasks []models.Task, dateStr string) ([]models.Task, error) {
	now := time.Now()
	var targetDate time.Time

	switch strings.ToLower(dateStr) {
	case "today":
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "tomorrow":
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	case "none":
		var filtered []models.Task
		for _, task := range tasks {
			if task.DueDate == nil {
				filtered = append(filtered, task)
			}
		}
		return filtered, nil
	case "recurring":
		var filtered []models.Task
		for _, task := range tasks {
			if task.RecurFrequency != "" {
				filtered = append(filtered, task)
			}
		}
		return filtered, nil
	default:
		parsedDate, err := parseDate(dateStr)
		if err != nil {
			return nil, err
		}
		targetDate = parsedDate
	}

	var filtered []models.Task
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())
		if taskDate.Equal(targetDate) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func filterByDateRange(tasks []models.Task, fromDateStr, toDateStr string) ([]models.Task, error) {
	now := time.Now()
	var fromDate, toDate *time.Time

	if fromDateStr != "" {
		parsed, err := parseDate(fromDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid from date: %w", err)
		}
		fromDate = &parsed
	}

	if toDateStr != "" {
		parsed, err := parseDate(toDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid to date: %w", err)
		}
		toDate = &parsed
	}

	var filtered []models.Task
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, now.Location())

		if fromDate != nil && taskDate.Before(*fromDate) {
			continue
		}

		if toDate != nil && taskDate.After(*toDate) {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered, nil
}

func parseDate(dateStr string) (time.Time, error) {
	now := time.Now()

	switch strings.ToLower(dateStr) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "tomorrow":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1), nil
	}

	formats := []string{
		"2006-01-02",
		"Jan 2",
		"Jan 2 2006",
		"Monday",
		"Mon",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			}
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location()), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

func filterByTags(tasks []models.Task, tags []string) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		for _, tag := range tags {
			if strings.EqualFold(task.Tag, tag) {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered
}

func filterByProjects(tasks []models.Task, projects []string) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		for _, project := range projects {
			if strings.EqualFold(task.Project, project) {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered
}

func filterByPriorities(tasks []models.Task, priorities []string) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		for _, priority := range priorities {
			if strings.EqualFold(task.Priority, priority) {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered
}

func filterByQuery(tasks []models.Task, query string) []models.Task {
	query = strings.ToLower(query)
	var filtered []models.Task

	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.Name), query) ||
		strings.Contains(strings.ToLower(task.Tag), query) ||
		strings.Contains(strings.ToLower(task.Project), query) ||
		strings.Contains(strings.ToLower(task.Note), query) {
			filtered = append(filtered, task)
		}
	}

	return filtered
}
