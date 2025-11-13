package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

func List(database *sql.DB, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			PrintHelp("ls")
			return nil
		}
	}

	var showAll bool
	var showDone bool
	var showToday bool
	var showTomorrow bool
	var showOverdue bool
	var showUpcoming bool
	var dueDate string
	var fromDate string
	var toDate string
	var query string
	var tags []string
	var projects []string
	var priorities []string
	var listTags bool
	var listProjects bool
	var listPriorities bool
	var showNote bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all", "-a":
			showAll = true
		case "--done", "-x":
			showDone = true
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
		case "--from":
			if i+1 < len(args) {
				fromDate = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				toDate = args[i+1]
				i++
			}
		case "--query", "-q":
			if i+1 < len(args) {
				query = args[i+1]
				i++
			}
		case "--tags":
			listTags = true
		case "--projects":
			listProjects = true
		case "--priorities":
			listPriorities = true
		case "--note", "-n":
			showNote = true
		case "--tag", "-t":
			// Collect all following non-flag arguments as tags
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				tags = append(tags, args[i+1])
				i++
			}
		case "--project", "-p":
			// Collect all following non-flag arguments as projects
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				projects = append(projects, args[i+1])
				i++
			}
		case "--priority", "-P":
			// Collect all following non-flag arguments as priorities
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				priorities = append(priorities, args[i+1])
				i++
			}
		}
	}

	// Check for invalid flag combinations
	listingFlags := []bool{listTags, listProjects, listPriorities}
	listingCount := 0
	for _, flag := range listingFlags {
		if flag {
			listingCount++
		}
	}

	if listingCount > 1 {
		return fmt.Errorf("cannot combine --tags, --projects, and --priorities flags")
	}

	if listingCount > 0 {
		// Check if combined with other flags
		hasOtherFlags := showAll || showToday || showTomorrow || showOverdue || showUpcoming ||
		dueDate != "" || fromDate != "" || toDate != "" ||
		len(tags) > 0 || len(projects) > 0 || len(priorities) > 0 || showNote

		if hasOtherFlags {
			return fmt.Errorf("--tags, --projects, and --priorities cannot be combined with other filters")
		}

		// Handle listing commands
		if listTags {
			return displayTags(database)
		}
		if listProjects {
			return displayProjects(database)
		}
		if listPriorities {
			return displayPriorities(database)
		}
	}

	showDashboard := !showAll && !showDone && !showToday && !showTomorrow && !showOverdue && !showUpcoming &&
		dueDate == "" && fromDate == "" && toDate == "" && query == "" &&
		len(tags) == 0 && len(projects) == 0 && len(priorities) == 0

	if showDashboard {
		return displayDashboard(database, showNote)
	}

	tasks, err := getFilteredTasks(database, showAll, showDone, showToday, showTomorrow, showOverdue, showUpcoming,
		dueDate, fromDate, toDate, query, tags, projects, priorities)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	printTasks(tasks, showNote)
	return nil
}
