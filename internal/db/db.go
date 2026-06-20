package db

import (
	"database/sql"
	"fmt"
	"github.com/ishrq/recur/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		due_date DATETIME,
		created_date DATETIME NOT NULL,
		completed_date DATETIME,
		tag TEXT,
		project TEXT,
		priority TEXT,
		note TEXT,
		last_task_id INTEGER,
		deleted INTEGER DEFAULT 0,
		recur_frequency TEXT,
		recur_end_date DATETIME
	);
	CREATE INDEX IF NOT EXISTS idx_due_date ON tasks(due_date);
	CREATE INDEX IF NOT EXISTS idx_tag ON tasks(tag);
	CREATE INDEX IF NOT EXISTS idx_project ON tasks(project);
	CREATE INDEX IF NOT EXISTS idx_completed_date ON tasks(completed_date);
	CREATE INDEX IF NOT EXISTS idx_deleted ON tasks(deleted);
	CREATE INDEX IF NOT EXISTS idx_recur_frequency ON tasks(recur_frequency);
	`
	_, err := db.Exec(schema)
	return err
}

func toUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}

func InsertTask(db *sql.DB, task *models.Task) (int64, error) {
	query := `
		INSERT INTO tasks (name, due_date, created_date, completed_date, tag, project, priority, note, last_task_id, recur_frequency, recur_end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query,
		task.Name,
		toUTC(task.DueDate),
		task.CreatedDate.UTC(),
		toUTC(task.CompletedDate),
		task.Tag,
		task.Project,
		task.Priority,
		task.Note,
		task.LastTaskID,
		task.RecurFrequency,
		toUTC(task.RecurEndDate),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert task '%s': %w", task.Name, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func GetTaskByID(db *sql.DB, id int) (*models.Task, error) {
	query := `
		SELECT id, name, due_date, created_date, completed_date,
		       tag, project, priority, note, last_task_id, recur_frequency, recur_end_date, deleted
		FROM tasks
		WHERE id = ?
	`

	row := db.QueryRow(query, id)

	var task models.Task
	var dueDate, completedDate, recurEndDate sql.NullTime
	var tag, project, priority, note, recurFrequency sql.NullString
	var lastTaskID sql.NullInt64
	var deleted int

	err := row.Scan(
		&task.ID,
		&task.Name,
		&dueDate,
		&task.CreatedDate,
		&completedDate,
		&tag,
		&project,
		&priority,
		&note,
		&lastTaskID,
		&recurFrequency,
		&recurEndDate,
		&deleted,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task %d: %w", id, err)
	}

	convertNullableFields(&task, dueDate, completedDate, recurEndDate, tag, project, priority, note, recurFrequency, lastTaskID, deleted)

	return &task, nil
}

