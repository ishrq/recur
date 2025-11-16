package integration

import (
	"testing"
)

func TestCopyCommand_SingleTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Original task", WithTag("work"))

	err := TestCopy(database, []string{"1"})
	if err != nil {
		t.Fatalf("Copy command failed: %v", err)
	}

	// Should have 2 tasks now
	AssertTaskCount(t, database, false, false, 2)

	// Check original still exists
	original := AssertTaskExists(t, database, 1)
	AssertTaskName(t, original, "Original task")
	AssertTaskTag(t, original, "work")

	// Check copy
	copy := AssertTaskExists(t, database, 2)
	AssertTaskName(t, copy, "Original task")
	AssertTaskTag(t, copy, "work")
}

func TestCopyCommand_MultipleTasks(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task 1")
	CreateTestTask(t, database, "Task 2")

	err := TestCopy(database, []string{"1", "2"})
	if err != nil {
		t.Fatalf("Copy command failed: %v", err)
	}

	// Should have 4 tasks now (2 originals + 2 copies)
	AssertTaskCount(t, database, false, false, 4)
}

func TestCopyCommand_WithModification(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Original task", WithTag("work"))

	err := TestCopy(database, []string{"1", "Modified copy #personal"})
	if err != nil {
		t.Fatalf("Copy command failed: %v", err)
	}

	AssertTaskCount(t, database, false, false, 2)

	// Check copy has modifications
	copy := AssertTaskExists(t, database, 2)
	AssertTaskName(t, copy, "Modified copy")
	AssertTaskTag(t, copy, "personal")
}

func TestCopyCommand_CompletedTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Completed task", WithCompleted(Today()))

	// Copy of completed task should be incomplete
	err := TestCopy(database, []string{"1"})
	if err != nil {
		t.Fatalf("Copy command failed: %v", err)
	}

	copy := AssertTaskExists(t, database, 2)
	AssertTaskNotCompleted(t, database, int(copy.ID))
}

func TestCopyCommand_DeletedTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task")
	TestRemove(database, []string{"1"})

	// Should not copy deleted task
	err := TestCopy(database, []string{"1"})
	if err != nil {
		t.Fatalf("Copy command failed: %v", err)
	}

	// Should still have only 1 task (the deleted one)
	AssertTaskCount(t, database, false, false, 0)
	AssertTaskCount(t, database, true, false, 1)
}
