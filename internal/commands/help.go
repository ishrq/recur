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
		"add":   helpAdd,
		"dates": helpDates,
		"ls":    helpLs,
		"done":  helpDone,
		"rm":    helpRm,
		"cp":    helpCp,
		"mv":    helpMv,
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
