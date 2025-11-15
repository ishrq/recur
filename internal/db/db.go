package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ishrq/recur/internal/models"
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

func InsertTask(db *sql.DB, task *models.Task) (int64, error) {
	query := `
		INSERT INTO tasks (name, due_date, created_date, completed_date, tag, project, priority, note, last_task_id, recur_frequency, recur_end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query,
		task.Name,
		task.DueDate,
		task.CreatedDate,
		task.CompletedDate,
		task.Tag,
		task.Project,
		task.Priority,
		task.Note,
		task.LastTaskID,
		task.RecurFrequency,
		task.RecurEndDate,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetTaskByID gets any task by ID (no filtering)
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
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, err
	}

	task.Deleted = deleted == 1

	// Convert nullable fields
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if completedDate.Valid {
		task.CompletedDate = &completedDate.Time
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
	if lastTaskID.Valid {
		id := int(lastTaskID.Int64)
		task.LastTaskID = &id
	}
	if recurFrequency.Valid {
		task.RecurFrequency = recurFrequency.String
	}
	if recurEndDate.Valid {
		task.RecurEndDate = &recurEndDate.Time
	}

	return &task, nil
}

// GetTasks gets tasks based on deleted and completed status
// deleted: true = only deleted tasks, false = only non-deleted tasks
// completed: true = only completed tasks, false = only incomplete tasks
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
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
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

	task.Deleted = deleted == 1

	// Convert nullable fields
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if completedDate.Valid {
		task.CompletedDate = &completedDate.Time
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
	if lastTaskID.Valid {
		id := int(lastTaskID.Int64)
		task.LastTaskID = &id
	}
	if recurFrequency.Valid {
		task.RecurFrequency = recurFrequency.String
	}
	if recurEndDate.Valid {
		task.RecurEndDate = &recurEndDate.Time
	}

	return task, nil
}

func MarkTaskDone(db *sql.DB, id int) error {
	query := `
		UPDATE tasks
		SET completed_date = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already deleted")
	}

	return nil
}

func UpdateTask(db *sql.DB, task *models.Task) error {
	query := `
		UPDATE tasks
		SET name = ?, due_date = ?, tag = ?, project = ?, priority = ?, note = ?, last_task_id = ?, recur_frequency = ?, recur_end_date = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := db.Exec(query,
		task.Name,
		task.DueDate,
		task.Tag,
		task.Project,
		task.Priority,
		task.Note,
		task.LastTaskID,
		task.RecurFrequency,
		task.RecurEndDate,
		task.ID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already deleted")
	}

	return nil
}

func DeleteTask(db *sql.DB, id int) error {
	query := `
		UPDATE tasks
		SET deleted = 1
		WHERE id = ? AND deleted = 0
	`

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already deleted")
	}

	return nil
}

func PermanentlyDeleteTask(db *sql.DB, id int) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func PurgeAllTasks(db *sql.DB) error {
	query := `DELETE FROM tasks`
	_, err := db.Exec(query)
	return err
}

func GetAllTasksCount(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tasks`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

func RestoreTask(db *sql.DB, id int) error {
	query := `
		UPDATE tasks
		SET deleted = 0
		WHERE id = ? AND deleted = 1
	`

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already restored")
	}

	return nil
}
