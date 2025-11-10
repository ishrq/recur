package main

import (
	"fmt"
	"os"

	"github.com/ishrq/recur/internal/commands"
	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/utils"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := utils.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := utils.GetDefaultConfig()
	database, err := db.InitDB(config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	commands.Execute(database)

	return nil
}
