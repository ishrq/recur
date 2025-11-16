package integration

import (
	"testing"
	"time"
)

func TestRecurringTask_Creation(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	dueDate := time.Date(2025, 11, 15, 10, 0, 0, 0, time.Local)
	id := CreateTestTask(t, database, "Weekly task",
		WithDueDate(dueDate),
		WithRecurring("1w", nil))

	task := AssertTaskExists(t, database, int(id))

	if task.RecurFrequency != "1w" {
		t.Errorf("Expected frequency '1w', got %q", task.RecurFrequency)
	}

	if task.DueDate == nil {
		t.Fatal("Expected due date to be set")
	}
}

func TestRecurringTask_WithEndDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	dueDate := time.Date(2025, 11, 15, 10, 0, 0, 0, time.Local)
	endDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)

	id := CreateTestTask(t, database, "Limited recurring task",
		WithDueDate(dueDate),
		WithRecurring("1w", &endDate))

	task := AssertTaskExists(t, database, int(id))

	if task.RecurEndDate == nil {
		t.Fatal("Expected recur end date to be set")
	}

	if !task.RecurEndDate.Equal(endDate) {
		t.Errorf("Expected end date %v, got %v", endDate, task.RecurEndDate)
	}
}

// Note: Testing the automatic creation of next occurrence when marking done
// requires handling the confirmation prompt or refactoring the done command
