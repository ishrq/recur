package filter

import (
	"time"

	"github.com/ishrq/recur/internal/parser"
)

// DateRange represents a start and end time for filtering
type DateRange struct {
	Start time.Time
	End   time.Time
}

// ParseDateFilter parses a date string and returns a date range for filtering
// For imprecise dates (like "November"), returns the full range
func ParseDateFilter(input string) (*DateRange, error) {
	parsed, err := parser.ParseDateTime(input)
	if err != nil {
		return nil, err
	}

	return &DateRange{
		Start: parsed.RangeStart(),
		End:   parsed.RangeEnd(),
	}, nil
}

// Contains checks if a given time falls within the date range
func (dr DateRange) Contains(t time.Time) bool {
	return !t.Before(dr.Start) && !t.After(dr.End)
}

// ContainsDate checks if a given date (ignoring time) falls within the date range
func (dr DateRange) ContainsDate(t time.Time) bool {
	// Normalize to start of day
	normalized := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	rangeStart := time.Date(dr.Start.Year(), dr.Start.Month(), dr.Start.Day(), 0, 0, 0, 0, dr.Start.Location())
	rangeEnd := time.Date(dr.End.Year(), dr.End.Month(), dr.End.Day(), 23, 59, 59, 0, dr.End.Location())

	return !normalized.Before(rangeStart) && !normalized.After(rangeEnd)
}

// ParseDate parses a date string for simple date filtering
// Returns the start of the date range (for backwards compatibility)
func ParseDate(dateStr string) (time.Time, error) {
	dateRange, err := ParseDateFilter(dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return dateRange.Start, nil
}
