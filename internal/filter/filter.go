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
	switch strings.ToLower(dateStr) {
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
	}

	dateRange, err := ParseDateFilter(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date filter '%s': %w", dateStr, err)
	}

	var filtered []models.Task
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		if dateRange.ContainsDate(*task.DueDate) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func filterByDateRange(tasks []models.Task, fromDateStr, toDateStr string) ([]models.Task, error) {
	var fromRange, toRange *DateRange
	var err error

	if fromDateStr != "" {
		fromRange, err = ParseDateFilter(fromDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid from date '%s': %w", fromDateStr, err)
		}
	}

	if toDateStr != "" {
		toRange, err = ParseDateFilter(toDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid to date '%s': %w", toDateStr, err)
		}
	}

	var filtered []models.Task
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())

		if fromRange != nil && taskDate.Before(fromRange.Start) {
			continue
		}

		if toRange != nil && taskDate.After(toRange.End) {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered, nil
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

func filterByField(tasks []models.Task, values []string, getField func(models.Task) string) []models.Task {
	if len(values) == 0 {
		return tasks
	}

	var filtered []models.Task
	for _, task := range tasks {
		taskValue := getField(task)
		for _, value := range values {
			if strings.EqualFold(taskValue, value) {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered
}

func filterByTags(tasks []models.Task, tags []string) []models.Task {
	return filterByField(tasks, tags, func(t models.Task) string { return t.Tag })
}

func filterByProjects(tasks []models.Task, projects []string) []models.Task {
	return filterByField(tasks, projects, func(t models.Task) string { return t.Project })
}

func filterByPriorities(tasks []models.Task, priorities []string) []models.Task {
	return filterByField(tasks, priorities, func(t models.Task) string { return t.Priority })
}
