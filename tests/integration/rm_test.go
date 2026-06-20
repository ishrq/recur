package integration

import (
	"github.com/ishrq/recur/internal/db"
	"testing"
)

func TestRemoveCommand_SingleTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task to delete")

	err := TestRemove(database, []string{"1"})
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	AssertTaskDeleted(t, database, int(id))
}

func TestRemoveCommand_MultipleTasks(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id1 := CreateTestTask(t, database, "Task 1")
	id2 := CreateTestTask(t, database, "Task 2")
	id3 := CreateTestTask(t, database, "Task 3")

	err := TestRemove(database, []string{"1", "2", "3"})
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	AssertTaskDeleted(t, database, int(id1))
	AssertTaskDeleted(t, database, int(id2))
	AssertTaskDeleted(t, database, int(id3))
}

func TestRemoveCommand_NonExistent(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	// Should handle gracefully
	err := TestRemove(database, []string{"999"})
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}
}

func TestRemoveCommand_AlreadyDeleted(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")
	TestRemove(database, []string{"1"})

	// Should handle gracefully
	err := TestRemove(database, []string{"1"})
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	AssertTaskDeleted(t, database, int(id))
}

func TestRemoveCommand_Undo(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")

	// Delete
	TestRemove(database, []string{"1"})
	AssertTaskDeleted(t, database, int(id))

	// Restore
	err := TestRemove(database, []string{"--undo", "1"})
	if err != nil {
		t.Fatalf("Remove undo failed: %v", err)
	}

	task := GetTaskByID(t, database, int(id))
	if task.Deleted {
		t.Error("Task should be restored (not deleted)")
	}
}

func TestRemoveCommand_UndoMultiple(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id1 := CreateTestTask(t, database, "Task 1")
	id2 := CreateTestTask(t, database, "Task 2")

	// Delete
	TestRemove(database, []string{"1", "2"})

	// Restore
	err := TestRemove(database, []string{"--undo", "1", "2"})
	if err != nil {
		t.Fatalf("Remove undo failed: %v", err)
	}

	task1 := GetTaskByID(t, database, int(id1))
	task2 := GetTaskByID(t, database, int(id2))

	if task1.Deleted || task2.Deleted {
		t.Error("Tasks should be restored")
	}
}

func TestRemoveCommand_PermanentDelete(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")

	// Soft delete first
	TestRemove(database, []string{"1"})
	AssertTaskDeleted(t, database, int(id))

	// Permanent delete
	err := TestRemove(database, []string{"--trash", "1"})
	if err != nil {
		t.Fatalf("Remove trash failed: %v", err)
	}

	// Task should not exist at all now
	task, err := db.GetTaskByID(database, int(id))
	if err == nil && task != nil {
		t.Error("Task should be permanently deleted")
	}
}
