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
