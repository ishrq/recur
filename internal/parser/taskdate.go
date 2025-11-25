package parser

import (
	"time"
)

const (
	DefaultTimeHour   = 12
	DefaultTimeMinute = 0
)

// ParseTaskDate parses a date string for task creation/modification
// Returns an exact time suitable for setting as a task's due date
// For imprecise dates (like "November"), returns the first day with default time
func ParseTaskDate(input string) (*time.Time, error) {
	parsed, err := ParseDateTime(input)
	if err != nil {
		return nil, err
	}

	// Get exact time with default hour/minute for imprecise dates
	t := parsed.ExactTime(DefaultTimeHour, DefaultTimeMinute)
	return &t, nil
}

// ParseTaskDateWithTime parses a date string with custom default time
func ParseTaskDateWithTime(input string, defaultHour, defaultMinute int) (*time.Time, error) {
	parsed, err := ParseDateTime(input)
	if err != nil {
		return nil, err
	}

	t := parsed.ExactTime(defaultHour, defaultMinute)
	return &t, nil
}

// IsDefaultTime checks if a time has the default time (12:00:00)
func IsDefaultTime(t time.Time) bool {
	return t.Hour() == DefaultTimeHour && t.Minute() == DefaultTimeMinute && t.Second() == 0
}
