package parser

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 6, 20, 14, 30, 0, 0, time.UTC) // Saturday

func parseOrFail(t *testing.T, input string) *ParsedDate {
	t.Helper()
	pd, err := ParseDateTimeWithNow(input, fixedNow)
	if err != nil {
		t.Fatalf("ParseDateTimeWithNow(%q) fixedNow=%v: %v", input, fixedNow, err)
	}
	return pd
}

func expectError(t *testing.T, input string) {
	t.Helper()
	_, err := ParseDateTimeWithNow(input, fixedNow)
	if err == nil {
		t.Fatalf("ParseDateTime(%q) expected error, got nil", input)
	}
}

func isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// --- Semantic dates ---

func TestParseSemantic_Now(t *testing.T) {
	pd := parseOrFail(t, "now")
	if pd.Precision != PrecisionExact {
		t.Errorf("now: expected PrecisionExact, got %v", pd.Precision)
	}
	if !pd.Time.Equal(fixedNow) {
		t.Errorf("now: expected %v, got %v", fixedNow, pd.Time)
	}
}

func TestParseSemantic_Today(t *testing.T) {
	pd := parseOrFail(t, "today")
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())

	if pd.Precision != PrecisionDay {
		t.Errorf("today: expected PrecisionDay, got %v", pd.Precision)
	}
	if !pd.Time.Equal(fixedToday) {
		t.Errorf("today: expected %v, got %v", fixedToday, pd.Time)
	}
	if pd.Time.Hour() != 0 || pd.Time.Minute() != 0 {
		t.Errorf("today: expected midnight, got %v", pd.Time)
	}
}

