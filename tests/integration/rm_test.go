package integration

import (
	"testing"

	"github.com/ishrq/recur/internal/commands"
	"github.com/ishrq/recur/internal/db"
)

func TestRemoveCommand_SingleTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task to delete")

	err := commands.Remove(database, []string{"1"})
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

	err := commands.Remove(database, []string{"1", "2", "3"})
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

	// Should report error since no task with this ID exists
	err := commands.Remove(database, []string{"999"})
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
}

func TestRemoveCommand_AlreadyDeleted(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")
	commands.Remove(database, []string{"1"})

	// Should report error since task is already deleted
	err := commands.Remove(database, []string{"1"})
	if err == nil {
		t.Fatal("Expected error for already deleted task")
	}

	AssertTaskDeleted(t, database, int(id))
}

func TestRemoveCommand_Undo(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	id := CreateTestTask(t, database, "Task")

	// Delete
	commands.Remove(database, []string{"1"})
	AssertTaskDeleted(t, database, int(id))

	// Restore
	err := commands.Remove(database, []string{"--undo", "1"})
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
	commands.Remove(database, []string{"1", "2"})

	// Restore
	err := commands.Remove(database, []string{"--undo", "1", "2"})
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
	commands.Remove(database, []string{"1"})
	AssertTaskDeleted(t, database, int(id))

	// Permanent delete
	err := commands.Remove(database, []string{"--trash", "1"})
	if err != nil {
		t.Fatalf("Remove trash failed: %v", err)
	}

	// Task should not exist at all now
	task, err := db.GetTaskByID(database, int(id))
	if err == nil && task != nil {
		t.Error("Task should be permanently deleted")
	}
}
