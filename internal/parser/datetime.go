package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Can be configured by user (default: Monday)
var WeekStartDay = time.Monday

// DatePrecision indicates how specific the parsed date is
type DatePrecision int

const (
	PrecisionExact   DatePrecision = iota // Full date/time
	PrecisionDay                          // Specific day
	PrecisionWeek                         // Week
	PrecisionMonth                        // Month
	PrecisionQuarter                      // Quarter
	PrecisionYear                         // Year only
)

// ParsedDate holds the result of parsing a date string
type ParsedDate struct {
	Time       time.Time     // The parsed time (start of range for imprecise dates)
	Precision  DatePrecision // How precise the input was
	Year       int           // Year component
	Month      int           // Month component (1-12, 0 if not specified)
	Day        int           // Day component (0 if not specified)
	Quarter    int           // Quarter (1-4, 0 if not specified)
	WeekNumber int           // ISO week number (0 if not specified)
}

// RangeStart returns the start of the date range
func (p ParsedDate) RangeStart() time.Time {
	return p.Time
}

// RangeEnd returns the end of the date range based on precision
func (p ParsedDate) RangeEnd() time.Time {
	switch p.Precision {
	case PrecisionYear:
		return time.Date(p.Year, 12, 31, 23, 59, 59, 0, p.Time.Location())
	case PrecisionQuarter:
		endMonth := p.Quarter * 3
		return endOfMonth(p.Year, time.Month(endMonth), p.Time.Location())
	case PrecisionMonth:
		return endOfMonth(p.Year, time.Month(p.Month), p.Time.Location())
	case PrecisionWeek:
		// End of week (6 days after start)
		return time.Date(p.Time.Year(), p.Time.Month(), p.Time.Day(), 23, 59, 59, 0, p.Time.Location()).AddDate(0, 0, 6)
	case PrecisionDay:
		return time.Date(p.Time.Year(), p.Time.Month(), p.Time.Day(), 23, 59, 59, 0, p.Time.Location())
	default:
		return p.Time
	}
}

// ExactTime returns the time to use when setting a task date
// For imprecise dates, returns a sensible default (start + default time)
func (p ParsedDate) ExactTime(defaultHour, defaultMinute int) time.Time {
	switch p.Precision {
	case PrecisionExact:
		return p.Time
	default:
		// For imprecise dates, use start date with default time
		return time.Date(p.Time.Year(), p.Time.Month(), p.Time.Day(), defaultHour, defaultMinute, 0, 0, p.Time.Location())
	}
}

// ParseDateTime parses a date/time string into a ParsedDate using the current time.
func ParseDateTime(input string) (*ParsedDate, error) {
	return ParseDateTimeWithNow(input, time.Now())
}

// ParseDateTimeWithNow is like ParseDateTime but uses a caller-supplied "now" for deterministic parsing.
func ParseDateTimeWithNow(input string, now time.Time) (*ParsedDate, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty date string")
	}

	loc := now.Location()

	// Try each parser in order of specificity
	if pd, ok := parseSemanticDate(input, now); ok {
		return pd, nil
	}

	if pd, ok := parseRelativeDate(input, now); ok {
		return pd, nil
	}

	if pd, ok := parseYearOnly(input, loc); ok {
		return pd, nil
	}

	if pd, ok := parseQuarter(input, now, loc); ok {
		return pd, nil
	}

	if pd, ok := parseYearMonth(input, now, loc); ok {
		return pd, nil
	}

	if pd, ok := parseMonthOnly(input, now, loc); ok {
		return pd, nil
	}

	if pd, ok := parseWeekNumber(input, now, loc); ok {
		return pd, nil
	}

	if pd, ok := parseWeekday(input, now); ok {
		return pd, nil
	}

	if pd, ok := parseStandardFormats(input, now, loc); ok {
		return pd, nil
	}

	return nil, fmt.Errorf("unrecognized date format: %s", input)
}

