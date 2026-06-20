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
	defaultTimeHour   = 12
	defaultTimeMinute = 0
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

		local := task.DueDate.Local()
		taskDate := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, now.Location())

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
	}

	if len(todayTasks) > 0 {
		fmt.Println()
		fmt.Printf("%s%sTODAY (%d)%s\n", color(Bold), color(Green), len(todayTasks), color(Reset))
		printTaskList(todayTasks, showNote, today)
	}

	if len(tomorrowTasks) > 0 {
		fmt.Println()
		fmt.Printf("%s%sTOMORROW (%d)%s\n", color(Bold), color(Yellow), len(tomorrowTasks), color(Reset))
		printTaskList(tomorrowTasks, showNote, today)
	}

	if len(upcoming) > 0 {
		fmt.Println()
		fmt.Printf("%sUPCOMING (%d)%s\n", color(Bold), len(upcoming), color(Reset))
		printTaskList(upcoming, showNote, today)
	}

	if len(noDueDate) > 0 {
		fmt.Println()
		fmt.Printf("%sNO DUES (%d)%s\n", color(Bold), len(noDueDate), color(Reset))
		printTaskList(noDueDate, showNote, today)
	}

	if len(allTasks) == 0 {
		fmt.Println("No tasks found. Add one with: recur add \"Task name\"")
		fmt.Println()
	}

	return nil
}

func displayCounts(database *sql.DB, fieldFn func(*models.Task) []string, prefix, label string, lessFn func(a, b string) bool) error {
	tasks, err := db.GetTasks(database, false, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	counts := make(map[string]int)
	for _, task := range tasks {
		for _, v := range fieldFn(&task) {
			if v != "" {
				counts[v]++
			}
		}
	}

	if len(counts) == 0 {
		fmt.Printf("No %s found.\n", strings.ToLower(label))
		return nil
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}

	if lessFn != nil {
		sort.Slice(keys, func(i, j int) bool { return lessFn(keys[i], keys[j]) })
	} else {
		sort.Strings(keys)
	}

	fmt.Println()
	fmt.Printf("%s:\n", label)
	fmt.Println()
	for _, k := range keys {
		fmt.Printf("  %s%-20s (%d)\n", prefix, k, counts[k])
	}
	fmt.Println()
	fmt.Printf("Total: %d %s\n", len(keys), strings.ToLower(label))

	return nil
}

func displayTags(database *sql.DB) error {
	return displayCounts(database, func(t *models.Task) []string {
		if t.Tag == "" {
			return nil
		}
		return strings.Split(t.Tag, ",")
	}, "#", "Tags", nil)
}

func displayProjects(database *sql.DB) error {
	return displayCounts(database, func(t *models.Task) []string {
		if t.Project == "" {
			return nil
		}
		return []string{t.Project}
	}, "+", "Projects", nil)
}

func priorityLess(a, b string) bool {
	order := map[string]int{"urgent": 1, "high": 2}
	aOrder, aOk := order[strings.ToLower(a)]
	bOrder, bOk := order[strings.ToLower(b)]

	if aOk && bOk {
		return aOrder < bOrder
	}
	if aOk {
		return true
	}
	if bOk {
		return false
	}
	return a < b
}

func displayPriorities(database *sql.DB) error {
	return displayCounts(database, func(t *models.Task) []string {
		if t.Priority == "" {
			return nil
		}
		return []string{t.Priority}
	}, "!", "Priorities", priorityLess)
}

func printTasks(tasks []models.Task, showNote bool) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	fmt.Println()
	printTaskList(tasks, showNote, today)
}

