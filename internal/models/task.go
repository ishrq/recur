package models

import "time"

type Task struct {
	ID            int
	Name          string
	DueDate       *time.Time
	CreatedDate   time.Time
	CompletedDate *time.Time
	Tag           string
	Project       string
	Priority      string
	Note          string
	LastTaskID    *int
}
