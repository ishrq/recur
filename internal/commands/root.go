package commands

import (
	"fmt"
	"os"
)

func Execute() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		fmt.Println("Add command - not yet implemented")
	case "ls":
		fmt.Println("List command - not yet implemented")
	case "done":
		fmt.Println("Done command - not yet implemented")
	case "rm":
		fmt.Println("Remove command - not yet implemented")
	case "cp":
		fmt.Println("Copy command - not yet implemented")
	case "mv":
		fmt.Println("Move command - not yet implemented")
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
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