func TestParseSemantic_Tomorrow(t *testing.T) {
	pd := parseOrFail(t, "tomorrow")
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	expected := fixedToday.AddDate(0, 0, 1)

	if pd.Precision != PrecisionDay {
		t.Errorf("tomorrow: expected PrecisionDay, got %v", pd.Precision)
	}
	if !pd.Time.Equal(expected) {
		t.Errorf("tomorrow: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_Tmr(t *testing.T) {
	pd := parseOrFail(t, "tmr")
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	expected := fixedToday.AddDate(0, 0, 1)
	if !pd.Time.Equal(expected) {
		t.Errorf("tmr: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_Yesterday(t *testing.T) {
	pd := parseOrFail(t, "yesterday")
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	expected := fixedToday.AddDate(0, 0, -1)

	if pd.Precision != PrecisionDay {
		t.Errorf("yesterday: expected PrecisionDay, got %v", pd.Precision)
	}
	if !pd.Time.Equal(expected) {
		t.Errorf("yesterday: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_ThisWeek(t *testing.T) {
	pd := parseOrFail(t, "this week")
	if pd.Precision != PrecisionWeek {
		t.Errorf("this week: expected PrecisionWeek, got %v", pd.Precision)
	}
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	start := startOfWeek(fixedToday, WeekStartDay)
	if !pd.Time.Equal(start) {
		t.Errorf("this week: expected %v, got %v", start, pd.Time)
	}
}

func TestParseSemantic_NextWeek(t *testing.T) {
	pd := parseOrFail(t, "next week")
	if pd.Precision != PrecisionWeek {
		t.Errorf("next week: expected PrecisionWeek, got %v", pd.Precision)
	}
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	start := startOfWeek(fixedToday, WeekStartDay).AddDate(0, 0, 7)
	if !pd.Time.Equal(start) {
		t.Errorf("next week: expected %v, got %v", start, pd.Time)
	}
}

func TestParseSemantic_LastWeek(t *testing.T) {
	pd := parseOrFail(t, "last week")
	if pd.Precision != PrecisionWeek {
		t.Errorf("last week: expected PrecisionWeek, got %v", pd.Precision)
	}
	fixedToday := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	start := startOfWeek(fixedToday, WeekStartDay).AddDate(0, 0, -7)
	if !pd.Time.Equal(start) {
		t.Errorf("last week: expected %v, got %v", start, pd.Time)
	}
}

func TestParseSemantic_ThisMonth(t *testing.T) {
	pd := parseOrFail(t, "this month")
	if pd.Precision != PrecisionMonth {
		t.Errorf("this month: expected PrecisionMonth, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year(), fixedNow.Month(), 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("this month: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_NextMonth(t *testing.T) {
	pd := parseOrFail(t, "next month")
	if pd.Precision != PrecisionMonth {
		t.Errorf("next month: expected PrecisionMonth, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year(), fixedNow.Month(), 1, 0, 0, 0, 0, fixedNow.Location()).AddDate(0, 1, 0)
	if !pd.Time.Equal(expected) {
		t.Errorf("next month: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_LastMonth(t *testing.T) {
	pd := parseOrFail(t, "last month")
	if pd.Precision != PrecisionMonth {
		t.Errorf("last month: expected PrecisionMonth, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year(), fixedNow.Month(), 1, 0, 0, 0, 0, fixedNow.Location()).AddDate(0, -1, 0)
	if !pd.Time.Equal(expected) {
		t.Errorf("last month: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_ThisQuarter(t *testing.T) {
	pd := parseOrFail(t, "this quarter")
	if pd.Precision != PrecisionQuarter {
		t.Errorf("this quarter: expected PrecisionQuarter, got %v", pd.Precision)
	}
	q := (int(fixedNow.Month())-1)/3 + 1
	startMonth := (q-1)*3 + 1
	expected := time.Date(fixedNow.Year(), time.Month(startMonth), 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("this quarter: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_NextQuarter(t *testing.T) {
	pd := parseOrFail(t, "next quarter")
	if pd.Precision != PrecisionQuarter {
		t.Errorf("next quarter: expected PrecisionQuarter, got %v", pd.Precision)
	}
	q := (int(fixedNow.Month())-1)/3 + 1
	nextQ := q + 1
	year := fixedNow.Year()
	if nextQ > 4 {
		nextQ = 1
		year++
	}
	startMonth := (nextQ-1)*3 + 1
	expected := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("next quarter: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_ThisYear(t *testing.T) {
	pd := parseOrFail(t, "this year")
	if pd.Precision != PrecisionYear {
		t.Errorf("this year: expected PrecisionYear, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year(), 1, 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("this year: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_NextYear(t *testing.T) {
	pd := parseOrFail(t, "next year")
	if pd.Precision != PrecisionYear {
		t.Errorf("next year: expected PrecisionYear, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year()+1, 1, 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("next year: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseSemantic_LastYear(t *testing.T) {
	pd := parseOrFail(t, "last year")
	if pd.Precision != PrecisionYear {
		t.Errorf("last year: expected PrecisionYear, got %v", pd.Precision)
	}
	expected := time.Date(fixedNow.Year()-1, 1, 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("last year: expected %v, got %v", expected, pd.Time)
	}
}

// --- Relative dates ---

func TestParseRelative_PlusDays(t *testing.T) {
	fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	pd := parseOrFail(t, "+3d")
	if pd.Precision != PrecisionDay {
		t.Errorf("+3d: expected PrecisionDay, got %v", pd.Precision)
	}
	expected := fixedMidnight.AddDate(0, 0, 3)
	if !pd.Time.Equal(expected) {
		t.Errorf("+3d: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_MinusWeeks(t *testing.T) {
	fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	pd := parseOrFail(t, "-2w")
	if pd.Precision != PrecisionWeek {
		t.Errorf("-2w: expected PrecisionWeek, got %v", pd.Precision)
	}
	expected := fixedMidnight.AddDate(0, 0, -14)
	if !pd.Time.Equal(expected) {
		t.Errorf("-2w: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_PlusMonths(t *testing.T) {
	fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	pd := parseOrFail(t, "+1m")
	if pd.Precision != PrecisionDay {
		t.Errorf("+1m: expected PrecisionDay, got %v", pd.Precision)
	}
	expected := fixedMidnight.AddDate(0, 1, 0)
	if !pd.Time.Equal(expected) {
		t.Errorf("+1m: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_PlusYears(t *testing.T) {
	fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	pd := parseOrFail(t, "+1y")
	if pd.Precision != PrecisionDay {
		t.Errorf("+1y: expected PrecisionDay, got %v", pd.Precision)
	}
	expected := fixedMidnight.AddDate(1, 0, 0)
	if !pd.Time.Equal(expected) {
		t.Errorf("+1y: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_PlusHours(t *testing.T) {
	pd := parseOrFail(t, "+6h")
	if pd.Precision != PrecisionExact {
		t.Errorf("+6h: expected PrecisionExact, got %v", pd.Precision)
	}
	expected := fixedNow.Add(6 * time.Hour)
	if !pd.Time.Equal(expected) {
		t.Errorf("+6h: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_NegativeDays(t *testing.T) {
	fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
	pd := parseOrFail(t, "-1d")
	expected := fixedMidnight.AddDate(0, 0, -1)
	if !pd.Time.Equal(expected) {
		t.Errorf("-1d: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseRelative_Invalid(t *testing.T) {
	expectError(t, "abc")
	expectError(t, "++3d")
	expectError(t, "+")
	expectError(t, "3d")
}

// --- Year-only ---

func TestParseYearOnly_CurrentCentury(t *testing.T) {
	pd := parseOrFail(t, "2025")
	if pd.Precision != PrecisionYear {
		t.Errorf("2025: expected PrecisionYear, got %v", pd.Precision)
	}
	if pd.Year != 2025 {
		t.Errorf("2025: expected year 2025, got %d", pd.Year)
	}
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("2025: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseYearOnly_OtherCenturies(t *testing.T) {
	tests := []int{1999, 2100, 1984, 2099, 2000}
	for _, year := range tests {
		pd := parseOrFail(t, formatYear(year))
		if pd.Year != year {
			t.Errorf("expected year %d, got %d", year, pd.Year)
		}
	}
}

func formatYear(y int) string {
	return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func TestParseYearOnly_Invalid(t *testing.T) {
	expectError(t, "123")
	expectError(t, "20xx")
	expectError(t, "abc")
}

// --- Quarter ---

func TestParseQuarter_Simple(t *testing.T) {
	tests := []struct {
		input   string
		quarter int
	}{
		{"Q1", 1}, {"Q2", 2}, {"Q3", 3}, {"Q4", 4},
		{"q1", 1}, {"q2", 2}, {"q3", 3}, {"q4", 4},
	}
	for _, tt := range tests {
		pd := parseOrFail(t, tt.input)
		if pd.Precision != PrecisionQuarter {
			t.Errorf("%s: expected PrecisionQuarter, got %v", tt.input, pd.Precision)
		}
		if pd.Quarter != tt.quarter {
			t.Errorf("%s: expected quarter %d, got %d", tt.input, tt.quarter, pd.Quarter)
		}
		startMonth := (tt.quarter-1)*3 + 1
		expected := time.Date(fixedNow.Year(), time.Month(startMonth), 1, 0, 0, 0, 0, fixedNow.Location())
		if !pd.Time.Equal(expected) {
			t.Errorf("%s: expected %v, got %v", tt.input, expected, pd.Time)
		}
	}
}

func TestParseQuarter_WithYear(t *testing.T) {
	tests := []struct {
		input   string
		year    int
		quarter int
	}{
		{"2025-Q1", 2025, 1},
		{"2025Q3", 2025, 3},
		{"1999-Q4", 1999, 4},
	}
	for _, tt := range tests {
		pd := parseOrFail(t, tt.input)
		if pd.Year != tt.year {
			t.Errorf("%s: expected year %d, got %d", tt.input, tt.year, pd.Year)
		}
		if pd.Quarter != tt.quarter {
			t.Errorf("%s: expected quarter %d, got %d", tt.input, tt.quarter, pd.Quarter)
		}
	}
}

func TestParseQuarter_Invalid(t *testing.T) {
	expectError(t, "Q5")
	expectError(t, "Q0")
	expectError(t, "quarter1")
}

// --- Year-Month ---

func TestParseYearMonth(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
	}{
		{"2025-11", 2025, 11},
		{"2025/11", 2025, 11},
		{"2025-03", 2025, 3},
		{"2025/3", 2025, 3},
	}
	for _, tt := range tests {
		pd := parseOrFail(t, tt.input)
		if pd.Precision != PrecisionMonth {
			t.Errorf("%s: expected PrecisionMonth, got %v", tt.input, pd.Precision)
		}
		if pd.Year != tt.year {
			t.Errorf("%s: expected year %d, got %d", tt.input, tt.year, pd.Year)
		}
		if pd.Month != tt.month {
			t.Errorf("%s: expected month %d, got %d", tt.input, tt.month, pd.Month)
		}
	}
}

func TestParseYearMonth_Invalid(t *testing.T) {
	expectError(t, "2025-13")
	expectError(t, "2025-00")
	expectError(t, "20-11")
	expectError(t, "2025-1-1")
}

// --- Month-only ---

func TestParseMonthOnly(t *testing.T) {
	tests := []string{
		"january", "jan",
		"february", "feb",
		"march", "mar",
		"april", "apr",
		"may",
		"june", "jun",
		"july", "jul",
		"august", "aug",
		"september", "sep", "sept",
		"october", "oct",
		"november", "nov",
		"december", "dec",
	}
	for _, input := range tests {
		pd := parseOrFail(t, input)
		if pd.Precision != PrecisionMonth {
			t.Errorf("%s: expected PrecisionMonth, got %v", input, pd.Precision)
		}
		expectedYear := fixedNow.Year()
		if pd.Time.Month() < fixedNow.Month() {
			expectedYear++
		}
		expected := time.Date(expectedYear, pd.Time.Month(), 1, 0, 0, 0, 0, fixedNow.Location())
		if !pd.Time.Equal(expected) {
			t.Errorf("%s: expected %v, got %v", input, expected, pd.Time)
		}
	}
}

func TestParseMonthOnly_Invalid(t *testing.T) {
	expectError(t, "xyz")
	expectError(t, "summer")
}

// --- Week number ---

func TestParseWeekNumber_Basic(t *testing.T) {
	pd := parseOrFail(t, "W47")
	if pd.Precision != PrecisionWeek {
		t.Errorf("W47: expected PrecisionWeek, got %v", pd.Precision)
	}
	if pd.WeekNumber != 47 {
		t.Errorf("W47: expected week 47, got %d", pd.WeekNumber)
	}
	expected := isoWeekStart(fixedNow.Year(), 47, fixedNow.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("W47: expected %v, got %v", expected, pd.Time)
	}
}

func TestParseWeekNumber_WeekKeyword(t *testing.T) {
	pd := parseOrFail(t, "week:3")
	if pd.WeekNumber != 3 {
		t.Errorf("week:3: expected week 3, got %d", pd.WeekNumber)
	}
}

func TestParseWeekNumber_WithYear(t *testing.T) {
	pd := parseOrFail(t, "2025-W03")
	if pd.Year != 2025 {
		t.Errorf("2025-W03: expected year 2025, got %d", pd.Year)
	}
	if pd.WeekNumber != 3 {
		t.Errorf("2025-W03: expected week 3, got %d", pd.WeekNumber)
	}
}

func TestParseWeekNumber_Invalid(t *testing.T) {
	expectError(t, "W99")
	expectError(t, "W0")
	expectError(t, "week:99")
	expectError(t, "week")
}

// --- Weekday ---

func TestParseWeekday(t *testing.T) {
	weekdays := []string{
		"sunday", "sun",
		"monday", "mon",
		"tuesday", "tue", "tues",
		"wednesday", "wed",
		"thursday", "thu", "thur", "thurs",
		"friday", "fri",
		"saturday", "sat",
	}
	for _, input := range weekdays {
		pd := parseOrFail(t, input)
		if pd.Precision != PrecisionDay {
			t.Errorf("%s: expected PrecisionDay, got %v", input, pd.Precision)
		}
		fixedMidnight := time.Date(fixedNow.Year(), fixedNow.Month(), fixedNow.Day(), 0, 0, 0, 0, fixedNow.Location())
		expectedDay := pd.Time.Weekday()
		if !pd.Time.Equal(fixedMidnight) && !pd.Time.After(fixedMidnight) {
			t.Errorf("%s: got %v which is before today %v", input, pd.Time, fixedMidnight)
		}
		if fixedMidnight.Weekday() == expectedDay {
			if !pd.Time.Equal(fixedMidnight) {
				t.Errorf("%s: today is %s, expected today, got %v", input, fixedMidnight.Weekday(), pd.Time)
			}
		}
	}
}

func TestParseWeekday_Invalid(t *testing.T) {
	expectError(t, "xyz")
	expectError(t, "monnday")
	expectError(t, "tuesda")
}

// --- Standard formats ---

func TestParseStandard_ISODateTime(t *testing.T) {
	pd := parseOrFail(t, "2025-11-15 15:04:05")
	if pd.Precision != PrecisionExact {
		t.Errorf("expected PrecisionExact, got %v", pd.Precision)
	}
	if pd.Year != 2025 || pd.Month != 11 || pd.Day != 15 {
		t.Errorf("expected 2025-11-15, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
	expected := time.Date(2025, 11, 15, 15, 4, 5, 0, pd.Time.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, pd.Time)
	}
}

func TestParseStandard_ISODateOnly(t *testing.T) {
	pd := parseOrFail(t, "2025-11-15")
	if pd.Precision != PrecisionDay {
		t.Errorf("expected PrecisionDay, got %v", pd.Precision)
	}
	expected := time.Date(2025, 11, 15, 0, 0, 0, 0, pd.Time.Location())
	if !pd.Time.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, pd.Time)
	}
}

func TestParseStandard_SlashDate(t *testing.T) {
	pd := parseOrFail(t, "2025/11/20")
	if pd.Year != 2025 || pd.Month != 11 || pd.Day != 20 {
		t.Errorf("expected 2025-11-20, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
}

func TestParseStandard_MonthNameDay(t *testing.T) {
	pd := parseOrFail(t, "Jan 2 2006")
	if pd.Year != 2006 || pd.Month != 1 || pd.Day != 2 {
		t.Errorf("expected 2006-01-02, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
}

func TestParseStandard_MonthNameDayTime(t *testing.T) {
	pd := parseOrFail(t, "Jan 2 15:04")
	if pd.Precision != PrecisionExact {
		t.Errorf("expected PrecisionExact, got %v", pd.Precision)
	}
	if pd.Month != 1 || pd.Day != 2 {
		t.Errorf("expected Jan 2, got month=%d day=%d", pd.Month, pd.Day)
	}
}

func TestParseStandard_EuropeanDate(t *testing.T) {
	pd := parseOrFail(t, "20 Nov 2025")
	if pd.Year != 2025 || pd.Month != 11 || pd.Day != 20 {
		t.Errorf("expected 2025-11-20, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
}

func TestParseStandard_EuropeanNumeric(t *testing.T) {
	pd := parseOrFail(t, "20-11-2025")
	if pd.Year != 2025 || pd.Month != 11 || pd.Day != 20 {
		t.Errorf("expected 2025-11-20, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
}

func TestParseStandard_USFormat(t *testing.T) {
	pd := parseOrFail(t, "01/02/2006")
	if pd.Month != 1 || pd.Day != 2 || pd.Year != 2006 {
		t.Errorf("expected 2006-01-02, got %d-%d-%d", pd.Year, pd.Month, pd.Day)
	}
}

func TestParseStandard_12HourTime(t *testing.T) {
	tests := []string{
		"2025-11-15 3:04pm",
		"2025-11-15 3:04 pm",
		"2025-11-15 3:04PM",
		"2025-11-15 3:04 PM",
	}
	for _, input := range tests {
		pd := parseOrFail(t, input)
		if pd.Precision != PrecisionExact {
			t.Errorf("%s: expected PrecisionExact, got %v", input, pd.Precision)
		}
		if pd.Time.Hour() != 15 || pd.Time.Minute() != 4 {
			t.Errorf("%s: expected 15:04, got %d:%d", input, pd.Time.Hour(), pd.Time.Minute())
		}
	}
}

func TestParseStandard_HoursOnly(t *testing.T) {
	tests := []string{
		"2025-11-15 3pm",
		"2025-11-15 3 pm",
	}
	for _, input := range tests {
		pd := parseOrFail(t, input)
		if pd.Precision != PrecisionExact {
			t.Errorf("%s: expected PrecisionExact, got %v", input, pd.Precision)
		}
		if pd.Time.Hour() != 15 || pd.Time.Minute() != 0 {
			t.Errorf("%s: expected 15:00, got %d:%d", input, pd.Time.Hour(), pd.Time.Minute())
		}
	}
}

func TestParseStandard_MonthDayNoYear(t *testing.T) {
	pd := parseOrFail(t, "Jan 2")
	if pd.Month != 1 || pd.Day != 2 {
		t.Errorf("expected Jan 2, got month=%d day=%d", pd.Month, pd.Day)
	}
	if pd.Year != fixedNow.Year() {
		t.Errorf("expected current year %d, got %d", fixedNow.Year(), pd.Year)
	}
}

// --- Edge cases ---

func TestParse_EmptyString(t *testing.T) {
	expectError(t, "")
}

func TestParse_Whitespace(t *testing.T) {
	expectError(t, "  ")
}

func TestParse_Garbage(t *testing.T) {
	expectError(t, "not-a-date")
	expectError(t, "!!")
	expectError(t, "12345")
}

// --- ParseTaskDate ---

func TestParseTaskDate_ImpreciseGetsDefaultTime(t *testing.T) {
	dt, err := ParseTaskDate("2026-06-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.Hour() != DefaultTimeHour || dt.Minute() != DefaultTimeMinute {
		t.Errorf("expected %d:%d, got %d:%d", DefaultTimeHour, DefaultTimeMinute, dt.Hour(), dt.Minute())
	}
}

func TestParseTaskDate_ExactKeepsTime(t *testing.T) {
	dt, err := ParseTaskDate("2025-11-15 15:04")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.Hour() != 15 || dt.Minute() != 4 {
		t.Errorf("expected 15:04, got %d:%d", dt.Hour(), dt.Minute())
	}
}

func TestParseTaskDate_Invalid(t *testing.T) {
	_, err := ParseTaskDate("not-a-date")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTaskDateWithTime(t *testing.T) {
	dt, err := ParseTaskDateWithTime("2026-06-20", 8, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.Hour() != 8 || dt.Minute() != 30 {
		t.Errorf("expected 8:30, got %d:%d", dt.Hour(), dt.Minute())
	}
}

// --- ValidateFrequency ---

func TestValidateFrequency_Numeric(t *testing.T) {
	valid := []string{"1h", "12h", "1d", "7d", "1w", "4w", "1m", "6m", "1y", "2y"}
	for _, f := range valid {
		if err := ValidateFrequency(f); err != nil {
			t.Errorf("expected %q to be valid, got: %v", f, err)
		}
	}
}

func TestValidateFrequency_Semantic(t *testing.T) {
	valid := []string{"hourly", "daily", "weekly", "monthly", "yearly"}
	for _, f := range valid {
		if err := ValidateFrequency(f); err != nil {
			t.Errorf("expected %q to be valid, got: %v", f, err)
		}
	}
}

func TestValidateFrequency_Invalid(t *testing.T) {
	invalid := []string{"abc", "1x", "0d", "-1d", "", "1.5d"}
	for _, f := range invalid {
		if err := ValidateFrequency(f); err == nil {
			t.Errorf("expected %q to be invalid", f)
		}
	}
}

// --- CalculateNextOccurrence ---

func TestCalcNext_Semantic(t *testing.T) {
	base := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		freq   string
		expect time.Time
	}{
		{"hourly", base.Add(time.Hour)},
		{"daily", base.AddDate(0, 0, 1)},
		{"weekly", base.AddDate(0, 0, 7)},
		{"monthly", base.AddDate(0, 1, 0)},
		{"yearly", base.AddDate(1, 0, 0)},
	}
	for _, tt := range tests {
		got, err := CalculateNextOccurrence(base, tt.freq)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.freq, err)
			continue
		}
		if !got.Equal(tt.expect) {
			t.Errorf("%s: expected %v, got %v", tt.freq, tt.expect, got)
		}
	}
}

func TestCalcNext_Numeric(t *testing.T) {
	base := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		freq   string
		expect time.Time
	}{
		{"2h", base.Add(2 * time.Hour)},
		{"3d", base.AddDate(0, 0, 3)},
		{"2w", base.AddDate(0, 0, 14)},
		{"1m", base.AddDate(0, 1, 0)}, // Jan 31 + 1 month = Feb 28 (calendar-aware)
		{"2y", base.AddDate(2, 0, 0)},
	}
	for _, tt := range tests {
		got, err := CalculateNextOccurrence(base, tt.freq)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.freq, err)
			continue
		}
		if !got.Equal(tt.expect) {
			t.Errorf("%s: expected %v, got %v", tt.freq, tt.expect, got)
		}
	}
}

func TestCalcNext_InvalidFrequency(t *testing.T) {
	base := time.Now()
	_, err := CalculateNextOccurrence(base, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid frequency")
	}
}

// --- RangeEnd ---

func TestRangeEnd_Exact(t *testing.T) {
	pd := ParsedDate{Time: time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC), Precision: PrecisionExact}
	end := pd.RangeEnd()
	if !end.Equal(pd.Time) {
		t.Errorf("Exact: expected %v, got %v", pd.Time, end)
	}
}

func TestRangeEnd_Day(t *testing.T) {
	pd := ParsedDate{Time: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), Precision: PrecisionDay}
	end := pd.RangeEnd()
	expected := time.Date(2025, 6, 15, 23, 59, 59, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("Day: expected %v, got %v", expected, end)
	}
}

func TestRangeEnd_Week(t *testing.T) {
	start := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC) // Sunday
	pd := ParsedDate{Time: start, Precision: PrecisionWeek, Year: 2025, WeekNumber: 24}
	end := pd.RangeEnd()
	expected := time.Date(2025, 6, 21, 23, 59, 59, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("Week: expected %v, got %v", expected, end)
	}
}

func TestRangeEnd_Month(t *testing.T) {
	pd := ParsedDate{Time: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), Precision: PrecisionMonth, Year: 2025, Month: 6}
	end := pd.RangeEnd()
	expected := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("Month: expected %v, got %v", expected, end)
	}
}

func TestRangeEnd_Quarter(t *testing.T) {
	pd := ParsedDate{Time: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), Precision: PrecisionQuarter, Year: 2025, Quarter: 2}
	end := pd.RangeEnd()
	expected := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("Quarter: expected %v, got %v", expected, end)
	}
}

func TestRangeEnd_Year(t *testing.T) {
	pd := ParsedDate{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Precision: PrecisionYear, Year: 2025}
	end := pd.RangeEnd()
	expected := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("Year: expected %v, got %v", expected, end)
	}
}

// --- ParseTaskString error propagation ---

func TestParseTaskString_InvalidDateReturnsError(t *testing.T) {
	_, err := ParseTaskString("Task @(not-a-date)")
	if err == nil {
		t.Fatal("expected error for invalid date string")
	}
}

func TestParseTaskString_EmptyParens(t *testing.T) {
	// @() doesn't match the regex (requires content inside parens)
	// so it's left in the task name
	task, err := ParseTaskString("Task @()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.DueDate != nil {
		t.Fatal("expected no due date for empty @()")
	}
	if task.Name != "Task @()" {
		t.Errorf("expected name 'Task @()', got %q", task.Name)
	}
}

func TestParseTaskString_FrequencyWithoutDate(t *testing.T) {
	_, err := ParseTaskString("Task @(, 1w)")
	if err == nil {
		t.Fatal("expected error for frequency without date")
	}
}

func TestParseTaskString_RecurrenceWithoutDate(t *testing.T) {
	_, err := ParseTaskString("Task @(invalid, 1w, 2025-12-31)")
	if err == nil {
		t.Fatal("expected error for recurrence without valid date")
	}
}

// --- SetDefaultTaskTime ---

func TestSetDefaultTaskTime_Valid(t *testing.T) {
	savedH, savedM := DefaultTimeHour, DefaultTimeMinute
	defer func() { DefaultTimeHour, DefaultTimeMinute = savedH, savedM }()

	if err := SetDefaultTaskTime("08:30"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if DefaultTimeHour != 8 || DefaultTimeMinute != 30 {
		t.Errorf("expected 8:30, got %d:%d", DefaultTimeHour, DefaultTimeMinute)
	}
}

func TestSetDefaultTaskTime_Invalid(t *testing.T) {
	savedH, savedM := DefaultTimeHour, DefaultTimeMinute
	defer func() { DefaultTimeHour, DefaultTimeMinute = savedH, savedM }()

	invalid := []string{"abc", "25:00", "12:60", "12", "12:00:00"}
	for _, s := range invalid {
		if err := SetDefaultTaskTime(s); err == nil {
			t.Errorf("expected error for %q", s)
		}
	}
}

func TestSetDefaultTaskTime_AffectsParseTaskDate(t *testing.T) {
	savedH, savedM := DefaultTimeHour, DefaultTimeMinute
	defer func() { DefaultTimeHour, DefaultTimeMinute = savedH, savedM }()

	if err := SetDefaultTaskTime("08:30"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dt, err := ParseTaskDate("today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.Hour() != 8 || dt.Minute() != 30 {
		t.Errorf("expected 8:30, got %d:%d", dt.Hour(), dt.Minute())
	}
}

// --- Multi-tag ---

func TestParseTaskString_MultipleTags(t *testing.T) {
	task, err := ParseTaskString("Task #work #urgent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Tag != "work,urgent" {
		t.Errorf("expected 'work,urgent', got %q", task.Tag)
	}
	if task.Name != "Task" {
		t.Errorf("expected 'Task', got %q", task.Name)
	}
}

func TestParseTaskString_SingleTag(t *testing.T) {
	task, err := ParseTaskString("Task #shopping")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Tag != "shopping" {
		t.Errorf("expected 'shopping', got %q", task.Tag)
	}
}

func TestParseTaskString_NoTags(t *testing.T) {
	task, err := ParseTaskString("Task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Tag != "" {
		t.Errorf("expected empty tag, got %q", task.Tag)
	}
}

func TestParseTaskString_TagWithOtherFields(t *testing.T) {
	task, err := ParseTaskString("Task #work #urgent +Project !high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Tag != "work,urgent" {
		t.Errorf("expected 'work,urgent', got %q", task.Tag)
	}
	if task.Project != "Project" {
		t.Errorf("expected 'Project', got %q", task.Project)
	}
	if task.Priority != "high" {
		t.Errorf("expected 'high', got %q", task.Priority)
	}
	if task.Name != "Task" {
		t.Errorf("expected 'Task', got %q", task.Name)
	}
}
