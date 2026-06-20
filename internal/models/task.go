package models

import "time"

type Task struct {
	ID             int
	Name           string
	DueDate        *time.Time
	CreatedDate    time.Time
	CompletedDate  *time.Time
	Tag            string
	Project        string
	Priority       string
	Note           string
	LastTaskID     *int
	RecurFrequency string
	RecurEndDate   *time.Time
	Deleted        bool
}

func (t *Task) Clone() *Task {
	return &Task{
		Name:           t.Name,
		DueDate:        t.DueDate,
		CreatedDate:    time.Now(),
		Tag:            t.Tag,
		Project:        t.Project,
		Priority:       t.Priority,
		Note:           t.Note,
		RecurFrequency: t.RecurFrequency,
		RecurEndDate:   t.RecurEndDate,
	}
}

func (t *Task) Merge(changes *Task) {
	t.Name = changes.Name
	if changes.DueDate != nil {
		t.DueDate = changes.DueDate
	}
	if changes.Tag != "" {
		t.Tag = changes.Tag
	}
	if changes.Project != "" {
		t.Project = changes.Project
	}
	if changes.Priority != "" {
		t.Priority = changes.Priority
	}
	if changes.Note != "" {
		t.Note = changes.Note
	}
	if changes.RecurFrequency != "" {
		t.RecurFrequency = changes.RecurFrequency
	}
	if changes.RecurEndDate != nil {
		t.RecurEndDate = changes.RecurEndDate
	}
}
