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

	filters, remaining := extractFilterFlags(args)

	var showAll bool
	var showDone bool
	var showTrash bool
	var showNote bool
	var listTags bool
	var listProjects bool
	var listPriorities bool
	var exportPath string
	var exportFlag bool

	for i := 0; i < len(remaining); i++ {
		arg := remaining[i]
		switch arg {
		case "--all", "-a":
			showAll = true
		case "--done", "-x":
			showDone = true
		case "--trash":
			showTrash = true
		case "--tags":
			listTags = true
		case "--projects":
			listProjects = true
		case "--priorities":
			listPriorities = true
		case "--note", "-n":
			showNote = true
		case "--export":
			exportFlag = true
			if i+1 < len(remaining) && !strings.HasPrefix(remaining[i+1], "-") {
				exportPath = remaining[i+1]
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
			len(filters.Projects) > 0 || len(filters.Priorities) > 0 || showNote || exportFlag

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

	if showDashboard && !exportFlag {
		return displayDashboard(database, showNote)
	}

	// Get initial task set
	var tasks []models.Task
	var err error

	if showTrash {
		tasks, err = db.GetTasks(database, true, false)
	} else if showDone {
		tasks, err = db.GetTasks(database, false, true)
	} else if showAll {
		// Get both completed and incomplete
		incompleteTasks, err1 := db.GetTasks(database, false, false)
		completedTasks, err2 := db.GetTasks(database, false, true)
		if err1 != nil {
			return fmt.Errorf("failed to get tasks: %w", err1)
		}
		if err2 != nil {
			return fmt.Errorf("failed to get tasks: %w", err2)
		}
		tasks = append(incompleteTasks, completedTasks...)
	} else {
		tasks, err = db.GetTasks(database, false, false)
	}

	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	tasks, err = filter.ApplyFilters(tasks, filters)
	if err != nil {
		return fmt.Errorf("failed to apply filters: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	if exportFlag {
		if exportPath == "" {
			exportPath = generateExportFilename(filters, showAll, showDone, showTrash)
		}
		return exportTasksToCSV(tasks, exportPath)
	}

	printTasks(tasks, showNote)
	return nil
}
