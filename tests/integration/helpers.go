package integration

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/ishrq/recur/internal/commands"
	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

func init() {
	commands.ConfirmPrompt = func(_ string) (bool, error) { return true, nil }
	commands.ConfirmSpecific = func(_, _ string) (bool, error) { return true, nil }
}

// SetupTestDB creates a temporary database for testing
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

// CreateTestTask creates a task with given parameters
func CreateTestTask(t *testing.T, database *sql.DB, name string, opts ...TaskOption) int64 {
	t.Helper()

	task := &models.Task{
		Name:        name,
		CreatedDate: time.Now(),
	}

	for _, opt := range opts {
		opt(task)
	}

	id, err := db.InsertTask(database, task)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	return id
}

// TaskOption is a function that modifies a task
type TaskOption func(*models.Task)

// WithDueDate sets the due date
func WithDueDate(date time.Time) TaskOption {
	return func(t *models.Task) { t.DueDate = &date }
}

// WithTag sets the tag
func WithTag(tag string) TaskOption {
	return func(t *models.Task) { t.Tag = tag }
}

// WithProject sets the project
func WithProject(project string) TaskOption {
	return func(t *models.Task) { t.Project = project }
}

// WithPriority sets the priority
func WithPriority(priority string) TaskOption {
	return func(t *models.Task) { t.Priority = priority }
}

// WithNote sets the note
func WithNote(note string) TaskOption {
	return func(t *models.Task) { t.Note = note }
}

// WithRecurring sets recurring frequency and optional end date
func WithRecurring(frequency string, endDate *time.Time) TaskOption {
	return func(t *models.Task) {
		t.RecurFrequency = frequency
		t.RecurEndDate = endDate
	}
}

// WithCompleted marks task as completed
func WithCompleted(date time.Time) TaskOption {
	return func(t *models.Task) { t.CompletedDate = &date }
}

// GetTaskByID is a helper to retrieve a task and fail if not found
func GetTaskByID(t *testing.T, database *sql.DB, id int) *models.Task {
	t.Helper()

	task, err := db.GetTaskByID(database, id)
	if err != nil {
		t.Fatalf("Failed to get task %d: %v", id, err)
	}

	return task
}

// AssertTaskExists checks if a task exists and returns it
func AssertTaskExists(t *testing.T, database *sql.DB, id int) *models.Task {
	t.Helper()
	return GetTaskByID(t, database, id)
}

// AssertTaskNotFound checks if a task doesn't exist
func AssertTaskNotFound(t *testing.T, database *sql.DB, id int) {
	t.Helper()

	task, err := db.GetTaskByID(database, id)
	if err == nil && task != nil && !task.Deleted {
		t.Fatalf("Expected task %d to not exist, but it does", id)
	}
}

// AssertTaskDeleted checks if a task is soft-deleted
func AssertTaskDeleted(t *testing.T, database *sql.DB, id int) {
	t.Helper()

	task := GetTaskByID(t, database, id)
	if !task.Deleted {
		t.Fatalf("Expected task %d to be deleted, but it's not", id)
	}
}

// AssertTaskCompleted checks if a task is completed
func AssertTaskCompleted(t *testing.T, database *sql.DB, id int) {
	t.Helper()

	task := GetTaskByID(t, database, id)
	if task.CompletedDate == nil {
		t.Fatalf("Expected task %d to be completed, but it's not", id)
	}
}

// AssertTaskNotCompleted checks if a task is not completed
func AssertTaskNotCompleted(t *testing.T, database *sql.DB, id int) {
	t.Helper()

	task := GetTaskByID(t, database, id)
	if task.CompletedDate != nil {
		t.Fatalf("Expected task %d to not be completed, but it is", id)
	}
}

// AssertTaskCount checks if the number of tasks matches expected
func AssertTaskCount(t *testing.T, database *sql.DB, deleted, completed bool, expected int) {
	t.Helper()

	tasks, err := db.GetTasks(database, deleted, completed)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != expected {
		t.Fatalf("Expected %d tasks, got %d", expected, len(tasks))
	}
}

// AssertTaskField checks if a task field matches expected value
func AssertTaskField(t *testing.T, task *models.Task, field, expected, actual string) {
	t.Helper()

	if actual != expected {
		t.Errorf("Task #%d %s: expected %q, got %q", task.ID, field, expected, actual)
	}
}

// AssertTaskName checks if task name matches
func AssertTaskName(t *testing.T, task *models.Task, expected string) {
	t.Helper()
	AssertTaskField(t, task, "name", expected, task.Name)
}

// AssertTaskTag checks if task tag matches
func AssertTaskTag(t *testing.T, task *models.Task, expected string) {
	t.Helper()
	AssertTaskField(t, task, "tag", expected, task.Tag)
}

// AssertTaskProject checks if task project matches
func AssertTaskProject(t *testing.T, task *models.Task, expected string) {
	t.Helper()
	AssertTaskField(t, task, "project", expected, task.Project)
}

// AssertTaskPriority checks if task priority matches
func AssertTaskPriority(t *testing.T, task *models.Task, expected string) {
	t.Helper()
	AssertTaskField(t, task, "priority", expected, task.Priority)
}

// Today returns today at midnight
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow returns tomorrow at midnight
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// Yesterday returns yesterday at midnight
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// DaysFromNow returns a date n days from now
func DaysFromNow(days int) time.Time {
	return Today().AddDate(0, 0, days)
}
