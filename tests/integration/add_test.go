package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/ishrq/recur/internal/commands"
)

func TestAddCommand_Simple(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Test simple task addition
	err := commands.Add(database, []string{"Buy groceries"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	// Verify task was created
	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Buy groceries")

	// Verify no other fields are set
	if task.Tag != "" {
		t.Errorf("Expected empty tag, got %q", task.Tag)
	}
	if task.Project != "" {
		t.Errorf("Expected empty project, got %q", task.Project)
	}
	if task.DueDate != nil {
		t.Errorf("Expected no due date, got %v", task.DueDate)
	}
}

func TestAddCommand_WithTag(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Buy groceries #shopping"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Buy groceries")
	AssertTaskTag(t, task, "shopping")
}

func TestAddCommand_WithProject(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Team meeting +Work"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Team meeting")
	AssertTaskProject(t, task, "Work")
}

func TestAddCommand_WithPriority(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Urgent task !urgent"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Urgent task")
	AssertTaskPriority(t, task, "urgent")
}

func TestAddCommand_WithNote(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Task with note *'This is a note'"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Task with note")
	if task.Note != "This is a note" {
		t.Errorf("Expected note 'This is a note', got %q", task.Note)
	}
}

func TestAddCommand_WithDueDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Meeting @(2025-11-15 3pm)"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Meeting")

	if task.DueDate == nil {
		t.Fatal("Expected due date to be set")
	}

	expected := time.Date(2025, 11, 15, 15, 0, 0, 0, task.DueDate.Location())
	if !task.DueDate.Equal(expected) {
		t.Errorf("Expected due date %v, got %v", expected, task.DueDate)
	}
}

func TestAddCommand_WithRecurring(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{"Weekly task @(2025-11-15, 1w)"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Weekly task")

	if task.RecurFrequency != "1w" {
		t.Errorf("Expected frequency '1w', got %q", task.RecurFrequency)
	}
}

func TestAddCommand_Complete(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Test with all fields
	err := commands.Add(database, []string{
		"Complete task @(2025-11-15 3pm, 1w, 2025-12-31) #tag +Project !urgent *'A note'",
	})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Complete task")
	AssertTaskTag(t, task, "tag")
	AssertTaskProject(t, task, "Project")
	AssertTaskPriority(t, task, "urgent")

	if task.Note != "A note" {
		t.Errorf("Expected note 'A note', got %q", task.Note)
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

func TestAddCommand_EmptyName(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	err := commands.Add(database, []string{})
	if err == nil {
		t.Fatal("Expected error for empty task name, got nil")
	}

	if !strings.Contains(err.Error(), "task name required") {
		t.Errorf("Expected 'task name required' error, got: %v", err)
	}
}

func TestAddCommand_MultipleArgs(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Arguments should be joined into a single task
	err := commands.Add(database, []string{"Buy", "groceries", "and", "milk"})
	if err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Buy groceries and milk")
}
