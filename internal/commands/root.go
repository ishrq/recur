package commands

import (
	"database/sql"
	"fmt"
	"os"
)

func Execute(db *sql.DB) {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "add":
		err = Add(db, args)
	case "ls", "list":
		err = List(db, args)
	case "done":
		err = Done(db, args)
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
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
