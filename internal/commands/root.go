package commands

import (
	"database/sql"
	"fmt"
	"os"
)

func Execute(db *sql.DB) {
	if len(os.Args) < 2 {
		PrintHelp("")
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
	case "rm", "remove":
		err = Remove(db, args)
	case "cp":
		fmt.Println("Copy command - not yet implemented")
	case "mv":
		fmt.Println("Move command - not yet implemented")
	case "help", "--help", "-h":
		if len(args) > 0 {
			PrintHelp(args[0])
		} else {
			PrintHelp("")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		PrintHelp("")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