// parseSemanticDate handles: now, today, tomorrow, yesterday, this week, next week, etc.
func parseSemanticDate(input string, now time.Time) (*ParsedDate, bool) {
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch strings.ToLower(input) {
	case "now":
		return &ParsedDate{Time: now, Precision: PrecisionExact, Year: now.Year(), Month: int(now.Month()), Day: now.Day()}, true
	case "today":
		return &ParsedDate{Time: today, Precision: PrecisionDay, Year: today.Year(), Month: int(today.Month()), Day: today.Day()}, true
	case "tomorrow", "tmr":
		t := today.AddDate(0, 0, 1)
		return &ParsedDate{Time: t, Precision: PrecisionDay, Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, true
	case "yesterday":
		t := today.AddDate(0, 0, -1)
		return &ParsedDate{Time: t, Precision: PrecisionDay, Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, true
	case "this week":
		start := startOfWeek(today, WeekStartDay)
		_, week := start.ISOWeek()
		return &ParsedDate{Time: start, Precision: PrecisionWeek, Year: start.Year(), WeekNumber: week}, true
	case "next week":
		start := startOfWeek(today, WeekStartDay).AddDate(0, 0, 7)
		_, week := start.ISOWeek()
		return &ParsedDate{Time: start, Precision: PrecisionWeek, Year: start.Year(), WeekNumber: week}, true
	case "last week":
		start := startOfWeek(today, WeekStartDay).AddDate(0, 0, -7)
		_, week := start.ISOWeek()
		return &ParsedDate{Time: start, Precision: PrecisionWeek, Year: start.Year(), WeekNumber: week}, true
	case "this month":
		t := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionMonth, Year: t.Year(), Month: int(t.Month())}, true
	case "next month":
		t := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
		return &ParsedDate{Time: t, Precision: PrecisionMonth, Year: t.Year(), Month: int(t.Month())}, true
	case "last month":
		t := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
		return &ParsedDate{Time: t, Precision: PrecisionMonth, Year: t.Year(), Month: int(t.Month())}, true
	case "this quarter":
		q := (int(now.Month())-1)/3 + 1
		startMonth := (q-1)*3 + 1
		t := time.Date(now.Year(), time.Month(startMonth), 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionQuarter, Year: now.Year(), Quarter: q}, true
	case "next quarter":
		q := (int(now.Month())-1)/3 + 1
		nextQ := q + 1
		year := now.Year()
		if nextQ > 4 {
			nextQ = 1
			year++
		}
		startMonth := (nextQ-1)*3 + 1
		t := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionQuarter, Year: year, Quarter: nextQ}, true
	case "this year":
		t := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionYear, Year: now.Year()}, true
	case "next year":
		t := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionYear, Year: now.Year() + 1}, true
	case "last year":
		t := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, loc)
		return &ParsedDate{Time: t, Precision: PrecisionYear, Year: now.Year() - 1}, true
	}

	return nil, false
}

