package commands

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

func List(database *sql.DB, args []string) error {
	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			PrintHelp("ls")
			return nil
		}
	}

	// Check for --all flag
	showAll := false
	for _, arg := range args {
		if arg == "--all" || arg == "-a" {
			showAll = true
			break
		}
	}

	tasks, err := db.GetTasks(database, showAll)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	printTasks(tasks)
	return nil
}

func printTasks(tasks []models.Task) {
	fmt.Println()
	for _, task := range tasks {
		// Format: #ID Task Name [due date]
		line := fmt.Sprintf("#%-4d %s", task.ID, task.Name)

		if task.DueDate != nil {
			dueStr := formatDueDate(*task.DueDate)
			line += fmt.Sprintf(" [%s]", dueStr)
		}

		// Add project/tag/priority if present
		var metadata []string
		if task.Project != "" {
			metadata = append(metadata, "!"+task.Project)
		}
		if task.Tag != "" {
			metadata = append(metadata, "#"+task.Tag)
		}
		if task.Priority != "" {
			metadata = append(metadata, "!"+task.Priority)
		}

		if len(metadata) > 0 {
			line += " ("
			for i, m := range metadata {
				if i > 0 {
					line += ", "
				}
				line += m
			}
			line += ")"
		}

		fmt.Println(line)
	}
	fmt.Println()
}

func formatDueDate(dueDate time.Time) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())

	daysDiff := int(due.Sub(today).Hours() / 24)

	switch {
	case daysDiff < 0:
		return fmt.Sprintf("Overdue: %s", dueDate.Format("Mon Jan 2, 15:04"))
	case daysDiff == 0:
		return fmt.Sprintf("Today %s", dueDate.Format("15:04"))
	case daysDiff == 1:
		return fmt.Sprintf("Tomorrow %s", dueDate.Format("15:04"))
	case daysDiff <= 7:
		return dueDate.Format("Mon 15:04")
	default:
		return dueDate.Format("Jan 2, 15:04")
	}
}
