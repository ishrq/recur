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

func displayDashboard(database *sql.DB, showNote bool) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)

	// Get all incomplete tasks
	allTasks, err := db.GetTasks(database, false)
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
		printTasksCompact(overdue, showNote)
		fmt.Println()
	}

	if len(todayTasks) > 0 {
		fmt.Printf("═══ TODAY (%d) ═══\n", len(todayTasks))
		printTasksCompact(todayTasks, showNote)
		fmt.Println()
	}

	if len(tomorrowTasks) > 0 {
		fmt.Printf("═══ TOMORROW (%d) ═══\n", len(tomorrowTasks))
		printTasksCompact(tomorrowTasks, showNote)
		fmt.Println()
	}

	if len(upcoming) > 0 {
		fmt.Printf("═══ UPCOMING (%d) ═══\n", len(upcoming))
		printTasksCompact(upcoming, showNote)
		fmt.Println()
	}

	if len(noDueDate) > 0 {
		fmt.Printf("═══ NO DUE DATE (%d) ═══\n", len(noDueDate))
		printTasksCompact(noDueDate, showNote)
		fmt.Println()
	}

	if len(allTasks) == 0 {
		fmt.Println("No tasks found. Add one with: recur add \"Task name\"")
		fmt.Println()
	}

	return nil
}

func displayTags(database *sql.DB) error {
	tasks, err := db.GetTasks(database, false)
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
	tasks, err := db.GetTasks(database, false)
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
	tasks, err := db.GetTasks(database, false)
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
	printTasksCompact(tasks, showNote)
	fmt.Println()
}

func printTasksCompact(tasks []models.Task, showNote bool) {
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

		if showNote && task.Note != "" {
			fmt.Printf("      Note: %s\n", task.Note)
		}
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
