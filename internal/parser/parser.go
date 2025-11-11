package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/models"
)

// ParseTaskString parses a task string and extracts task components
func ParseTaskString(input string) (*models.Task, error) {
	task := &models.Task{
		CreatedDate: time.Now(),
	}

	remaining := input

	remaining = extractDueDate(remaining, task)
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

func extractDueDate(input string, task *models.Task) string {
	// Match @(...) with proper parenthesis handling
	re := regexp.MustCompile(`@\(([^)]+)\)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		dateStr := matches[1]
		// NOTE: Parse the date string (basic implementation for now)
		// Implement full date parsing later
		parseDueDateString(dateStr, task)

		// Remove the matched portion from input
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractNote(input string, task *models.Task) string {
	// Match *'...' or *"..."
	re := regexp.MustCompile(`\*['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Note = matches[1]
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractTags(input string, task *models.Task) string {
	// Match #word (not followed by another #)
	re := regexp.MustCompile(`#(\w+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) > 0 {
		// NOTE: Use the first tag (we can extend to multiple tags later)
		task.Tag = matches[0][1]
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractProject(input string, task *models.Task) string {
	// Match +word (projects use + prefix)
	re := regexp.MustCompile(`\+(\w[\w-]*)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Project = matches[1]
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractPriority(input string, task *models.Task) string {
	// Match !word (priorities use ! prefix)
	re := regexp.MustCompile(`!(\w+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Priority = matches[1]
		return re.ReplaceAllString(input, "")
	}

	return input
}

// NOTE: Parse the due date string (basic implementation)
// Format: "date time, frequency, end"
func parseDueDateString(dateStr string, task *models.Task) {
	parts := strings.Split(dateStr, ",")

	// Parse first part: date and time
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		dateTimeStr := strings.TrimSpace(parts[0])
		parsedDate := parseDateTime(dateTimeStr)
		if parsedDate != nil {
			task.DueDate = parsedDate
		}
	}

	// TODO: Parse frequency (second part) - will implement in next step
	// TODO: Parse end date (third part) - will implement in next step
}

// parseDateTime parses various date/time formats
func parseDateTime(input string) *time.Time {
	input = strings.TrimSpace(input)
	now := time.Now()

	// Handle semantic dates
	switch strings.ToLower(input) {
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
		return &t
	case "tomorrow":
		t := now.AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t
	case "next week":
		t := now.AddDate(0, 0, 7)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t
	}

	// Parse common formats
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 3:04pm",
		"2006-01-02 3:04 pm",
		"2006-01-02",
		"Jan 2 15:04",
		"Jan 2 3:04pm",
		"Jan 2",
		"Monday",
		"Mon",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// If year is 0, set it to current year
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
			}
			// If only date is specified (no time), set default time to 12:00
			if t.Hour() == 0 && t.Minute() == 0 && !strings.Contains(input, ":") {
				t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
			}
			return &t
		}
	}

	// Parse weekday names
	weekdays := map[string]time.Weekday{
		"sunday": time.Sunday, "sun": time.Sunday,
		"monday": time.Monday, "mon": time.Monday,
		"tuesday": time.Tuesday, "tue": time.Tuesday,
		"wednesday": time.Wednesday, "wed": time.Wednesday,
		"thursday": time.Thursday, "thu": time.Thursday,
		"friday": time.Friday, "fri": time.Friday,
		"saturday": time.Saturday, "sat": time.Saturday,
	}

	if weekday, ok := weekdays[strings.ToLower(input)]; ok {
		daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
		if daysUntil == 0 {
			daysUntil = 7 // Next occurrence
		}
		t := now.AddDate(0, 0, daysUntil)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t
	}

	return nil
}
