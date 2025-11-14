package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/models"
)

type Filters struct {
	Today      bool
	Tomorrow   bool
	Overdue    bool
	Upcoming   bool
	DueDate    string
	FromDate   string
	ToDate     string
	Query      string
	Tags       []string
	Projects   []string
	Priorities []string
}

func ApplyFilters(tasks []models.Task, filters Filters) ([]models.Task, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextWeek := today.AddDate(0, 0, 7)

	if filters.DueDate != "" {
		filtered, err := filterByDate(tasks, filters.DueDate)
		if err != nil {
			return nil, err
		}
		tasks = filtered
	}

	if filters.FromDate != "" || filters.ToDate != "" {
		filtered, err := filterByDateRange(tasks, filters.FromDate, filters.ToDate)
		if err != nil {
			return nil, err
		}
		tasks = filtered
	}

	if filters.Overdue || filters.Today || filters.Tomorrow || filters.Upcoming {
		var dateFiltered []models.Task
		for _, task := range tasks {
			if task.DueDate == nil {
				continue
			}

			taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, now.Location())

			if filters.Overdue && taskDate.Before(today) {
				dateFiltered = append(dateFiltered, task)
			} else if filters.Today && taskDate.Equal(today) {
				dateFiltered = append(dateFiltered, task)
			} else if filters.Tomorrow && taskDate.Equal(today.AddDate(0, 0, 1)) {
				dateFiltered = append(dateFiltered, task)
			} else if filters.Upcoming && taskDate.After(today) && taskDate.Before(nextWeek.AddDate(0, 0, 1)) {
				dateFiltered = append(dateFiltered, task)
			}
		}
		tasks = dateFiltered
	}

	if filters.Query != "" {
		tasks = filterByQuery(tasks, filters.Query)
	}

	if len(filters.Tags) > 0 {
		tasks = filterByTags(tasks, filters.Tags)
	}

	if len(filters.Projects) > 0 {
		tasks = filterByProjects(tasks, filters.Projects)
	}

	if len(filters.Priorities) > 0 {
		tasks = filterByPriorities(tasks, filters.Priorities)
	}

	return tasks, nil
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
		parsedDate, err := ParseDate(dateStr)
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
		parsed, err := ParseDate(fromDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid from date: %w", err)
		}
		fromDate = &parsed
	}

	if toDateStr != "" {
		parsed, err := ParseDate(toDateStr)
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

func ParseDate(dateStr string) (time.Time, error) {
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
