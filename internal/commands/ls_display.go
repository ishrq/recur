package commands

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

const (
	defaultTimeHour   = 12
	defaultTimeMinute = 0
	idColumnWidth     = 8
	dueDateColumnWidth = 15
)

func displayDashboard(database *sql.DB, showNote bool) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)

	// Get all incomplete tasks
	allTasks, err := db.GetTasks(database, false, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	var overdue, todayTasks, tomorrowTasks, upcoming, noDueDate []models.Task

	for _, task := range allTasks {
		if task.DueDate == nil {
			noDueDate = append(noDueDate, task)
			continue
		}

		// Use the same location as 'today' for comparison
		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, now.Location())

		if taskDate.Before(today) {
			overdue = append(overdue, task)
		} else if taskDate.Equal(today) {
			todayTasks = append(todayTasks, task)
		} else if taskDate.Equal(tomorrow) {
			tomorrowTasks = append(tomorrowTasks, task)
		} else if taskDate.After(tomorrow) && taskDate.Before(nextWeek.AddDate(0, 0, 1)) {
			upcoming = append(upcoming, task)
		}
	}

	fmt.Println()
	if len(overdue) > 0 {
		fmt.Printf("═══ OVERDUE (%d) ═══\n", len(overdue))
		printTasksTabular(overdue, showNote)
		fmt.Println()
	}

	if len(todayTasks) > 0 {
		fmt.Printf("═══ TODAY (%d) ═══\n", len(todayTasks))
		printTasksTabular(todayTasks, showNote)
		fmt.Println()
	}

	if len(tomorrowTasks) > 0 {
		fmt.Printf("═══ TOMORROW (%d) ═══\n", len(tomorrowTasks))
		printTasksTabular(tomorrowTasks, showNote)
		fmt.Println()
	}

	if len(upcoming) > 0 {
		fmt.Printf("═══ UPCOMING (%d) ═══\n", len(upcoming))
		printTasksTabular(upcoming, showNote)
		fmt.Println()
	}

	if len(noDueDate) > 0 {
		fmt.Printf("═══ NO DUE DATE (%d) ═══\n", len(noDueDate))
		printTasksTabular(noDueDate, showNote)
		fmt.Println()
	}

	if len(allTasks) == 0 {
		fmt.Println("No tasks found. Add one with: recur add \"Task name\"")
		fmt.Println()
	}

	return nil
}

func displayTags(database *sql.DB) error {
	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	tagCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Tag != "" {
			tagCounts[task.Tag]++
		}
	}

	if len(tagCounts) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	tags := make([]string, 0, len(tagCounts))
	for tag := range tagCounts {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	fmt.Println()
	fmt.Println("Tags:")
	fmt.Println()
	for _, tag := range tags {
		fmt.Printf("  #%-20s (%d)\n", tag, tagCounts[tag])
	}
	fmt.Println()
	fmt.Printf("Total: %d tags\n", len(tags))

	return nil
}

func displayProjects(database *sql.DB) error {
	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	projectCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Project != "" {
			projectCounts[task.Project]++
		}
	}

	if len(projectCounts) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	projects := make([]string, 0, len(projectCounts))
	for project := range projectCounts {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	fmt.Println()
	fmt.Println("Projects:")
	fmt.Println()
	for _, project := range projects {
		fmt.Printf("  +%-20s (%d)\n", project, projectCounts[project])
	}
	fmt.Println()
	fmt.Printf("Total: %d projects\n", len(projects))

	return nil
}

func displayPriorities(database *sql.DB) error {
	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	priorityCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Priority != "" {
			priorityCounts[task.Priority]++
		}
	}

	if len(priorityCounts) == 0 {
		fmt.Println("No priorities found.")
		return nil
	}

	priorities := make([]string, 0, len(priorityCounts))
	for priority := range priorityCounts {
		priorities = append(priorities, priority)
	}

	sort.Slice(priorities, func(i, j int) bool {
		order := map[string]int{
			"urgent": 1,
			"high":   2,
			"medium": 3,
			"low":    4,
		}

		iOrder, iExists := order[strings.ToLower(priorities[i])]
		jOrder, jExists := order[strings.ToLower(priorities[j])]

		if iExists && jExists {
			return iOrder < jOrder
		}
		if iExists {
			return true
		}
		if jExists {
			return false
		}
		return priorities[i] < priorities[j]
	})

	fmt.Println()
	fmt.Println("Priorities:")
	fmt.Println()
	for _, priority := range priorities {
		fmt.Printf("  !%-20s (%d)\n", priority, priorityCounts[priority])
	}
	fmt.Println()
	fmt.Printf("Total: %d priorities\n", len(priorities))

	return nil
}

func printTasks(tasks []models.Task, showNote bool) {
	fmt.Println()
	printTasksTabular(tasks, showNote)
	fmt.Println()
}

