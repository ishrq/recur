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
			},
			Examples: []string{
				"recur add \"Buy groceries\"",
				"recur add \"Team meeting @(2025-11-12 15:00) +Work\"",
				"recur add \"Water plants @(today 9am, 1d) #chores !urgent\"",
			},
			ComingSoon: []string{
				"@(date time, frequency, end)  - Due date and recurrence",
				"#tag                          - Add a tag",
				"+project                      - Add to project",
				"!priority                     - Set priority",
				"*note                         - Add a note",
			},
		},
		"ls": {
			Name:        "ls",
			Description: "List tasks",
			Usage: []string{
				"recur ls                    Show task dashboard",
				"recur ls --all              Show all tasks including completed",
				"recur ls --done             Show completed tasks",
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
				"recur ls --note             Include notes in task list",
			},
			Options: []Option{
				{"-a, --all", "Show all tasks including completed"},
				{"-x, --done", "Show completed tasks"},
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
				{"-n, --note", "Include notes in task list"},
				{"--tags", "List all tags with task counts"},
				{"--projects", "List all projects with task counts"},
				{"--priorities", "List all priorities with task counts"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur ls",
				"recur ls --tags",
				"recur ls --today",
				"recur ls --tag work",
				"recur ls --today --note",
				"recur ls --project Work --note",
				"recur ls --from today --to 2025-11-20 --note",
				"recur ls --done",
				"recur ls --done --tag work",
				"recur ls --done --from 2025-11-01",
			},
			ComingSoon: []string{
				"--deleted",
				"--query for search",
				"--export to CSV",
			},
		},
		"done": {
			Name:        "done",
			Description: "Mark tasks as complete",
			Usage: []string{
				"recur done <id>              Complete a task",
				"recur done <id1> <id2> ...   Complete multiple tasks",
			},
			Options: []Option{
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur done 1",
				"recur done 1 2 3",
				"recur done 5",
			},
			ComingSoon: []string{
				"--tag, --project, --priority",
				"--due, --query",
				"--undo",
			},
		},
		"rm": {
			Name:        "rm",
			Description: "Delete tasks",
			Usage: []string{
				"recur rm <id>               Delete a task",
				"recur rm <id1> <id2> ...    Delete multiple tasks",
			},
			Options: []Option{
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur rm 1",
				"recur rm 1 2 3",
			},
			ComingSoon: []string{
				"--tag, --project, --priority",
				"--due, --query",
				"--undo",
				"--completed, --all, --purge",
			},
		},
		"cp": {
			Name:        "cp",
			Description: "Copy/duplicate tasks",
			Usage: []string{
				"recur cp <id>               Duplicate a task",
				"recur cp <id1> <id2> ...    Duplicate multiple tasks",
				"recur cp <id> \"New name @(date) #tag !project !priority *note\"",
			},
			Options: []Option{
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur cp 1",
				"recur cp 1 2 3",
				"recur cp 5 \"Modified copy @(tomorrow) !Urgent\"",
			},
		},
		"mv": {
			Name:        "mv",
			Description: "Edit/modify tasks",
			Usage: []string{
				"recur mv <id> \"New name @(date) #tag !project !priority *note\"",
				"recur mv <id>               Edit task in $EDITOR",
				"recur mv <id1> <id2> ...    Edit multiple tasks in $EDITOR",
			},
			Options: []Option{
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur mv 1 \"Updated task name\"",
				"recur mv 3 \"@(tomorrow 3pm)\"",
				"recur mv 5",
			},
			ComingSoon: []string{
				"Bulk edit with --tag, --project, etc.",
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

	Run 'recur <command> --help' for more information on a command.`

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
			fmt.Printf("  %-28s %s\n", opt.Flag, opt.Description)
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
