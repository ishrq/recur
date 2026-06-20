package integration

import (
	"testing"
	"time"

	"github.com/ishrq/recur/internal/commands"
)

func TestDoneCommand_SingleTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Test task")

	err := commands.Done(database, []string{"1"})
	if err != nil {
		t.Fatalf("Done command failed: %v", err)
	}

	AssertTaskCompleted(t, database, int(id))
}

func TestDoneCommand_MultipleTasks(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id1 := CreateTestTask(t, database, "Task 1")
	id2 := CreateTestTask(t, database, "Task 2")
	id3 := CreateTestTask(t, database, "Task 3")

	err := commands.Done(database, []string{"1", "2", "3"})
	if err != nil {
		t.Fatalf("Done command failed: %v", err)
	}

	AssertTaskCompleted(t, database, int(id1))
	AssertTaskCompleted(t, database, int(id2))
	AssertTaskCompleted(t, database, int(id3))
}

func TestDoneCommand_AlreadyCompleted(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task", WithCompleted(Today()))

	// Should report error since all tasks are already completed
	err := commands.Done(database, []string{"1"})
	if err == nil {
		t.Fatal("Expected error for already completed task")
	}
}

func TestDoneCommand_DeletedTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")
	commands.Remove(database, []string{"1"})

	// Should report error since the task is deleted
	err := commands.Done(database, []string{"1"})
	if err == nil {
		t.Fatal("Expected error for deleted task")
	}

	task := GetTaskByID(t, database, int(id))
	if task.CompletedDate != nil {
		t.Error("Deleted task should not be marked as done")
	}
}

func TestDoneCommand_RecurringTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	dueDate := time.Date(2025, 11, 15, 10, 0, 0, 0, time.Local)
	id := CreateTestTask(t, database, "Weekly task",
		WithDueDate(dueDate),
		WithRecurring("1w", nil))

	err := commands.Done(database, []string{"1"})
	if err != nil {
		t.Fatalf("Done command failed: %v", err)
	}

	// Original task should be completed
	AssertTaskCompleted(t, database, int(id))

	// New occurrence should be created
	AssertTaskCount(t, database, false, false, 1)

	// Get the new task (ID 2)
	newTask := AssertTaskExists(t, database, 2)
	AssertTaskName(t, newTask, "Weekly task")

	// Verify due date is 1 week later
	expectedDate := dueDate.AddDate(0, 0, 7)
	if newTask.DueDate == nil {
		t.Fatal("New occurrence should have due date")
	}
	if !newTask.DueDate.Equal(expectedDate) {
		t.Errorf("Expected due date %v, got %v", expectedDate, newTask.DueDate)
	}

	// Verify LastTaskID points to original
	if newTask.LastTaskID == nil || *newTask.LastTaskID != int(id) {
		t.Errorf("Expected LastTaskID to be %d, got %v", id, newTask.LastTaskID)
	}
}

func TestDoneCommand_RecurringWithEndDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	dueDate := time.Date(2025, 11, 15, 10, 0, 0, 0, time.Local)
	endDate := time.Date(2025, 11, 16, 0, 0, 0, 0, time.Local) // End before next occurrence

	id := CreateTestTask(t, database, "Limited task",
		WithDueDate(dueDate),
		WithRecurring("1w", &endDate))

	err := commands.Done(database, []string{"1"})
	if err != nil {
		t.Fatalf("Done command failed: %v", err)
	}

	// Original task should be completed
	AssertTaskCompleted(t, database, int(id))

	// No new occurrence should be created (past end date)
	AssertTaskCount(t, database, false, false, 0)
}

func TestDoneCommand_Undo(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")

	// Mark as done
	commands.Done(database, []string{"1"})
	AssertTaskCompleted(t, database, int(id))

	// Undo
	err := commands.Done(database, []string{"--undo", "1"})
	if err != nil {
		t.Fatalf("Done undo failed: %v", err)
	}

	AssertTaskNotCompleted(t, database, int(id))
}

func TestDoneCommand_UndoMultiple(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id1 := CreateTestTask(t, database, "Task 1")
	id2 := CreateTestTask(t, database, "Task 2")

	// Mark as done
	commands.Done(database, []string{"1", "2"})

	// Undo both
	err := commands.Done(database, []string{"--undo", "1", "2"})
	if err != nil {
		t.Fatalf("Done undo failed: %v", err)
	}

	AssertTaskNotCompleted(t, database, int(id1))
	AssertTaskNotCompleted(t, database, int(id2))
}
