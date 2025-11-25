package commands

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

const (
	defaultTimeHour    = 12
	defaultTimeMinute  = 0
	idColumnWidth      = 4
	dueDateColumnWidth = 15
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

var colorsEnabled = true

func init() {
	colorsEnabled = supportsColor()
}

func supportsColor() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

func color(code string) string {
	if !colorsEnabled {
		return ""
	}
	return code
}

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
		fmt.Printf("%s%sOVERDUE (%d)%s\n", color(Bold), color(Red), len(overdue), color(Reset))
		printTaskList(overdue, showNote, today)
		fmt.Println()
	}

	if len(todayTasks) > 0 {
		fmt.Printf("%s%sTODAY (%d)%s\n", color(Bold), color(Green), len(todayTasks), color(Reset))
		printTaskList(todayTasks, showNote, today)
		fmt.Println()
	}

	if len(tomorrowTasks) > 0 {
		fmt.Printf("%s%sTOMORROW (%d)%s\n", color(Bold), color(Yellow), len(tomorrowTasks), color(Reset))
		printTaskList(tomorrowTasks, showNote, today)
		fmt.Println()
	}

	if len(upcoming) > 0 {
		fmt.Printf("%sUPCOMING (%d)%s\n", color(Bold), len(upcoming), color(Reset))
		printTaskList(upcoming, showNote, today)
		fmt.Println()
	}

	if len(noDueDate) > 0 {
		fmt.Printf("%sNO DUES (%d)%s\n", color(Bold), len(noDueDate), color(Reset))
		printTaskList(noDueDate, showNote, today)
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
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	fmt.Println()
	printTaskList(tasks, showNote, today)
	fmt.Println()
}

func printTaskList(tasks []models.Task, showNote bool, today time.Time) {
	fmt.Println(strings.Repeat("─", 5))

	for _, task := range tasks {
		// First line: ID, Due Date, Task Name
		idStr := fmt.Sprintf("#%d", task.ID)

		// Add note indicator to task name
		taskName := color(Bold) + task.Name
		if task.Note != "" && !showNote {
			taskName += color(Red) + "*" + color(Reset)
		}

		dueDateStr, recurStr := formatDueDateCompact(task.DueDate, task.RecurFrequency, today)

		// Dim ID, normal date, normal task name
		fmt.Printf("%s%-*s%s %-*s %s\n",
			color(Dim), idColumnWidth, idStr, color(Reset),
			dueDateColumnWidth, dueDateStr,
			taskName)

		// Second line: Recurrence (if on separate line) and/or Metadata
		hasRecurLine := recurStr != "" && dueDateStr == ""
		metadata := formatMetadata(task)

		if hasRecurLine && metadata != "" {
			fmt.Printf("%-*s %-*s   %s%s%s\n",
				idColumnWidth, "",
				dueDateColumnWidth, recurStr,
				color(Dim), metadata, color(Reset))
		} else if hasRecurLine {
			fmt.Printf("%-*s %s\n", idColumnWidth, "", recurStr)
		} else if metadata != "" {
			fmt.Printf("%-*s %-*s   %s%s%s\n",
				idColumnWidth, "",
				dueDateColumnWidth, "",
				color(Dim), metadata, color(Reset))
		}

		// Third line: Note (if --note flag is set)
		if showNote && task.Note != "" {
			fmt.Printf("%-*s %-*s   %s%s*%s%s\n",
				idColumnWidth, "",
				dueDateColumnWidth, "",
				color(Red),
				color(Dim), task.Note, color(Reset))
		}
	}
}

// formatDueDateCompact formats the due date according to display rules
// Returns: (dueDateStr, recurStr)
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

	// Today - show date for consistency in dashboard
	if taskDate.Equal(today) {
		if hasDefaultTime {
			dateStr := dueDate.Format("Jan 02")
			if recurIndicator != "" {
				return dateStr, recurIndicator
			}
			return dateStr, ""
		}
		// Today with custom time
		dateStr := dueDate.Format("Jan 02 15:04")
		return dateStr, recurIndicator
	}

	// Tomorrow - show date for consistency
	if taskDate.Equal(tomorrow) {
		if hasDefaultTime {
			dateStr := dueDate.Format("Jan 02")
			if recurIndicator != "" {
				return dateStr, recurIndicator
			}
			return dateStr, ""
		}
		dateStr := dueDate.Format("Jan 02 15:04")
		return dateStr, recurIndicator
	}

	// Overdue
	if taskDate.Before(today) {
		dateStr := dueDate.Format("Jan 02")
		// Add time if not default
		if !hasDefaultTime {
			dateStr = dueDate.Format("Jan 02 15:04")
		}
		return dateStr, recurIndicator
	}

	// Future dates
	if hasDefaultTime {
		dateStr := dueDate.Format("Jan 02")
		if recurIndicator != "" {
			return dateStr, recurIndicator
		}
		return dateStr, ""
	}

	// Future with time
	dateStr := dueDate.Format("Jan 02 15:04")
	return dateStr, recurIndicator
}

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
			if numPart == "1" {
				return "↻ " + unitSymbol
			}
			return "↻ " + numPart + unitSymbol
		}
	}

	return "↻ " + freq
}

func formatMetadata(task models.Task) string {
	var parts []string

	if task.Tag != "" {
		parts = append(parts, "#"+task.Tag)
	}
	if task.Project != "" {
		parts = append(parts, "+"+task.Project)
	}

	// Show urgent (!!) and high (!)
	if task.Priority != "" {
		priority := strings.ToLower(task.Priority)
		switch priority {
		case "urgent":
			parts = append(parts, color(Reset)+color(Red)+"!!"+color(Reset)+color(Dim))
		case "high":
			parts = append(parts, color(Reset)+color(Yellow)+"!"+color(Reset)+color(Dim))
		}
	}

	return strings.Join(parts, " ")
}

func isDefaultTime(t time.Time) bool {
	return t.Hour() == defaultTimeHour && t.Minute() == defaultTimeMinute && t.Second() == 0
}
