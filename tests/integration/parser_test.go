package integration

import (
	"testing"
	"time"

	"github.com/ishrq/recur/internal/parser"
)

func TestParser_SimpleName(t *testing.T) {
	task, err := parser.ParseTaskString("Buy groceries")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Name != "Buy groceries" {
		t.Errorf("Expected name 'Buy groceries', got %q", task.Name)
	}
}

func TestParser_WithTag(t *testing.T) {
	task, err := parser.ParseTaskString("Task name #tag")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Name != "Task name" {
		t.Errorf("Expected name 'Task name', got %q", task.Name)
	}

	if task.Tag != "tag" {
		t.Errorf("Expected tag 'tag', got %q", task.Tag)
	}
}

func TestParser_WithProject(t *testing.T) {
	task, err := parser.ParseTaskString("Task name +Project")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Project != "Project" {
		t.Errorf("Expected project 'Project', got %q", task.Project)
	}
}

func TestParser_WithPriority(t *testing.T) {
	task, err := parser.ParseTaskString("Task name !urgent")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Priority != "urgent" {
		t.Errorf("Expected priority 'urgent', got %q", task.Priority)
	}
}

func TestParser_WithNote(t *testing.T) {
	task, err := parser.ParseTaskString("Task name *'This is a note'")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Note != "This is a note" {
		t.Errorf("Expected note 'This is a note', got %q", task.Note)
	}
}

func TestParser_WithDueDate(t *testing.T) {
	task, err := parser.ParseTaskString("Task name @(2025-11-15 3pm)")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.DueDate == nil {
		t.Fatal("Expected due date to be set")
	}

	expected := time.Date(2025, 11, 15, 15, 0, 0, 0, task.DueDate.Location())
	if !task.DueDate.Equal(expected) {
		t.Errorf("Expected due date %v, got %v", expected, task.DueDate)
	}
}

func TestParser_WithRecurring(t *testing.T) {
	task, err := parser.ParseTaskString("Task name @(2025-11-15, 1w)")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.RecurFrequency != "1w" {
		t.Errorf("Expected frequency '1w', got %q", task.RecurFrequency)
	}
}

func TestParser_Complete(t *testing.T) {
	task, err := parser.ParseTaskString("Task @(2025-11-15 3pm, 1w, 2025-12-31) #tag +Project !urgent *'note'")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.Name != "Task" {
		t.Errorf("Expected name 'Task', got %q", task.Name)
	}

	if task.Tag != "tag" {
		t.Errorf("Expected tag 'tag', got %q", task.Tag)
	}

	if task.Project != "Project" {
		t.Errorf("Expected project 'Project', got %q", task.Project)
	}

	if task.Priority != "urgent" {
		t.Errorf("Expected priority 'urgent', got %q", task.Priority)
	}

	if task.Note != "note" {
		t.Errorf("Expected note 'note', got %q", task.Note)
	}

	if task.DueDate == nil {
		t.Fatal("Expected due date to be set")
	}

	if task.RecurFrequency != "1w" {
		t.Errorf("Expected frequency '1w', got %q", task.RecurFrequency)
	}

	if task.RecurEndDate == nil {
		t.Fatal("Expected recur end date to be set")
	}
}

func TestParser_SemanticDates(t *testing.T) {
	tests := []struct {
		input string
		check func(*time.Time) bool
	}{
		{"Task @(today)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, Today())
		}},
		{"Task @(tomorrow)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, Tomorrow())
		}},
		{"Task @(tmr)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, Tomorrow())
		}},
		{"Task @(yesterday)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, Yesterday())
		}},
		{"Task @(this week)", func(d *time.Time) bool {
			if d == nil {
				return false
			}
			today := Today()
			earliest := today.AddDate(0, 0, -6)
			return !d.After(today) && !d.Before(earliest)
		}},
		{"Task @(this month)", func(d *time.Time) bool {
			return d != nil && d.Year() == Today().Year() && d.Month() == Today().Month()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			task, err := parser.ParseTaskString(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if !tt.check(task.DueDate) {
				t.Errorf("Date check failed for %q", tt.input)
			}
		})
	}
}

