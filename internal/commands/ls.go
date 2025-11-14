package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/filter"
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
	var showDone bool
	var showTrash bool
	var showNote bool
	var listTags bool
	var listProjects bool
	var listPriorities bool

	filters := filter.Filters{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all", "-a":
			showAll = true
		case "--done", "-x":
			showDone = true
		case "--trash":
			showTrash = true
		case "--today":
			filters.Today = true
		case "--tomorrow":
			filters.Tomorrow = true
		case "--overdue":
			filters.Overdue = true
		case "--upcoming":
			filters.Upcoming = true
		case "--due", "-d":
			if i+1 < len(args) {
				filters.DueDate = args[i+1]
				i++
			}
		case "--from":
			if i+1 < len(args) {
				filters.FromDate = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				filters.ToDate = args[i+1]
				i++
			}
		case "--query", "-q":
			if i+1 < len(args) {
				filters.Query = args[i+1]
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
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Tags = append(filters.Tags, args[i+1])
				i++
			}
		case "--project", "-p":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Projects = append(filters.Projects, args[i+1])
				i++
			}
		case "--priority", "-P":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Priorities = append(filters.Priorities, args[i+1])
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
		hasOtherFlags := showAll || showDone || showTrash || filters.Today || filters.Tomorrow ||
			filters.Overdue || filters.Upcoming || filters.DueDate != "" || filters.FromDate != "" ||
			filters.ToDate != "" || filters.Query != "" || len(filters.Tags) > 0 ||
			len(filters.Projects) > 0 || len(filters.Priorities) > 0 || showNote

		if hasOtherFlags {
			return fmt.Errorf("--tags, --projects, and --priorities cannot be combined with other filters")
		}

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

	// If no specific filter, show dashboard
	showDashboard := !showAll && !showDone && !showTrash && !filters.Today && !filters.Tomorrow &&
		!filters.Overdue && !filters.Upcoming && filters.DueDate == "" && filters.FromDate == "" &&
		filters.ToDate == "" && filters.Query == "" && len(filters.Tags) == 0 &&
		len(filters.Projects) == 0 && len(filters.Priorities) == 0

	if showDashboard {
		return displayDashboard(database, showNote)
	}

	// Get initial task set
	var tasks []models.Task
	var err error

	if showTrash {
		tasks, err = db.GetDeletedTasks(database)
	} else if showDone {
		allTasks, err := db.GetTasks(database, true)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
		for _, task := range allTasks {
			if task.CompletedDate != nil {
				tasks = append(tasks, task)
			}
		}
	} else {
		tasks, err = db.GetTasks(database, showAll)
	}

	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// Apply filters
	tasks, err = filter.ApplyFilters(tasks, filters)
	if err != nil {
		return fmt.Errorf("failed to apply filters: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	printTasks(tasks, showNote)
	return nil
}
