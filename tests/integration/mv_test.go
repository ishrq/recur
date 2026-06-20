package integration

import (
	"testing"

	"github.com/ishrq/recur/internal/commands"
)

func TestMoveCommand_SingleTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Original name", WithTag("work"))

	err := commands.Move(database, []string{"1", "Updated name #personal"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskName(t, task, "Updated name")
	AssertTaskTag(t, task, "personal")
}

func TestMoveCommand_MultipleTasks(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task 1")
	CreateTestTask(t, database, "Task 2")

	err := commands.Move(database, []string{"1", "2", "Updated name +Project"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task1 := AssertTaskExists(t, database, 1)
	task2 := AssertTaskExists(t, database, 2)

	AssertTaskName(t, task1, "Updated name")
	AssertTaskProject(t, task1, "Project")

	AssertTaskName(t, task2, "Updated name")
	AssertTaskProject(t, task2, "Project")
}

func TestMoveCommand_UpdateTag(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task", WithTag("old"))

	err := commands.Move(database, []string{"1", "Task #new"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskTag(t, task, "new")
}

func TestMoveCommand_UpdateProject(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task", WithProject("OldProject"))

	err := commands.Move(database, []string{"1", "Task +NewProject"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskProject(t, task, "NewProject")
}

func TestMoveCommand_UpdatePriority(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task", WithPriority("low"))

	err := commands.Move(database, []string{"1", "Task !urgent"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	AssertTaskPriority(t, task, "urgent")
}

func TestMoveCommand_UpdateDueDate(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task", WithDueDate(Today()))

	err := commands.Move(database, []string{"1", "Task @(2025-12-31 5pm)"})
	if err != nil {
		t.Fatalf("Move command failed: %v", err)
	}

	task := AssertTaskExists(t, database, 1)
	if task.DueDate == nil {
		t.Fatal("Expected due date to be set")
	}
}

func TestMoveCommand_DeletedTask(t *testing.T) {
	database, cleanup := SetupTestDB(t)
	defer cleanup()

	CreateTestTask(t, database, "Task")
	commands.Remove(database, []string{"1"})

	// Should report error since task is deleted
	err := commands.Move(database, []string{"1", "Updated name"})
	if err == nil {
		t.Fatal("Expected error for deleted task")
	}

	task := GetTaskByID(t, database, 1)
	AssertTaskName(t, task, "Task") // Name should not change
}
