package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/models"
)

func List(database *sql.DB, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			PrintHelp("ls")
			return nil
		}
	}

	var showAll bool
	var showToday bool
	var showTomorrow bool
	var showOverdue bool
	var showUpcoming bool
	var dueDate string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all", "-a":
			showAll = true
		case "--today":
			showToday = true
		case "--tomorrow":
			showTomorrow = true
		case "--overdue":
			showOverdue = true
		case "--upcoming":
			showUpcoming = true
		case "--due", "-d":
			if i+1 < len(args) {
				dueDate = args[i+1]
				i++
			}
		}
	}

	showDashboard := !showAll && !showToday && !showTomorrow && !showOverdue && !showUpcoming && dueDate == ""

	if showDashboard {
		return displayDashboard(database)
	}

	tasks, err := getFilteredTasks(database, showAll, showToday, showTomorrow, showOverdue, showUpcoming, dueDate)
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

func displayDashboard(database *sql.DB) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)

	// Get all incomplete tasks
	allTasks, err := db.GetTasks(database, false)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// Categorize tasks
	var overdue, todayTasks, tomorrowTasks, upcoming, noDueDate []models.Task

	for _, task := range allTasks {
		if task.DueDate == nil {
			noDueDate = append(noDueDate, task)
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())

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
		printTasksCompact(overdue)
		fmt.Println()
	}

	if len(todayTasks) > 0 {
		fmt.Printf("═══ TODAY (%d) ═══\n", len(todayTasks))
		printTasksCompact(todayTasks)
		fmt.Println()
	}

	if len(tomorrowTasks) > 0 {
		fmt.Printf("═══ TOMORROW (%d) ═══\n", len(tomorrowTasks))
		printTasksCompact(tomorrowTasks)
		fmt.Println()
	}

	if len(upcoming) > 0 {
		fmt.Printf("═══ UPCOMING (%d) ═══\n", len(upcoming))
		printTasksCompact(upcoming)
		fmt.Println()
	}

	if len(noDueDate) > 0 {
		fmt.Printf("═══ NO DUE DATE (%d) ═══\n", len(noDueDate))
		printTasksCompact(noDueDate)
		fmt.Println()
	}

	if len(allTasks) == 0 {
		fmt.Println("No tasks found. Add one with: recur add \"Task name\"")
		fmt.Println()
	}

	return nil
}

func getFilteredTasks(database *sql.DB, showAll, showToday, showTomorrow, showOverdue, showUpcoming bool, dueDate string) ([]models.Task, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextWeek := today.AddDate(0, 0, 7)

	allTasks, err := db.GetTasks(database, showAll)
	if err != nil {
		return nil, err
	}

	if dueDate != "" {
		return filterByDate(allTasks, dueDate)
	}

	var filtered []models.Task
	for _, task := range allTasks {
		if task.DueDate == nil {
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, now.Location())

		if showOverdue && taskDate.Before(today) {
			filtered = append(filtered, task)
		} else if showToday && taskDate.Equal(today) {
			filtered = append(filtered, task)
		} else if showTomorrow && taskDate.Equal(today.AddDate(0, 0, 1)) {
			filtered = append(filtered, task)
		} else if showUpcoming && taskDate.After(today) && taskDate.Before(nextWeek.AddDate(0, 0, 1)) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func filterByDate(tasks []models.Task, dateStr string) ([]models.Task, error) {
	// Handle special keywords
	now := time.Now()
	var targetDate time.Time

	switch strings.ToLower(dateStr) {
	case "today":
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "tomorrow":
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	case "none":
		// Return tasks with no due date
		var filtered []models.Task
		for _, task := range tasks {
			if task.DueDate == nil {
				filtered = append(filtered, task)
			}
		}
		return filtered, nil
	case "recurring":
		// Return recurring tasks
		var filtered []models.Task
		for _, task := range tasks {
			if task.RecurFrequency != "" {
				filtered = append(filtered, task)
			}
		}
		return filtered, nil
	default:
		// Try parsing as date
		formats := []string{
			"2006-01-02",
			"Jan 2",
			"Jan 2 2006",
			"Monday",
			"Mon",
		}

		parsed := false
		for _, format := range formats {
			if t, err := time.Parse(format, dateStr); err == nil {
				if t.Year() == 0 {
					t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
				}
				targetDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				parsed = true
				break
			}
		}

		if !parsed {
			return nil, fmt.Errorf("invalid date format: %s", dateStr)
		}
	}

	// Filter tasks by date
	var filtered []models.Task
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		taskDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())
		if taskDate.Equal(targetDate) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func printTasks(tasks []models.Task) {
	fmt.Println()
	printTasksCompact(tasks)
	fmt.Println()
}

func printTasksCompact(tasks []models.Task) {
	for _, task := range tasks {
		// Format: #ID Task Name [due date]
		line := fmt.Sprintf("#%-4d %s", task.ID, task.Name)

		if task.DueDate != nil {
			dueStr := formatDueDate(*task.DueDate)
			line += fmt.Sprintf(" [%s]", dueStr)
		}

		if task.RecurFrequency != "" {
			line += fmt.Sprintf(" ↻ %s", task.RecurFrequency)
		}

		var metadata []string
		if task.Project != "" {
			metadata = append(metadata, "+"+task.Project)
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
