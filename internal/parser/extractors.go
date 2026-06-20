package parser

import (
	"regexp"
	"strings"

	"github.com/ishrq/recur/internal/models"
)

// extractDueDate extracts the @(...) date expression from input
func extractDueDate(input string, task *models.Task) (string, error) {
	re := regexp.MustCompile(`@\(([^)]+)\)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		dateStr := matches[1]
		if err := parseDueDateString(dateStr, task); err != nil {
			return input, err
		}
		return re.ReplaceAllString(input, ""), nil
	}

	return input, nil
}

// extractNote extracts the *'...' or *"..." note from input
func extractNote(input string, task *models.Task) string {
	// Match *'...' or *"..." with support for escaped quotes
	re := regexp.MustCompile(`\*['"]([^'"]*(?:\\['"][^'"]*)*)['"]`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		// Unescape any escaped quotes
		note := matches[1]
		note = strings.ReplaceAll(note, `\'`, `'`)
		note = strings.ReplaceAll(note, `\"`, `"`)
		task.Note = note
		return re.ReplaceAllString(input, "")
	}

	return input
}

// extractTags extracts #tag(s) from input, storing them comma-separated
func extractTags(input string, task *models.Task) string {
	re := regexp.MustCompile(`#([\w-]+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) > 0 {
		tags := make([]string, len(matches))
		for i, m := range matches {
			tags[i] = m[1]
		}
		task.Tag = strings.Join(tags, ",")
		return re.ReplaceAllString(input, "")
	}

	return input
}

// extractProject extracts the +project from input
func extractProject(input string, task *models.Task) string {
	// Match +word (alphanumeric, underscore, hyphen)
	re := regexp.MustCompile(`\+([\w-]+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Project = matches[1]
		return strings.Replace(input, matches[0], "", 1)
	}

	return input
}

// extractPriority extracts the !priority from input
func extractPriority(input string, task *models.Task) string {
	// Match !word (alphanumeric, underscore, hyphen)
	re := regexp.MustCompile(`!([\w-]+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		task.Priority = matches[1]
		return strings.Replace(input, matches[0], "", 1)
	}

	return input
}
