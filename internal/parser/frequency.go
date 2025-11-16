package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
		// Use AddDate for days (calendar-aware)
		return baseDate.AddDate(0, 0, num), nil
	case "w":
		// Use AddDate for weeks (calendar-aware)
		return baseDate.AddDate(0, 0, num*7), nil
	case "m":
		// Use AddDate for months (calendar-aware, handles variable month lengths)
		return baseDate.AddDate(0, num, 0), nil
	case "y":
		// Use AddDate for years (calendar-aware, handles leap years)
		return baseDate.AddDate(num, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown frequency unit: %s", unit)
	}
}
