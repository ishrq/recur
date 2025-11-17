package commands

import (
	"fmt"
)

type CommandHelp struct {
	Name        string
	Description string
	Usage       []string
	Options     []Option
	Examples    []string
	ComingSoon  []string
}

type Option struct {
	Flag        string
	Description string
}

func PrintHelp(command string) {
	if command == "" || command == "help" {
		printMainHelp()
		return
	}

	helps := map[string]CommandHelp{
		"add": {
			Name:        "add",
			Description: "Add a new task",
			Usage: []string{
				"recur add \"Task name\"",
				"recur add \"Task name @(date time, frequency, end) #tag +project !priority *note\"",
				"recur add --edit                      Opens editor to add task(s)",
				"recur add --edit \"Task name #tag\"    Opens editor with prefilled content",
			},
			Options: []Option{
				{"-e, --edit", "Open $EDITOR to add task(s) (one per line)"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur add \"Buy groceries\"",
				"recur add \"Team meeting @(2025-11-20 15:00) +Work\"",
				"recur add \"Water plants @(today 9am, 1d) #chores\"",
				"recur add \"Weekly review @(Friday 5pm, 1w, 2025-12-31) +Personal\"",
				"recur add \"Dentist @(2025-11-20 2pm) !urgent *'Bring insurance card'\"",
				"recur add --edit",
				"recur add --edit \"Read book #leisure\"",
			},
		},

		"dates": {
			Name:        "Date Parsing Reference",
			Description: "Comprehensive guide to date and time formats in Recur",
			Usage: []string{
				"Date format: @(date time, frequency, end date)",
				"  - date time: When the task is due",
				"  - frequency: How often it repeats (optional)",
				"  - end date: When recurring stops (optional)",
			},
			Options: []Option{
				{"Semantic Dates", ""},
				{"  now", "Current date and time"},
				{"  today", "Today at 12:00 PM"},
				{"  tomorrow, tmr", "Tomorrow at 12:00 PM"},
				{"  yesterday", "Yesterday at 12:00 PM"},
				{"  next week", "7 days from now at 12:00 PM"},
				{"  next month", "Same day next month at 12:00 PM"},
				{"", ""},

				{"Weekday Names", ""},
				{"  monday, mon", "Next Monday at 12:00 PM"},
				{"  tuesday, tue", "Next Tuesday at 12:00 PM"},
				{"  wednesday, wed", "Next Wednesday at 12:00 PM"},
				{"  thursday, thu", "Next Thursday at 12:00 PM"},
				{"  friday, fri", "Next Friday at 12:00 PM"},
				{"  saturday, sat", "Next Saturday at 12:00 PM"},
				{"  sunday, sun", "Next Sunday at 12:00 PM"},
				{"", ""},

				{"Relative Dates", ""},
				{"  +3d", "3 days from now"},
				{"  -2w", "2 weeks ago"},
				{"  +1m", "1 month from now"},
				{"  +1y", "1 year from now"},
				{"  +6h", "6 hours from now"},
				{"", ""},

				{"Standard Formats", ""},
				{"  2025-11-20", "ISO date (default 12:00 PM)"},
				{"  2025-11-20 15:04", "ISO date with time (24-hour)"},
				{"  2025-11-20 3:04pm", "ISO date with time (12-hour)"},
				{"  2025/11/20", "Date with slashes"},
				{"  11/20/2025", "US date format"},
				{"  20-11-2025", "European date format"},
				{"  Jan 20", "Month and day (current year, 12:00 PM)"},
				{"  Jan 20 3pm", "Month, day, and time"},
				{"  January 20 15:04", "Full month name with time"},
				{"", ""},

				{"Frequency Formats", ""},
				{"  1h, 2h, 12h", "Every N hours"},
				{"  1d, 2d, 7d", "Every N days"},
				{"  1w, 2w, 4w", "Every N weeks"},
				{"  1m, 2m, 6m", "Every N months (calendar-aware)"},
				{"  1y, 2y", "Every N years (calendar-aware)"},
				{"  hourly", "Every hour"},
				{"  daily", "Every day"},
				{"  weekly", "Every week"},
				{"  monthly", "Every month"},
				{"  yearly", "Every year"},
			},
			Examples: []string{
				"# Semantic dates",
				"recur add \"Call mom @(tomorrow)\"",
				"recur add \"Standup @(today 9am)\"",
				"recur add \"Lunch meeting @(friday 12pm)\"",
				"",
				"# Relative dates",
				"recur add \"Follow up @(+3d)\"",
				"recur add \"Review notes from @(-1w)\"",
				"recur add \"Plan trip @(+2m)\"",
				"",
				"# Standard formats",
				"recur add \"Deadline @(2025-12-31)\"",
				"recur add \"Appointment @(2025-11-20 14:30)\"",
				"recur add \"Call @(Jan 15 3pm)\"",
				"",
				"# Recurring tasks",
				"recur add \"Take vitamins @(today 8am, daily)\"",
				"recur add \"Backup files @(monday 6pm, weekly)\"",
				"recur add \"Pay rent @(2025-12-01, monthly, 2026-12-01)\"",
				"recur add \"Quarterly review @(2025-12-31, 3m, 2027-12-31)\"",
				"recur add \"Team sync @(tomorrow 10am, 2d, +30d)\"",
				"",
				"# Calendar-aware monthly recurrence",
				"recur add \"Bill due @(2025-01-31, monthly)\"  # Feb 28/29, Mar 31, etc.",
				"",
				"# Complex examples",
				"recur add \"Gym @(monday 6am, 2d, +60d) #health !high\"",
				"recur add \"Water plants @(today, 3d) #chores +home *'Living room only'\"",
			},
		},
		"ls": {
			Name:        "ls",
			Description: "List tasks",
			Usage: []string{
				"recur ls                    Show task dashboard",
				"recur ls --all              Show all tasks (completed and incomplete)",
				"recur ls --done             Show completed tasks",
				"recur ls --trash            Show deleted tasks",
				"recur ls --tags             List all tags with counts",
				"recur ls --projects         List all projects with counts",
				"recur ls --priorities       List all priorities with counts",
				"recur ls --today            Show tasks due today",
				"recur ls --tomorrow         Show tasks due tomorrow",
				"recur ls --overdue          Show overdue tasks",
				"recur ls --upcoming         Show upcoming tasks (next 7 days)",
				"recur ls --due <date>       Show tasks due on specific date",
				"recur ls --from <date>      Show tasks from date onwards",
				"recur ls --to <date>        Show tasks up to date",
				"recur ls --tag <tag>        Show tasks with specific tag",
				"recur ls --project <proj>   Show tasks in specific project",
				"recur ls --priority <pri>   Show tasks with specific priority",
				"recur ls --query <keyword>  Search tasks by keyword",
				"recur ls --note             Include notes in task list",
				"recur ls --export [path]    Export filtered tasks to CSV",  // ADD THIS LINE
			},
			Options: []Option{
				{"-a, --all", "Show all tasks (completed and incomplete)"},
				{"-x, --done", "Show completed tasks"},
				{"--trash", "Show deleted tasks"},
				{"--tags", "List all tags with task counts"},
				{"--projects", "List all projects with task counts"},
				{"--priorities", "List all priorities with task counts"},
				{"--today", "Show tasks due today"},
				{"--tomorrow", "Show tasks due tomorrow"},
				{"--overdue", "Show overdue tasks"},
				{"--upcoming", "Show upcoming tasks (next 7 days)"},
				{"-d, --due <date>", "Show tasks due on specific date"},
				{"--from <date>", "Show tasks from date onwards (inclusive)"},
				{"--to <date>", "Show tasks up to date (inclusive)"},
				{"-t, --tag <tag>", "Filter by tag (can specify multiple)"},
				{"-p, --project <proj>", "Filter by project (can specify multiple)"},
				{"-P, --priority <pri>", "Filter by priority (can specify multiple)"},
				{"-q, --query <keyword>", "Search in task name, tag, project, and note"},
				{"-n, --note", "Include notes in task list"},
				{"--export [path]", "Export to CSV (auto-generates filename if not provided)"},  // ADD THIS LINE
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"# Basic listing",
				"recur ls                        # Show dashboard",
				"recur ls --all                  # All tasks",
				"recur ls --done                 # Completed tasks",
				"recur ls --trash                # Deleted tasks",
				"",
				"# Metadata listings",
				"recur ls --tags                 # List all tags",
				"recur ls --projects             # List all projects",
				"recur ls --priorities           # List all priorities",
				"",
				"# Date filters",
				"recur ls --today",
				"recur ls --tomorrow",
				"recur ls --overdue",
				"recur ls --upcoming",
				"recur ls --due 2025-11-15",
				"recur ls --due tomorrow",
				"recur ls --due none             # Tasks with no due date",
				"recur ls --due recurring        # Recurring tasks",
				"recur ls --from 2025-11-15",
				"recur ls --to 2025-12-31",
				"recur ls --from today --to 2025-11-20",
				"",
				"# Attribute filters",
				"recur ls --tag work",
				"recur ls --project Home",
				"recur ls --priority urgent high",
				"recur ls --query meeting",
				"",
				"# Combined filters",
				"recur ls --today --note",
				"recur ls --tag work --project Office",
				"recur ls --from 2025-11-15 --tag work --priority high",
				"",
				"# Export to CSV",
				"recur ls --export                                 # recur_20250117_143022.csv",
				"recur ls --today --export                         # recur_today.csv",
				"recur ls --due 2025-11 --export                   # recur_202511.csv",
				"recur ls --tag work --priority high --export      # recur_work_high.csv",
				"recur ls --overdue --export overdue_tasks.csv     # Custom filename",
			},
		},
		"done": {
			Name:        "done",
			Description: "Mark tasks as complete",
			Usage: []string{
				"recur done <id>              Complete a task",
				"recur done <id1> <id2> ...   Complete multiple tasks",
				"recur done --tag <tag>       Complete tasks by tag",
				"recur done --project <proj>  Complete tasks by project",
				"recur done --priority <pri>  Complete tasks by priority",
				"recur done --due <date>      Complete tasks due on date",
				"recur done --undo <id>       Unmark completed task(s)",
				"recur done --undo --tag <tag> Unmark completed tasks by filter",
			},
			Options: []Option{
				{"--today", "Complete tasks due today"},
				{"--tomorrow", "Complete tasks due tomorrow"},
				{"--overdue", "Complete overdue tasks"},
				{"--upcoming", "Complete upcoming tasks"},
				{"-d, --due <date>", "Complete tasks due on specific date"},
				{"--from <date>", "Complete tasks from date onwards"},
				{"--to <date>", "Complete tasks up to date"},
				{"-t, --tag <tag>", "Complete tasks with specific tag"},
				{"-p, --project <proj>", "Complete tasks in specific project"},
				{"-P, --priority <pri>", "Complete tasks with specific priority"},
				{"-q, --query <keyword>", "Complete tasks matching search"},
				{"--undo", "Unmark completed tasks as incomplete"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur done 1",
				"recur done 1 2 3",
				"recur done --today",
				"recur done --tag work",
				"recur done --project Home",
				"recur done --priority urgent",
				"recur done --due today",
				"recur done --query meeting",
				"recur done --tag work --due today",
				"recur done --undo 1 2 3",
				"recur done --undo --tag work",
				"recur done --undo --today",
			},
		},
		"rm": {
			Name:        "rm",
			Description: "Delete tasks",
			Usage: []string{
				"recur rm <id>               Delete a task",
				"recur rm <id1> <id2> ...    Delete multiple tasks",
				"recur rm --tag <tag>        Delete tasks by tag",
				"recur rm --project <proj>   Delete tasks by project",
				"recur rm --priority <pri>   Delete tasks by priority",
				"recur rm --due <date>       Delete tasks due on date",
				"recur rm --all              Delete all incomplete tasks",
				"recur rm --done             Delete all completed tasks",
				"recur rm --trash            Permanently delete all trashed tasks",
				"recur rm --purge            Permanently delete ALL tasks from database",
				"recur rm --undo <id>        Restore deleted task(s)",
				"recur rm --undo --tag <tag> Restore deleted tasks by filter",
			},
			Options: []Option{
				{"--all", "Delete all incomplete tasks"},
				{"--done", "Delete all completed tasks"},
				{"--trash", "Permanently delete all trashed tasks (cannot be undone)"},
				{"--purge", "Permanently delete ALL tasks from database (cannot be undone)"},
				{"--undo", "Restore deleted tasks (cannot combine with other operations)"},
				{"--today", "Delete tasks due today"},
				{"--tomorrow", "Delete tasks due tomorrow"},
				{"--overdue", "Delete overdue tasks"},
				{"--upcoming", "Delete upcoming tasks"},
				{"-d, --due <date>", "Delete tasks due on specific date"},
				{"--from <date>", "Delete tasks from date onwards"},
				{"--to <date>", "Delete tasks up to date"},
				{"-t, --tag <tag>", "Delete tasks with specific tag (can specify multiple)"},
				{"-p, --project <proj>", "Delete tasks in specific project (can specify multiple)"},
				{"-P, --priority <pri>", "Delete tasks with specific priority (can specify multiple)"},
				{"-q, --query <keyword>", "Delete tasks matching search"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur rm 1",
				"recur rm 1 2 3",
				"recur rm --tag work",
				"recur rm --project Home",
				"recur rm --priority urgent high",
				"recur rm --due today",
				"recur rm --from 2025-11-01 --to 2025-11-30",
				"recur rm --query meeting",
				"recur rm --tag work --due today",
				"recur rm --all",
				"recur rm --done",
				"recur rm --trash",
				"recur rm --purge",
				"recur rm --undo 1 2 3",
				"recur rm --undo --tag work",
				"recur rm --undo --query meeting",
			},
		},
		"cp": {
			Name:        "cp",
			Description: "Copy/duplicate tasks",
			Usage: []string{
				"recur cp <id>                      Duplicate a task",
				"recur cp <id1> <id2> ...           Duplicate multiple tasks",
				"recur cp <id> \"New name @(date) #tag +project !priority *note\"",
				"recur cp --edit <id>               Edit task in $EDITOR before copying",
				"recur cp --edit <id1> <id2> ...    Edit multiple tasks in $EDITOR before copying",
				"recur cp --tag <tag>               Copy tasks by tag",
				"recur cp --project <proj>          Copy tasks by project",
			},
			Options: []Option{
				{"--today", "Copy tasks due today"},
				{"--tomorrow", "Copy tasks due tomorrow"},
				{"--overdue", "Copy overdue tasks"},
				{"--upcoming", "Copy upcoming tasks"},
				{"-d, --due <date>", "Copy tasks due on specific date"},
				{"--from <date>", "Copy tasks from date onwards"},
				{"--to <date>", "Copy tasks up to date"},
				{"-t, --tag <tag>", "Copy tasks with specific tag"},
				{"-p, --project <proj>", "Copy tasks in specific project"},
				{"-P, --priority <pri>", "Copy tasks with specific priority"},
				{"-q, --query <keyword>", "Copy tasks matching search"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur cp 1",
				"recur cp 1 2 3",
				"recur cp 5 \"Modified copy @(tomorrow) !Urgent\"",
				"recur cp --edit 1 2 3                  # Opens editor",
				"recur cp --tag work",
				"recur cp --project Home",
				"recur cp --due today",
				"recur cp --tag work --project Office",
			},
		},
		"mv": {
			Name:        "mv",
			Description: "Edit/modify tasks",
			Usage: []string{
				"recur mv <id> \"New name @(date) #tag +project !priority *note\"",
				"recur mv <id>               Edit task in $EDITOR",
				"recur mv <id1> <id2> ...    Edit multiple tasks in $EDITOR",
				"recur mv --tag <tag> \"...\" Edit tasks by tag",
				"recur mv --project <proj> \"...\" Edit tasks by project",
			},
			Options: []Option{
				{"--today", "Edit tasks due today"},
				{"--tomorrow", "Edit tasks due tomorrow"},
				{"--overdue", "Edit overdue tasks"},
				{"--upcoming", "Edit upcoming tasks"},
				{"-d, --due <date>", "Edit tasks due on specific date"},
				{"--from <date>", "Edit tasks from date onwards"},
				{"--to <date>", "Edit tasks up to date"},
				{"-t, --tag <tag>", "Edit tasks with specific tag"},
				{"-p, --project <proj>", "Edit tasks in specific project"},
				{"-P, --priority <pri>", "Edit tasks with specific priority"},
				{"-q, --query <keyword>", "Edit tasks matching search"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur mv 1 \"Updated task name\"",
				"recur mv 3 \"@(tomorrow 3pm)\"",
				"recur mv 5 \"New name +NewProject\"",
				"recur mv 1",
				"recur mv 1 2 3",
				"recur mv --tag work",
				"recur mv --tag work \"@(tomorrow)\"",
				"recur mv --due today \"!urgent\"",
				"recur mv --project Home \"+Personal\"",
			},
		},
	}

	help, exists := helps[command]
	if !exists {
		fmt.Printf("No help available for command: %s\n", command)
		return
	}

	printCommandHelp(help)
}

func printMainHelp() {
	help := `Recur - A CLI todo app with recurring tasks

	Usage:
	recur <command> [arguments]

	Commands:
	add     Add a new task
	ls      List tasks
	done    Complete task(s)
	rm      Delete task(s)
	cp      Copy task(s)
	mv      Edit task(s)
	help    Show this help message

	Run 'recur <command> --help' for more information on a command.

	Examples:
	recur add "Buy groceries @(tomorrow) #shopping"
	recur ls --today
	recur done --tag work
	recur rm --undo 5`

	fmt.Println(help)
}

func printCommandHelp(help CommandHelp) {
	fmt.Printf("%s\n\n", help.Description)

	if len(help.Usage) > 0 {
		fmt.Println("Usage:")
		for _, usage := range help.Usage {
			fmt.Printf("  %s\n", usage)
		}
		fmt.Println()
	}

	if len(help.Options) > 0 {
		fmt.Println("Options:")
		for _, opt := range help.Options {
			fmt.Printf("  %-30s %s\n", opt.Flag, opt.Description)
		}
		fmt.Println()
	}

	if len(help.Examples) > 0 {
		fmt.Println("Examples:")
		for _, example := range help.Examples {
			fmt.Printf("  %s\n", example)
		}
		fmt.Println()
	}

	if len(help.ComingSoon) > 0 {
		fmt.Println("Coming soon:")
		for _, feature := range help.ComingSoon {
			fmt.Printf("  %s\n", feature)
		}
	}
}
