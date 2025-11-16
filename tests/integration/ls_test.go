package integration

import (
	"testing"

	"github.com/ishrq/recur/internal/commands"
)

func TestListCommand_EmptyDatabase(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Should not error on empty database
	err := commands.List(database, []string{})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}

func TestListCommand_ActiveTasks(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Create test tasks
	CreateTestTask(t, database, "Task 1")
	CreateTestTask(t, database, "Task 2", WithTag("work"))
	CreateTestTask(t, database, "Task 3", WithCompleted(Today()))

	// List should show only incomplete tasks by default
	err := commands.List(database, []string{})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	// Verify we have 2 active tasks
	AssertTaskCount(t, database, false, false, 2)
}

func TestListCommand_WithAll(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Active task")
	CreateTestTask(t, database, "Completed task", WithCompleted(Today()))

	err := commands.List(database, []string{"--all"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	// Should show both
	AssertTaskCount(t, database, false, false, 1)
	AssertTaskCount(t, database, false, true, 1)
}

func TestListCommand_WithDone(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Active task")
	CreateTestTask(t, database, "Completed task", WithCompleted(Today()))

	err := commands.List(database, []string{"--done"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	AssertTaskCount(t, database, false, true, 1)
}

func TestListCommand_WithTag(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Work task", WithTag("work"))
	CreateTestTask(t, database, "Personal task", WithTag("personal"))
	CreateTestTask(t, database, "Untagged task")

	err := commands.List(database, []string{"--tag", "work"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	// We can't easily verify output, but at least ensure no error
}

func TestListCommand_WithProject(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Work task", WithProject("Work"))
	CreateTestTask(t, database, "Home task", WithProject("Home"))

	err := commands.List(database, []string{"--project", "Work"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}

func TestListCommand_WithDueDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Today task", WithDueDate(Today()))
	CreateTestTask(t, database, "Tomorrow task", WithDueDate(Tomorrow()))

	err := commands.List(database, []string{"--due", "today"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}

func TestListCommand_DateRange(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task 1", WithDueDate(DaysFromNow(1)))
	CreateTestTask(t, database, "Task 2", WithDueDate(DaysFromNow(5)))
	CreateTestTask(t, database, "Task 3", WithDueDate(DaysFromNow(10)))

	err := commands.List(database, []string{"--from", "tomorrow", "--to", "2025-11-20"})
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}