func printTasksTabular(tasks []models.Task, showNote bool) {
	// Print header
	fmt.Printf("%-*s %-*s Task Name\n", idColumnWidth, "ID", dueDateColumnWidth, "Due Date")
	fmt.Println(strings.Repeat("─", 70))

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, task := range tasks {
		// First line: ID, Due Date, Task Name
		idStr := fmt.Sprintf("#%d", task.ID)

		// Add note indicator to task name
		taskName := task.Name
		if task.Note != "" {
			taskName += "*"
		}

		dueDateStr, recurStr := formatDueDateCompact(task.DueDate, task.RecurFrequency, today)

		fmt.Printf("%-*s %-*s %s\n", idColumnWidth, idStr, dueDateColumnWidth, dueDateStr, taskName)

		// Second line: Recurrence (if on separate line) and/or Metadata
		hasRecurLine := recurStr != "" && dueDateStr == ""
		metadata := formatMetadata(task)

		if hasRecurLine && metadata != "" {
			// Both recurrence and metadata
			fmt.Printf("%-*s %-*s %s\n", idColumnWidth, "", dueDateColumnWidth, recurStr, metadata)
		} else if hasRecurLine {
			// Only recurrence
			fmt.Printf("%-*s %s\n", idColumnWidth, "", recurStr)
		} else if metadata != "" {
			// Only metadata (recurrence was shown with due date)
			fmt.Printf("%-*s %-*s %s\n", idColumnWidth, "", dueDateColumnWidth, "", metadata)
		}

		// Third line: Note (if --note flag is set)
		if showNote && task.Note != "" {
			fmt.Printf("%-*s %-*s Note: %s\n", idColumnWidth, "", dueDateColumnWidth, "", task.Note)
		}
	}
}

// formatDueDateCompact formats the due date according to display rules
// Returns: (dueDateStr, recurStr)
// - dueDateStr: the formatted due date for the main line
// - recurStr: recurring indicator (may be empty if shown with date, or needs separate line)
func formatDueDateCompact(dueDate *time.Time, recurFreq string, today time.Time) (string, string) {
	if dueDate == nil {
		// No due date - return recurring on separate line if exists
		if recurFreq != "" {
			return "", formatRecurrence(recurFreq)
		}
		return "", ""
	}

	taskDate := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	tomorrow := today.AddDate(0, 0, 1)

	hasDefaultTime := isDefaultTime(*dueDate)
	recurIndicator := ""
	if recurFreq != "" {
		recurIndicator = formatRecurrence(recurFreq)
	}

	// Today
	if taskDate.Equal(today) {
		if hasDefaultTime {
			// Today with default time - skip date entirely, show recur if exists
			return recurIndicator, ""
		}
		// Today with custom time
		timeStr := dueDate.Format("15:04")
		if recurIndicator != "" {
			return timeStr, recurIndicator // Recur on separate line
		}
		return timeStr, ""
	}

	// Tomorrow
	if taskDate.Equal(tomorrow) {
		if hasDefaultTime {
			if recurIndicator != "" {
				return "Tomorrow", recurIndicator
			}
			return "Tomorrow", ""
		}
		timeStr := dueDate.Format("15:04")
		result := fmt.Sprintf("Tomorrow %s", timeStr)
		return result, recurIndicator
	}

	// Overdue
	if taskDate.Before(today) {
		dateStr := dueDate.Format("Jan 2")
		result := fmt.Sprintf("Overdue: %s", dateStr)
		return result, recurIndicator
	}

	// Future dates
	if hasDefaultTime {
		// Just the date
		dateStr := dueDate.Format("Jan 2")
		if recurIndicator != "" {
			return dateStr, recurIndicator
		}
		return dateStr, ""
	}

	// Future with time
	dateStr := dueDate.Format("Jan 2 15:04")
	return dateStr, recurIndicator
}

// formatRecurrence converts frequency string to compact display format
func formatRecurrence(freq string) string {
	if freq == "" {
		return ""
	}

	// Handle semantic frequencies
	switch strings.ToLower(freq) {
	case "hourly":
		return "↻ H"
	case "daily":
		return "↻ D"
	case "weekly":
		return "↻ W"
	case "monthly":
		return "↻ M"
	case "yearly":
		return "↻ Y"
	}

	// Handle numeric frequencies: 1d, 2w, 3m, etc.
	freq = strings.ToLower(freq)
	if len(freq) >= 2 {
		// Extract number and unit
		numPart := freq[:len(freq)-1]
		unit := freq[len(freq)-1:]

		unitMap := map[string]string{
			"h": "H",
			"d": "D",
			"w": "W",
			"m": "M",
			"y": "Y",
		}

		if unitSymbol, ok := unitMap[unit]; ok {
			// If number is 1, just show unit (↻D instead of ↻1D)
			if numPart == "1" {
				return "↻" + unitSymbol
			}
			return "↻" + numPart + unitSymbol
		}
	}

	// Fallback: show as-is with recurring symbol
	return "↻" + freq
}

// formatMetadata formats tags, projects, and priorities
func formatMetadata(task models.Task) string {
	var parts []string

	if task.Tag != "" {
		parts = append(parts, "#"+task.Tag)
	}
	if task.Project != "" {
		parts = append(parts, "+"+task.Project)
	}
	if task.Priority != "" {
		parts = append(parts, "!"+task.Priority)
	}

	return strings.Join(parts, " ")
}

// isDefaultTime checks if a time has the default time (12:00:00)
func isDefaultTime(t time.Time) bool {
	return t.Hour() == defaultTimeHour && t.Minute() == defaultTimeMinute && t.Second() == 0
}

// Kept for backwards compatibility, but not used in new tabular display
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