func TestParser_RelativeDates(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		checkFn func(*time.Time) bool
	}{
		{"+3d", "Task @(+3d)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, DaysFromNow(3))
		}},
		{"-2w", "Task @(-2w)", func(d *time.Time) bool {
			return d != nil && isSameDay(*d, DaysFromNow(-14))
		}},
		{"+1m", "Task @(+1m)", func(d *time.Time) bool {
			return d != nil && d.Month() == Today().AddDate(0, 1, 0).Month()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := parser.ParseTaskString(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if !tt.checkFn(task.DueDate) {
				t.Errorf("Date check failed for %q", tt.input)
			}
		})
	}
}

func TestParser_StandardFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		year  int
		month time.Month
		day   int
		hour  int
		min   int
		sec   int
	}{
		{"ISO datetime", "Task @(2025-11-15 15:04)", 2025, time.November, 15, 15, 4, 0},
		{"ISO datetime with seconds", "Task @(2025-11-15 15:04:05)", 2025, time.November, 15, 15, 4, 5},
		{"ISO date only", "Task @(2025-11-15)", 2025, time.November, 15, 12, 0, 0},
		{"Slash date", "Task @(2025/11/20)", 2025, time.November, 20, 12, 0, 0},
		{"European date", "Task @(20 Nov 2025)", 2025, time.November, 20, 12, 0, 0},
		{"European numeric", "Task @(20-11-2025)", 2025, time.November, 20, 12, 0, 0},
		{"US format", "Task @(11/20/2025)", 2025, time.November, 20, 12, 0, 0},
		{"Month name day year", "Task @(Jan 2 2006)", 2006, time.January, 2, 12, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := parser.ParseTaskString(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if task.DueDate == nil {
				t.Fatal("Expected due date")
			}
			expected := time.Date(tt.year, tt.month, tt.day, tt.hour, tt.min, tt.sec, 0, task.DueDate.Location())
			if !task.DueDate.Equal(expected) {
				t.Errorf("Expected %v, got %v", expected, task.DueDate)
			}
		})
	}
}

func TestParser_Frequencies(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		frequency string
	}{
		{"daily", "Task @(2025-11-15, daily)", "daily"},
		{"weekly", "Task @(2025-11-15, weekly)", "weekly"},
		{"monthly", "Task @(2025-11-15, monthly)", "monthly"},
		{"yearly", "Task @(2025-11-15, yearly)", "yearly"},
		{"n_hours", "Task @(2025-11-15, 12h)", "12h"},
		{"n_days", "Task @(2025-11-15, 3d)", "3d"},
		{"n_weeks", "Task @(2025-11-15, 2w)", "2w"},
		{"n_months", "Task @(2025-11-15, 6m)", "6m"},
		{"n_years", "Task @(2025-11-15, 1y)", "1y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := parser.ParseTaskString(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if task.RecurFrequency != tt.frequency {
				t.Errorf("Expected frequency %q, got %q", tt.frequency, task.RecurFrequency)
			}
		})
	}
}

func TestParser_WithEndDate(t *testing.T) {
	task, err := parser.ParseTaskString("Task @(2025-11-15, 1w, 2025-12-31)")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if task.RecurEndDate == nil {
		t.Fatal("Expected recur end date")
	}

	expected := time.Date(2025, 12, 31, 12, 0, 0, 0, task.RecurEndDate.Location())
	if !task.RecurEndDate.Equal(expected) {
		t.Errorf("Expected end date %v, got %v", expected, task.RecurEndDate)
	}
}

func TestParser_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid date", "Task @(not-a-date)"},
		{"frequency without date", "Task @(, 1w)"},
		{"recurrence without date", "Task @(bad-date, 1w, 2025-12-31)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseTaskString(tt.input)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

func isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func TestParser_EmptyString(t *testing.T) {
	_, err := parser.ParseTaskString("")
	if err == nil {
		t.Fatal("Expected error for empty string")
	}
}

func TestParser_OnlySpaces(t *testing.T) {
	_, err := parser.ParseTaskString("   ")
	if err == nil {
		t.Fatal("Expected error for only spaces")
	}
}
