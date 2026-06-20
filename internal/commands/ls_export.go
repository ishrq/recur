package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/filter"
	"github.com/ishrq/recur/internal/models"
)

func generateExportFilename(filters filter.Filters, showAll, showDone, showTrash bool) string {
	parts := []string{"recur"}
	now := time.Now()

	if showAll {
		parts = append(parts, "all")
	} else if showDone {
		parts = append(parts, "done")
	} else if showTrash {
		parts = append(parts, "trash")
	}

	if filters.Today {
		parts = append(parts, "today")
	} else if filters.Tomorrow {
		parts = append(parts, "tomorrow")
	} else if filters.Overdue {
		parts = append(parts, "overdue")
	} else if filters.Upcoming {
		parts = append(parts, "upcoming")
	} else if filters.DueDate != "" {
		dateClean := strings.ReplaceAll(filters.DueDate, " ", "")
		dateClean = strings.ReplaceAll(dateClean, "/", "")
		dateClean = strings.ReplaceAll(dateClean, "-", "")
		parts = append(parts, dateClean)
	}

	if filters.FromDate != "" {
		fromClean := strings.ReplaceAll(filters.FromDate, "-", "")
		fromClean = strings.ReplaceAll(fromClean, "/", "")
		parts = append(parts, "from"+fromClean)
	}
	if filters.ToDate != "" {
		toClean := strings.ReplaceAll(filters.ToDate, "-", "")
		toClean = strings.ReplaceAll(toClean, "/", "")
		parts = append(parts, "to"+toClean)
	}

	for _, tag := range filters.Tags {
		parts = append(parts, tag)
	}

	for _, project := range filters.Projects {
		parts = append(parts, project)
	}

	for _, priority := range filters.Priorities {
		parts = append(parts, priority)
	}

	if filters.Query != "" {
		queryClean := strings.ReplaceAll(filters.Query, " ", "_")
		parts = append(parts, queryClean)
	}

	if len(parts) == 1 {
		parts = append(parts, now.Format("20060102_150405"))
	}

	filename := strings.Join(parts, "_") + ".csv"
	return filename
}

func exportTasksToCSV(tasks []models.Task, filename string) error {
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		filename += ".csv"
	}

	if _, err := os.Stat(filename); err == nil {
		ok, err := confirmPrompt(fmt.Sprintf("File '%s' already exists. Overwrite? (y/n): ", filename))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Export cancelled.")
			return nil
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"ID",
		"Name",
		"Due Date",
		"Frequency",
		"End Date",
		"Tag",
		"Project",
		"Priority",
		"Note",
		"Date Created",
		"Date Completed",
		"Status",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, task := range tasks {
		row := []string{
			fmt.Sprintf("%d", task.ID),
			task.Name,
			formatDateForCSV(task.DueDate),
			task.RecurFrequency,
			formatDateForCSV(task.RecurEndDate),
			task.Tag,
			task.Project,
			task.Priority,
			task.Note,
			task.CreatedDate.Format("2006-01-02 15:04:05"),
			formatDateForCSV(task.CompletedDate),
			getTaskStatus(task),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	absPath, _ := filepath.Abs(filename)
	fmt.Printf("✓ Exported %d task(s) to: %s\n", len(tasks), absPath)

	return nil
}

func formatDateForCSV(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format("2006-01-02 15:04:05")
}

func getTaskStatus(task models.Task) string {
	if task.Deleted {
		return "deleted"
	}
	if task.CompletedDate != nil {
		return "completed"
	}
	return "active"
}
