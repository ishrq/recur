package integration

import (
	"testing"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/filter"
)

func TestFilter_ByTag(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Work task", WithTag("work"))
	CreateTestTask(t, database, "Personal task", WithTag("personal"))
	CreateTestTask(t, database, "Untagged task")

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	filters := filter.Filters{Tags: []string{"work"}}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task, got %d", len(filtered))
	}

	if filtered[0].Tag != "work" {
		t.Errorf("Expected work tag, got %q", filtered[0].Tag)
	}
}

func TestFilter_ByProject(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Work task", WithProject("Work"))
	CreateTestTask(t, database, "Home task", WithProject("Home"))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	filters := filter.Filters{Projects: []string{"Home"}}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task, got %d", len(filtered))
	}

	if filtered[0].Project != "Home" {
		t.Errorf("Expected Home project, got %q", filtered[0].Project)
	}
}

func TestFilter_ByPriority(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Urgent task", WithPriority("urgent"))
	CreateTestTask(t, database, "Low task", WithPriority("low"))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	filters := filter.Filters{Priorities: []string{"urgent"}}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task, got %d", len(filtered))
	}

	if filtered[0].Priority != "urgent" {
		t.Errorf("Expected urgent priority, got %q", filtered[0].Priority)
	}
}

func TestFilter_ByQuery(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Buy groceries", WithNote("milk and bread"))
	CreateTestTask(t, database, "Team meeting", WithTag("work"))
	CreateTestTask(t, database, "Call doctor", WithProject("Health"))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	// Search for "meeting"
	filters := filter.Filters{Query: "meeting"}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task, got %d", len(filtered))
	}

	if filtered[0].Name != "Team meeting" {
		t.Errorf("Expected 'Team meeting', got %q", filtered[0].Name)
	}

	// Search for "milk" (should find in note)
	filters = filter.Filters{Query: "milk"}
	filtered, err = filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task, got %d", len(filtered))
	}
}

func TestFilter_ByDueDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Today task", WithDueDate(Today()))
	CreateTestTask(t, database, "Tomorrow task", WithDueDate(Tomorrow()))
	CreateTestTask(t, database, "Future task", WithDueDate(DaysFromNow(10)))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	// Filter by today
	filters := filter.Filters{Today: true}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task for today, got %d", len(filtered))
	}

	// Filter by tomorrow
	filters = filter.Filters{Tomorrow: true}
	filtered, err = filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task for tomorrow, got %d", len(filtered))
	}
}

func TestFilter_DateRange(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task 1", WithDueDate(DaysFromNow(1)))
	CreateTestTask(t, database, "Task 2", WithDueDate(DaysFromNow(5)))
	CreateTestTask(t, database, "Task 3", WithDueDate(DaysFromNow(10)))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	// Filter from day 1 to day 7
	filters := filter.Filters{
		FromDate: "tomorrow",
		ToDate:   DaysFromNow(7).Format("2006-01-02"),
	}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tasks in range, got %d", len(filtered))
	}
}

func TestFilter_CombinedFilters(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Work task 1", WithTag("work"), WithProject("Office"), WithDueDate(Today()))
	CreateTestTask(t, database, "Work task 2", WithTag("work"), WithProject("Office"), WithDueDate(Tomorrow()))
	CreateTestTask(t, database, "Personal task", WithTag("personal"), WithProject("Home"), WithDueDate(Today()))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	// Filter by tag AND project AND today
	filters := filter.Filters{
		Tags:     []string{"work"},
		Projects: []string{"Office"},
		Today:    true,
	}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 task matching all filters, got %d", len(filtered))
	}

	if filtered[0].Name != "Work task 1" {
		t.Errorf("Expected 'Work task 1', got %q", filtered[0].Name)
	}
}

func TestFilter_Overdue(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Overdue task", WithDueDate(Yesterday()))
	CreateTestTask(t, database, "Today task", WithDueDate(Today()))
	CreateTestTask(t, database, "Future task", WithDueDate(Tomorrow()))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	filters := filter.Filters{Overdue: true}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 overdue task, got %d", len(filtered))
	}

	if filtered[0].Name != "Overdue task" {
		t.Errorf("Expected 'Overdue task', got %q", filtered[0].Name)
	}
}

func TestFilter_Upcoming(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Today task", WithDueDate(Today()))
	CreateTestTask(t, database, "Tomorrow task", WithDueDate(Tomorrow()))
	CreateTestTask(t, database, "This week", WithDueDate(DaysFromNow(5)))
	CreateTestTask(t, database, "Next week", WithDueDate(DaysFromNow(10)))

	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	// Upcoming should be after today and within next 7 days
	filters := filter.Filters{Upcoming: true}
	filtered, err := filter.ApplyFilters(tasks, filters)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 upcoming tasks, got %d", len(filtered))
	}
}
