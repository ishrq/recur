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
				"recur ls                    Show incomplete tasks",
				"recur ls --all              Show all tasks including completed",
			},
			Options: []Option{
				{"-a, --all", "Show all tasks including completed"},
				{"-h, --help", "Show this help message"},
			},
			Examples: []string{
				"recur ls",
				"recur ls --all",
			},
			ComingSoon: []string{
				"--today, --tomorrow, --overdue, --upcoming",
				"--due, --from, --to",
				"--tag, --project, --priority",
				"--completed, --deleted",
				"--query for search",
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
