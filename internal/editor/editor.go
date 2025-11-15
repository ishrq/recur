package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
)

// OpenEditor opens $EDITOR with content and returns edited content
func OpenEditor(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "", fmt.Errorf("$EDITOR environment variable not set. Please set it to your preferred editor (e.g., export EDITOR=vim)")
	}

	// Create temporary file
	tmpfile, err := os.CreateTemp("", "recur-edit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write content to temp file
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read edited content
	edited, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return string(edited), nil
}

// FormatTaskForEditor formats a single task for editing
func FormatTaskForEditor(task *models.Task) string {
	parts := []string{task.Name}

	// Add due date if present
	if task.DueDate != nil || task.RecurFrequency != "" || task.RecurEndDate != nil {
		dateParts := []string{}

		if task.DueDate != nil {
			dateParts = append(dateParts, task.DueDate.Format("2006-01-02 15:04"))
		}

		if task.RecurFrequency != "" {
			dateParts = append(dateParts, task.RecurFrequency)
		}

		if task.RecurEndDate != nil {
			dateParts = append(dateParts, task.RecurEndDate.Format("2006-01-02"))
		}

		if len(dateParts) > 0 {
			parts = append(parts, fmt.Sprintf("@(%s)", strings.Join(dateParts, ", ")))
		}
	}

	// Add tag
	if task.Tag != "" {
		parts = append(parts, "#"+task.Tag)
	}

	// Add project
	if task.Project != "" {
		parts = append(parts, "+"+task.Project)
	}

	// Add priority
	if task.Priority != "" {
		parts = append(parts, "!"+task.Priority)
	}

	// Add note
	if task.Note != "" {
		parts = append(parts, fmt.Sprintf("*'%s'", task.Note))
	}

	return strings.Join(parts, " ")
}

// FormatTasksForEditor formats multiple tasks for editing (with IDs)
func FormatTasksForEditor(tasks []models.Task) string {
	var lines []string
	for _, task := range tasks {
		// Format: #ID<tab>task content
		line := fmt.Sprintf("#%d\t%s", task.ID, FormatTaskForEditor(&task))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ParseEditedTask parses a single edited line (without ID prefix)
func ParseEditedTask(line string) (*models.Task, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}
	return parser.ParseTaskString(line)
}

// ParseEditedTasks parses multiple edited lines (with ID prefixes)
// Returns map[taskID]*editedTask and list of errors
func ParseEditedTasks(content string, originalTasks []models.Task) (map[int]*models.Task, []error) {
	lines := strings.Split(content, "\n")
	result := make(map[int]*models.Task)
	var errors []error

	// Create map of original tasks by ID for validation
	originalMap := make(map[int]models.Task)
	for _, task := range originalTasks {
		originalMap[task.ID] = task
	}

	lineNum := 0
	for _, line := range lines {
		lineNum++
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Must start with #<number>
		if !strings.HasPrefix(line, "#") {
			errors = append(errors, fmt.Errorf("line %d: missing task ID prefix (should start with #<number>)", lineNum))
			continue
		}

		// Extract ID and content
		parts := strings.SplitN(line[1:], "\t", 2)
		if len(parts) < 2 {
			// Try splitting by space if tab doesn't work
			parts = strings.SplitN(line[1:], " ", 2)
			if len(parts) < 2 {
				errors = append(errors, fmt.Errorf("line %d: invalid format (expected: #<id><tab>content)", lineNum))
				continue
			}
		}

		// Parse ID
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			errors = append(errors, fmt.Errorf("line %d: invalid task ID '%s'", lineNum, parts[0]))
			continue
		}

		// Validate ID exists in original tasks
		if _, exists := originalMap[id]; !exists {
			errors = append(errors, fmt.Errorf("line %d: task ID %d not found in original task list", lineNum, id))
			continue
		}

		// Parse task content
		content := strings.TrimSpace(parts[1])
		task, err := ParseEditedTask(content)
		if err != nil {
			errors = append(errors, fmt.Errorf("line %d (task #%d): %v", lineNum, id, err))
			continue
		}

		task.ID = id
		result[id] = task
	}

	// Check for missing tasks (deleted lines)
	for _, original := range originalTasks {
		if _, found := result[original.ID]; !found {
			errors = append(errors, fmt.Errorf("task #%d was removed from editor (line deletion not allowed)", original.ID))
		}
	}

	return result, errors
}

// HasChanges checks if edited task differs from original
func HasChanges(original, edited *models.Task) bool {
	if original.Name != edited.Name {
		return true
	}

	if !datesEqual(original.DueDate, edited.DueDate) {
		return true
	}

	if original.RecurFrequency != edited.RecurFrequency {
		return true
	}

	if !datesEqual(original.RecurEndDate, edited.RecurEndDate) {
		return true
	}

	if original.Tag != edited.Tag {
		return true
	}

	if original.Project != edited.Project {
		return true
	}

	if original.Priority != edited.Priority {
		return true
	}

	if original.Note != edited.Note {
		return true
	}

	return false
}

func datesEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