func GetTasks(db *sql.DB, deleted bool, completed bool) ([]models.Task, error) {
	query := `
		SELECT id, name, due_date, created_date, completed_date,
		       tag, project, priority, note, last_task_id, recur_frequency, recur_end_date, deleted
		FROM tasks
		WHERE deleted = ?
	`

	if completed {
		query += " AND completed_date IS NOT NULL"
	} else {
		query += " AND completed_date IS NULL"
	}

	query += " ORDER BY CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, id ASC"

	rows, err := db.Query(query, boolToInt(deleted))
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks (deleted=%v, completed=%v): %w", deleted, completed, err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

func scanTask(rows *sql.Rows) (models.Task, error) {
	var task models.Task
	var dueDate, completedDate, recurEndDate sql.NullTime
	var tag, project, priority, note, recurFrequency sql.NullString
	var lastTaskID sql.NullInt64
	var deleted int

	err := rows.Scan(
		&task.ID,
		&task.Name,
		&dueDate,
		&task.CreatedDate,
		&completedDate,
		&tag,
		&project,
		&priority,
		&note,
		&lastTaskID,
		&recurFrequency,
		&recurEndDate,
		&deleted,
	)

	if err != nil {
		return task, err
	}

	convertNullableFields(&task, dueDate, completedDate, recurEndDate, tag, project, priority, note, recurFrequency, lastTaskID, deleted)

	return task, nil
}

func MarkTaskDone(db *sql.DB, id int) error {
	task, err := GetTaskByID(db, id)
	if err != nil {
		return fmt.Errorf("failed to mark task %d as done: %w", id, err)
	}

	if task.Deleted {
		return fmt.Errorf("task %d is deleted", id)
	}

	if task.CompletedDate != nil {
		return fmt.Errorf("task %d is already completed", id)
	}

	now := time.Now()
	task.CompletedDate = &now
	return UpdateTask(db, task)
}

func UpdateTask(db *sql.DB, task *models.Task) error {
	query := `
		UPDATE tasks
		SET name = ?, due_date = ?, tag = ?, project = ?, priority = ?, note = ?, last_task_id = ?, recur_frequency = ?, recur_end_date = ?, completed_date = ?, deleted = ?
		WHERE id = ?
	`

	result, err := db.Exec(query,
		task.Name,
		toUTC(task.DueDate),
		task.Tag,
		task.Project,
		task.Priority,
		task.Note,
		task.LastTaskID,
		task.RecurFrequency,
		toUTC(task.RecurEndDate),
		toUTC(task.CompletedDate),
		boolToInt(task.Deleted),
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task %d: %w", task.ID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for task %d: %w", task.ID, err)
	}

	if rows == 0 {
		return fmt.Errorf("task %d not found", task.ID)
	}

	return nil
}

func DeleteTask(db *sql.DB, id int) error {
	task, err := GetTaskByID(db, id)
	if err != nil {
		return fmt.Errorf("failed to delete task %d: %w", id, err)
	}

	if task.Deleted {
		return fmt.Errorf("task %d is already deleted", id)
	}

	task.Deleted = true
	return UpdateTask(db, task)
}

func PermanentlyDeleteTask(db *sql.DB, id int) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete task %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for task %d: %w", id, err)
	}

	if rows == 0 {
		return fmt.Errorf("task %d not found", id)
	}

	return nil
}

func PurgeAllTasks(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM tasks`); err != nil {
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM sqlite_sequence WHERE name='tasks'`); err != nil {
		return fmt.Errorf("failed to reset auto-increment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetAllTasksCount(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tasks`
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get task count: %w", err)
	}
	return count, nil
}

func RestoreTask(db *sql.DB, id int) error {
	task, err := GetTaskByID(db, id)
	if err != nil {
		return fmt.Errorf("failed to restore task %d: %w", id, err)
	}

	if !task.Deleted {
		return fmt.Errorf("task %d is not deleted", id)
	}

	task.Deleted = false
	return UpdateTask(db, task)
}

func UndoCompleteTask(db *sql.DB, id int) error {
	task, err := GetTaskByID(db, id)
	if err != nil {
		return fmt.Errorf("failed to undo completion for task %d: %w", id, err)
	}

	if task.Deleted {
		return fmt.Errorf("task %d is deleted", id)
	}

	if task.CompletedDate == nil {
		return fmt.Errorf("task %d is not completed", id)
	}

	task.CompletedDate = nil
	return UpdateTask(db, task)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// convertNullableFields converts SQL nullable types to Go types
func convertNullableFields(task *models.Task, dueDate, completedDate, recurEndDate sql.NullTime,
	tag, project, priority, note, recurFrequency sql.NullString, lastTaskID sql.NullInt64, deleted int) {

	task.Deleted = deleted == 1

	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if completedDate.Valid {
		task.CompletedDate = &completedDate.Time
	}
	if recurEndDate.Valid {
		task.RecurEndDate = &recurEndDate.Time
	}
	if tag.Valid {
		task.Tag = tag.String
	}
	if project.Valid {
		task.Project = project.String
	}
	if priority.Valid {
		task.Priority = priority.String
	}
	if note.Valid {
		task.Note = note.String
	}
	if recurFrequency.Valid {
		task.RecurFrequency = recurFrequency.String
	}
	if lastTaskID.Valid {
		id := int(lastTaskID.Int64)
		task.LastTaskID = &id
	}
}