func printTaskList(tasks []models.Task, showNote bool, today time.Time) {
	if len(tasks) == 0 {
		return
	}

	// Calculate dynamic column widths
	maxIDWidth := 2 // Minimum for "#1"
	maxDateWidth := 0

	for _, task := range tasks {
		idLen := len(fmt.Sprintf("#%d", task.ID))
		if idLen > maxIDWidth {
			maxIDWidth = idLen
		}

		dateStr, recurStr := formatDueDateCompact(task.DueDate, task.RecurFrequency, today)
		dateLen := len(dateStr)
		if recurStr != "" {
			if dateLen > 0 {
				dateLen += len(recurStr) + 2 // +2 for "  " spaces
			} else {
				dateLen = len(recurStr)
			}
		}
		if dateLen > maxDateWidth {
			maxDateWidth = dateLen
		}
	}

	// Add padding
	idColWidth := maxIDWidth + 1
	dateColWidth := maxDateWidth

	fmt.Println(strings.Repeat("─", 5))

	for _, task := range tasks {
		// First line: ID, Due Date + Recurring, Task Name
		idStr := fmt.Sprintf("#%d", task.ID)

		// Add note indicator to task name
		taskName := color(Bold) + task.Name + color(Reset)
		if task.Note != "" && !showNote {
			taskName += color(Red) + "*" + color(Reset)
		}

		dueDateStr, recurStr := formatDueDateCompact(task.DueDate, task.RecurFrequency, today)

		// Build date + recurring field
		dateField := dueDateStr
		if recurStr != "" {
			if dateField != "" {
				dateField += "  " + recurStr
			} else {
				dateField = recurStr
			}
		}

		// Print first line
		fmt.Printf("%s%-*s%s %-*s  %s\n",
			color(Dim), idColWidth, idStr, color(Reset),
			dateColWidth, dateField,
			taskName)

		// Second line: Metadata
		metadata := formatMetadata(task)
		if metadata != "" {
			fmt.Printf("%-*s %-*s  %s%s%s\n",
				idColWidth, "",
				dateColWidth, "",
				color(Dim), metadata, color(Reset))
		}

		// Third line: Note (if --note flag is set)
		if showNote && task.Note != "" {
			fmt.Printf("%-*s %-*s  %s*%s %s%s%s\n",
				idColWidth, "",
				dateColWidth, "",
				color(Red), color(Reset),
				color(Dim), task.Note, color(Reset))
		}
	}
}

// formatDueDateCompact formats the due date according to display rules
// Returns: (dueDateStr, recurStr)
func formatDueDateCompact(dueDate *time.Time, recurFreq string, today time.Time) (string, string) {
	recurIndicator := ""
	if recurFreq != "" {
		recurIndicator = formatRecurrence(recurFreq)
	}

	if dueDate == nil {
		return "", recurIndicator
	}

	localDueDate := dueDate.Local()
	currentYear := today.Year()
	currentMonth := today.Month()

	hasDefaultTime := isDefaultTime(localDueDate)

	// Format with day of week
	var dateStr string

	// Same month and year: just day of week and day
	if localDueDate.Year() == currentYear && localDueDate.Month() == currentMonth {
		if hasDefaultTime {
			dateStr = localDueDate.Format("Mon 02")
		} else {
			dateStr = localDueDate.Format("Mon 02 15:04")
		}
	} else if localDueDate.Year() == currentYear {
		// Same year but different month: include month
		if hasDefaultTime {
			dateStr = localDueDate.Format("Mon 02 Jan")
		} else {
			dateStr = localDueDate.Format("Mon 02 Jan 15:04")
		}
	} else {
		// Different year: include full year
		if hasDefaultTime {
			dateStr = localDueDate.Format("Mon 02 Jan 2006")
		} else {
			// With time AND different year
			dateStr = localDueDate.Format("Mon 02 Jan 2006 15:04")
		}
	}

	return dateStr, recurIndicator
}

func formatRecurrence(freq string) string {
	if freq == "" {
		return ""
	}

	// Handle semantic frequencies
	switch strings.ToLower(freq) {
	case "hourly":
		return "↻H"
	case "daily":
		return "↻D"
	case "weekly":
		return "↻W"
	case "monthly":
		return "↻M"
	case "yearly":
		return "↻Y"
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
				return "↻" + unitSymbol
			}
			return "↻" + numPart + unitSymbol
		}
	}

	return "↻" + freq
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
