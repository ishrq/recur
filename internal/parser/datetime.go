package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
