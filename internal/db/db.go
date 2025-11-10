package db

import (
	"database/sql"
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
	deleted INTEGER DEFAULT 0
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS tasks_fts USING fts5(
	name,
	tag,
	project,
	note,
	content=tasks,
	content_rowid=id
	);

	CREATE TRIGGER IF NOT EXISTS tasks_ai AFTER INSERT ON tasks BEGIN
	INSERT INTO tasks_fts(rowid, name, tag, project, note)
	VALUES (new.id, new.name, new.tag, new.project, new.note);
	END;

	CREATE TRIGGER IF NOT EXISTS tasks_ad AFTER DELETE ON tasks BEGIN
	DELETE FROM tasks_fts WHERE rowid = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS tasks_au AFTER UPDATE ON tasks BEGIN
	UPDATE tasks_fts SET name=new.name, tag=new.tag, project=new.project, note=new.note
	WHERE rowid=old.id;
	END;

	CREATE INDEX IF NOT EXISTS idx_due_date ON tasks(due_date);
	CREATE INDEX IF NOT EXISTS idx_tag ON tasks(tag);
	CREATE INDEX IF NOT EXISTS idx_project ON tasks(project);
	CREATE INDEX IF NOT EXISTS idx_completed_date ON tasks(completed_date);
	CREATE INDEX IF NOT EXISTS idx_deleted ON tasks(deleted);
	`

	_, err := db.Exec(schema)
	return err
}

func InsertTask(db *sql.DB, task *models.Task) (int64, error) {
	query := `
		INSERT INTO tasks (name, due_date, created_date, completed_date, tag, project, priority, note, last_task_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
