package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
)

func Copy(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("cp")
		return nil
	}

	// Parse task IDs and optional modification string
	var ids []int
	var modifyStr string
	foundNonID := false

	for i, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			// First non-ID argument is the modification string
			modifyStr = strings.Join(args[i:], " ")
			foundNonID = true
			break
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return fmt.Errorf("at least one task ID required")
	}

	// If modification string and multiple IDs, return error
	if foundNonID && len(ids) > 1 {
		return fmt.Errorf("cannot modify multiple tasks at once")
	}

	// Copy tasks
	copied := 0
	for _, id := range ids {
		task, err := db.GetTaskByID(database, id)
		if err != nil {
			fmt.Printf("Warning: Task #%d not found\n", id)
			continue
		}

		// Create new task (copy of original)
		newTask := &models.Task{
			Name:        task.Name,
			DueDate:     task.DueDate,
			CreatedDate: time.Now(),
			Tag:         task.Tag,
			Project:     task.Project,
			Priority:    task.Priority,
			Note:        task.Note,
		}

		// If modification string provided, parse and merge changes
		if modifyStr != "" {
			parsedChanges, err := parser.ParseTaskString(modifyStr)
			if err == nil {
				// Merge changes (new values override old ones)
				newTask.Name = parsedChanges.Name
				if parsedChanges.DueDate != nil {
					newTask.DueDate = parsedChanges.DueDate
				}
				if parsedChanges.Tag != "" {
					newTask.Tag = parsedChanges.Tag
				}
				if parsedChanges.Project != "" {
					newTask.Project = parsedChanges.Project
				}
				if parsedChanges.Priority != "" {
					newTask.Priority = parsedChanges.Priority
				}
				if parsedChanges.Note != "" {
					newTask.Note = parsedChanges.Note
				}
			}
		}

		newID, err := db.InsertTask(database, newTask)
		if err != nil {
			fmt.Printf("Warning: Failed to copy task #%d: %v\n", id, err)
			continue
		}

		fmt.Printf("✓ Copied #%d → #%d: %s\n", id, newID, newTask.Name)
		copied++
	}

	if copied > 0 {
		fmt.Printf("\n%d task(s) copied\n", copied)
	}

	return nil
}