// parseRelativeDate handles: +3d, -2w, +1m, +1y
func parseRelativeDate(input string, now time.Time) (*ParsedDate, bool) {
	re := regexp.MustCompile(`^([+-])(\d+)([hdwmy])$`)
	matches := re.FindStringSubmatch(strings.ToLower(input))

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
	precision := PrecisionDay

	switch unit {
	case "h":
		t = now.Add(time.Duration(num) * time.Hour)
		precision = PrecisionExact
	case "d":
		t = now.AddDate(0, 0, num)
	case "w":
		t = now.AddDate(0, 0, num*7)
		precision = PrecisionWeek
	case "m":
		t = now.AddDate(0, num, 0)
	case "y":
		t = now.AddDate(num, 0, 0)
	default:
		return nil, false
	}

	// Normalize to start of day for non-hour units
	if unit != "h" {
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	return &ParsedDate{Time: t, Precision: precision, Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, true
}

// parseYearOnly handles: 2025
func parseYearOnly(input string, loc *time.Location) (*ParsedDate, bool) {
	re := regexp.MustCompile(`^(\d{4})$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 2 {
		return nil, false
	}

	year, _ := strconv.Atoi(matches[1])
	t := time.Date(year, 1, 1, 0, 0, 0, 0, loc)

	return &ParsedDate{Time: t, Precision: PrecisionYear, Year: year}, true
}

// parseQuarter handles: Q1, Q2, 2025-Q1, 2025Q1
func parseQuarter(input string, now time.Time, loc *time.Location) (*ParsedDate, bool) {
	re := regexp.MustCompile(`^(?:(\d{4})[-]?)?[Qq]([1-4])$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		return nil, false
	}

	year := now.Year()
	if matches[1] != "" {
		year, _ = strconv.Atoi(matches[1])
	}

	quarter, _ := strconv.Atoi(matches[2])
	startMonth := (quarter-1)*3 + 1
	t := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)

	return &ParsedDate{Time: t, Precision: PrecisionQuarter, Year: year, Quarter: quarter}, true
}

// parseYearMonth handles: 2025-11, 2025/11
func parseYearMonth(input string, now time.Time, loc *time.Location) (*ParsedDate, bool) {
	re := regexp.MustCompile(`^(\d{4})[-/](\d{1,2})$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		return nil, false
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])

	if month < 1 || month > 12 {
		return nil, false
	}

	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)

	return &ParsedDate{Time: t, Precision: PrecisionMonth, Year: year, Month: month}, true
}

func parseMonthOnly(input string, now time.Time, loc *time.Location) (*ParsedDate, bool) {
	months := map[string]time.Month{
		"january": time.January, "jan": time.January,
		"february": time.February, "feb": time.February,
		"march": time.March, "mar": time.March,
		"april": time.April, "apr": time.April,
		"may":  time.May,
		"june": time.June, "jun": time.June,
		"july": time.July, "jul": time.July,
		"august": time.August, "aug": time.August,
		"september": time.September, "sep": time.September, "sept": time.September,
		"october": time.October, "oct": time.October,
		"november": time.November, "nov": time.November,
		"december": time.December, "dec": time.December,
	}

	month, ok := months[strings.ToLower(input)]
	if !ok {
		return nil, false
	}

	// Use current year, or next year if month has passed
	year := now.Year()
	if month < now.Month() {
		year++
	}

	t := time.Date(year, month, 1, 0, 0, 0, 0, loc)

	return &ParsedDate{Time: t, Precision: PrecisionMonth, Year: year, Month: int(month)}, true
}

// parseWeekNumber handles: W47, week:47, 2025-W47, 2025W47
func parseWeekNumber(input string, now time.Time, loc *time.Location) (*ParsedDate, bool) {
	// ISO week format: 2025-W47, 2025W47
	re := regexp.MustCompile(`^(?:(\d{4})[-]?)?[Ww](\d{1,2})$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		// Try week:N format
		re2 := regexp.MustCompile(`^week[:\s]?(\d{1,2})$`)
		matches2 := re2.FindStringSubmatch(strings.ToLower(input))
		if len(matches2) != 2 {
			return nil, false
		}
		matches = []string{"", "", matches2[1]}
	}

	year := now.Year()
	if matches[1] != "" {
		year, _ = strconv.Atoi(matches[1])
	}

	week, _ := strconv.Atoi(matches[2])
	if week < 1 || week > 53 {
		return nil, false
	}

	// Calculate start of ISO week
	t := isoWeekStart(year, week, loc)

	return &ParsedDate{Time: t, Precision: PrecisionWeek, Year: year, WeekNumber: week}, true
}

func parseWeekday(input string, now time.Time) (*ParsedDate, bool) {
	weekdays := map[string]time.Weekday{
		"sunday": time.Sunday, "sun": time.Sunday,
		"monday": time.Monday, "mon": time.Monday,
		"tuesday": time.Tuesday, "tue": time.Tuesday, "tues": time.Tuesday,
		"wednesday": time.Wednesday, "wed": time.Wednesday,
		"thursday": time.Thursday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
		"friday": time.Friday, "fri": time.Friday,
		"saturday": time.Saturday, "sat": time.Saturday,
	}

	weekday, ok := weekdays[strings.ToLower(input)]
	if !ok {
		return nil, false
	}

	daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7

	t := now.AddDate(0, 0, daysUntil)
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return &ParsedDate{Time: t, Precision: PrecisionDay, Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, true
}

func parseStandardFormats(input string, now time.Time, loc *time.Location) (*ParsedDate, bool) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 3:04pm",
		"2006-01-02 3:04 pm",
		"2006-01-02 3:04PM",
		"2006-01-02 3:04 PM",
		"2006-01-02 3pm",
		"2006-01-02 3 pm",
		"2006-01-02",
		"2006/01/02 15:04",
		"2006/01/02",
		"Jan 2 15:04",
		"Jan 2 3:04pm",
		"Jan 2 3pm",
		"Jan 2 2006",
		"Jan 2",
		"January 2 15:04",
		"January 2 2006",
		"January 2",
		"02 Jan 2006",
		"02-01-2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// Fix year if not specified
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
			}

			precision := PrecisionDay
			if t.Hour() != 0 || t.Minute() != 0 || strings.Contains(input, ":") ||
				strings.Contains(strings.ToLower(input), "am") ||
				strings.Contains(strings.ToLower(input), "pm") {
				precision = PrecisionExact
			}

			return &ParsedDate{Time: t, Precision: precision, Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, true
		}
	}

	return nil, false
}

func startOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	daysBack := int(t.Weekday()) - int(weekStart)
	if daysBack < 0 {
		daysBack += 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-daysBack, 0, 0, 0, 0, t.Location())
}

func endOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	return nextMonth.Add(-time.Second)
}

func isoWeekStart(year, week int, loc *time.Location) time.Time {
	// Find Jan 4 of the year (always in week 1)
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, loc)

	// Find Monday of that week
	daysBack := int(jan4.Weekday()) - int(time.Monday)
	if daysBack < 0 {
		daysBack += 7
	}
	week1Start := jan4.AddDate(0, 0, -daysBack)

	// Add weeks
	return week1Start.AddDate(0, 0, (week-1)*7)
}
