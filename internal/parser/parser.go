package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/models"
)

// ParseTaskString parses a task string and extracts task components
func ParseTaskString(input string) (*models.Task, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	task := &models.Task{
		CreatedDate: time.Now(),
	}

	remaining := input

	// Extract in order: date, note, tags, project, priority
	// This order matters because we remove matched parts progressively
	var err error
	remaining, err = extractDueDate(remaining, task)
	if err != nil {
		return nil, err
	}
	remaining = extractNote(remaining, task)
	remaining = extractTags(remaining, task)
	remaining = extractProject(remaining, task)
	remaining = extractPriority(remaining, task)

	// Whatever remains is the task name
	task.Name = strings.TrimSpace(remaining)

	if task.Name == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	return task, nil
}

// parseDueDateString parses the date string format: "date time, frequency, end"
// Returns error if parsing fails
func parseDueDateString(dateStr string, task *models.Task) error {
	parts := strings.Split(dateStr, ",")
	var errors []string

	// Parse first part: date and time
	if len(parts) > 0 {
		dateTimeStr := strings.TrimSpace(parts[0])
		if dateTimeStr != "" {
			parsedDate, err := ParseTaskDate(dateTimeStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid date '%s': %v", dateTimeStr, err))
			} else {
				task.DueDate = parsedDate
			}
		}
	}

	// Parse second part: frequency (e.g., 1d, 1w, 2m, 1y, 12h)
	if len(parts) > 1 {
		frequencyStr := strings.TrimSpace(parts[1])
		if frequencyStr != "" {
			if err := ValidateFrequency(frequencyStr); err != nil {
				errors = append(errors, err.Error())
			} else {
				task.RecurFrequency = frequencyStr
			}
		}
	}

	// Parse third part: end date
	if len(parts) > 2 {
		endDateStr := strings.TrimSpace(parts[2])
		if endDateStr != "" {
			parsedEndDate, err := ParseTaskDate(endDateStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid end date '%s': %v", endDateStr, err))
			} else {
				task.RecurEndDate = parsedEndDate
			}
		}
	}

	// Validate invariant: frequency requires a due date
	if task.RecurFrequency != "" && task.DueDate == nil {
		errors = append(errors, "frequency specified without a valid due date")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}
