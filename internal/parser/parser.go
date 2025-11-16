package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/models"
)

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
		if err := parseDueDateString(dateStr, task); err != nil {
			// Return input unchanged if parsing fails
			// Could optionally log the error
			return input
		}
		// Remove the matched portion from input
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractNote(input string, task *models.Task) string {
	// Match *'...' or *"..." with support for escaped quotes
	re := regexp.MustCompile(`\*['"]([^'"]*(?:\\['"][^'"]*)*)['"]`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		// Unescape any escaped quotes
		note := matches[1]
		note = strings.ReplaceAll(note, `\'`, `'`)
		note = strings.ReplaceAll(note, `\"`, `"`)
		task.Note = note
		return re.ReplaceAllString(input, "")
	}

	return input
}

func extractTags(input string, task *models.Task) string {
	// Match #word (alphanumeric, underscore, hyphen)
	re := regexp.MustCompile(`#([\w-]+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) > 0 {
		// Use the first tag for now
		task.Tag = matches[0][1]
		// Remove only the first occurrence to avoid conflicts with other # symbols
		return strings.Replace(input, matches[0][0], "", 1)
	}

	return input
}

func extractProject(input string, task *models.Task) string {
	// Match +word (alphanumeric, underscore, hyphen)
	re := regexp.MustCompile(`\+([\w-]+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Project = matches[1]
		return strings.Replace(input, matches[0], "", 1)
	}

	return input
}

func extractPriority(input string, task *models.Task) string {
	// Match !word (alphanumeric, underscore, hyphen)
	re := regexp.MustCompile(`!([\w-]+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Priority = matches[1]
		return strings.Replace(input, matches[0], "", 1)
	}

	return input
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
			parsedDate, err := parseDateTime(dateTimeStr)
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
			parsedEndDate, err := parseDateTime(endDateStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid end date '%s': %v", endDateStr, err))
			} else {
				task.RecurEndDate = parsedEndDate
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateFrequency checks if a frequency string is valid
func ValidateFrequency(freq string) error {
	// Support formats: 1d, 2w, 3m, 1y, 12h
	if matched, _ := regexp.MatchString(`^\d+[hdwmy]$`, freq); matched {
		// Validate the number is positive
		re := regexp.MustCompile(`^(\d+)[hdwmy]$`)
		matches := re.FindStringSubmatch(freq)
		if len(matches) == 2 {
			num, _ := strconv.Atoi(matches[1])
			if num <= 0 {
				return fmt.Errorf("frequency number must be positive: %s", freq)
			}
		}
		return nil
	}

	// Support semantic frequencies
	semantic := []string{"hourly", "daily", "weekly", "monthly", "yearly"}
	for _, s := range semantic {
		if strings.EqualFold(freq, s) {
			return nil
		}
	}

	return fmt.Errorf("invalid frequency format: %s (expected: 1h, 1d, 1w, 1m, 1y, or daily/weekly/monthly/yearly)", freq)
}

// CalculateNextOccurrence calculates the next occurrence date based on frequency
// This function is calendar-aware and handles months/years correctly
func CalculateNextOccurrence(baseDate time.Time, frequency string) (time.Time, error) {
	if err := ValidateFrequency(frequency); err != nil {
		return time.Time{}, err
	}

	// Handle semantic frequencies
	switch strings.ToLower(frequency) {
	case "hourly":
		return baseDate.Add(time.Hour), nil
	case "daily":
		return baseDate.AddDate(0, 0, 1), nil
	case "weekly":
		return baseDate.AddDate(0, 0, 7), nil
	case "monthly":
		return baseDate.AddDate(0, 1, 0), nil
	case "yearly":
		return baseDate.AddDate(1, 0, 0), nil
	}

	// Parse numeric format: 2h, 3d, 1w, 2m, 1y
	re := regexp.MustCompile(`^(\d+)([hdwmy])$`)
	matches := re.FindStringSubmatch(frequency)

	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("invalid frequency format: %s", frequency)
	}

	num, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "h":
		// Use time.Add for hours (handles DST correctly)
		return baseDate.Add(time.Duration(num) * time.Hour), nil
	case "d":
		return baseDate.AddDate(0, 0, num), nil
	case "w":
		return baseDate.AddDate(0, 0, num*7), nil
	case "m":
		return baseDate.AddDate(0, num, 0), nil
	case "y":
		return baseDate.AddDate(num, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown frequency unit: %s", unit)
	}
}

// parseDateTime parses various date/time formats
// Returns error if parsing fails
func parseDateTime(input string) (*time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty date string")
	}

	now := time.Now()

	// Handle semantic dates
	switch strings.ToLower(input) {
	case "now":
		return &now, nil
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
		return &t, nil
	case "tomorrow", "tmr":
		t := now.AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t, nil
	case "yesterday":
		t := now.AddDate(0, 0, -1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t, nil
	case "next week":
		t := now.AddDate(0, 0, 7)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t, nil
	case "next month":
		t := now.AddDate(0, 1, 0)
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		return &t, nil
	}

	// Handle relative dates: +3d, -2w, +1m, etc.
	if relativeTime, ok := parseRelativeTime(input, now); ok {
		return relativeTime, nil
	}

	// Parse weekday names (next occurrence)
	if weekdayTime, ok := parseWeekday(input, now); ok {
		return weekdayTime, nil
	}

	// Try standard formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 3:04pm",
		"2006-01-02 3:04 pm",
		"2006-01-02 3:04PM",
		"2006-01-02 3:04 PM",
		"2006-01-02 3pm",
		"2006-01-02 3 pm",
		"2006-01-02 3PM",
		"2006-01-02 3 PM",
		"2006-01-02",
		"2006/01/02 15:04",
		"2006/01/02",
		"Jan 2 15:04",
		"Jan 2 3:04pm",
		"Jan 2 3pm",
		"Jan 2",
		"January 2 15:04",
		"January 2",
		"02 Jan 2006",
		"02-01-2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// If year is 0 or appears to be missing, set it to current year
			if t.Year() == 0 || t.Year() < 1000 {
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
			}

			// If only date is specified (no time), set default time to 12:00
			if t.Hour() == 0 && t.Minute() == 0 && !strings.Contains(input, ":") &&
				!strings.Contains(strings.ToLower(input), "am") &&
				!strings.Contains(strings.ToLower(input), "pm") {
				t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
			}

			return &t, nil
		}
	}

	return nil, fmt.Errorf("unrecognized date format")
}

// parseRelativeTime parses relative dates like +3d, -2w, +1m
func parseRelativeTime(input string, now time.Time) (*time.Time, bool) {
	re := regexp.MustCompile(`^([+-])(\d+)([hdwmy])$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 4 {
		return nil, false
	}

	sign := matches[1]
	num, _ := strconv.Atoi(matches[2])
	unit := matches[3]

	if sign == "-" {
		num = -num
	}

	var t time.Time
	switch unit {
	case "h":
		t = now.Add(time.Duration(num) * time.Hour)
	case "d":
		t = now.AddDate(0, 0, num)
	case "w":
		t = now.AddDate(0, 0, num*7)
	case "m":
		t = now.AddDate(0, num, 0)
	case "y":
		t = now.AddDate(num, 0, 0)
	default:
		return nil, false
	}

	// Set to 12:00 for date-based units
	if unit != "h" {
		t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
	}

	return &t, true
}

// parseWeekday parses weekday names and returns next occurrence
func parseWeekday(input string, now time.Time) (*time.Time, bool) {
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"sun":       time.Sunday,
		"monday":    time.Monday,
		"mon":       time.Monday,
		"tuesday":   time.Tuesday,
		"tue":       time.Tuesday,
		"tues":      time.Tuesday,
		"wednesday": time.Wednesday,
		"wed":       time.Wednesday,
		"thursday":  time.Thursday,
		"thu":       time.Thursday,
		"thur":      time.Thursday,
		"thurs":     time.Thursday,
		"friday":    time.Friday,
		"fri":       time.Friday,
		"saturday":  time.Saturday,
		"sat":       time.Saturday,
	}

	weekday, ok := weekdays[strings.ToLower(input)]
	if !ok {
		return nil, false
	}

	// Calculate days until next occurrence
	daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7 // Next week if today is that weekday
	}

	t := now.AddDate(0, 0, daysUntil)
	t = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())

	return &t, true
}
