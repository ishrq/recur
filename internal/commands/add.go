package commands

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/editor"
	"github.com/ishrq/recur/internal/models"
	"github.com/ishrq/recur/internal/parser"
)

func Add(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task name required (or use --edit flag)")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("add")
		return nil
	}

	if args[0] == "--edit" || args[0] == "-e" {
		// pre-populate editor with remaining args
		var initialContent string
		if len(args) > 1 {
			initialContent = strings.Join(args[1:], " ")
		}
		return addWithEditor(database, initialContent)
	}

	// Inline add
	taskString := strings.Join(args, " ")

	// Parse task string
	task, err := parser.ParseTaskString(taskString)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	id, err := CreateTask(database, task)
	if err != nil {
		return err
	}

	printTaskConfirmation(id, task)

	return nil
}

func addWithEditor(database *sql.DB, initialContent string) error {
	content := initialContent
	if content == "" {
		content = "# Add tasks below (one per line)\n# Example: Buy groceries @(tomorrow) #personal +home !high *'Don't forget milk'\n\n"
	} else {
		content += "\n"
	}

	for {
		edited, err := editor.OpenEditor(content)
		if err != nil {
			return err
		}

		tasks, errors := parseMultipleTaskLines(edited)

		if len(errors) > 0 {
			fmt.Println("\nErrors found:")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
			fmt.Print("\nPress Enter to re-edit, or Ctrl+C to cancel: ")
			bufio.NewReader(os.Stdin).ReadString('\n')
			content = edited // Use the invalid content so user can fix it
			continue
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks to add (empty or all lines were comments)")
			return nil
		}

		if len(tasks) > 1 {
			fmt.Printf("\nFound %d tasks to add:\n", len(tasks))
			for i, task := range tasks {
				fmt.Printf("%d. %s\n", i+1, task.Name)
			}

			ok, err := confirmPrompt(fmt.Sprintf("\nAdd these %d tasks? (y/n): ", len(tasks)))
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		added := 0
		for _, task := range tasks {
			id, err := CreateTask(database, task)
			if err != nil {
				fmt.Printf("Warning: Failed to add task '%s': %v\n", task.Name, err)
				continue
			}

			if len(tasks) == 1 {
				// Detailed output for single task
				printTaskConfirmation(id, task)
			} else {
				// Compact output for multiple tasks
				fmt.Printf("✓ Added #%d: %s\n", id, task.Name)
			}
			added++
		}

		if added > 1 {
			fmt.Printf("\n%d tasks added\n", added)
		}

		return nil
	}
}

func parseMultipleTaskLines(content string) ([]*models.Task, []error) {
	lines := strings.Split(content, "\n")
	var tasks []*models.Task
	var errors []error

	lineNum := 0
	for _, line := range lines {
		lineNum++
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		task, err := parser.ParseTaskString(line)
		if err != nil {
			errors = append(errors, fmt.Errorf("line %d: %v", lineNum, err))
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, errors
}

func CreateTask(database *sql.DB, task *models.Task) (int64, error) {
	id, err := db.InsertTask(database, task)
	if err != nil {
		return 0, fmt.Errorf("failed to add task: %w", err)
	}
	return id, nil
}

func printTaskConfirmation(id int64, task *models.Task) {
	fmt.Printf("Added task #%d: %s\n", id, task.Name)

	if task.DueDate != nil {
		fmt.Printf("  Due: %s\n", task.DueDate.Format("Mon Jan 2, 2006 15:04"))
	}
	if task.RecurFrequency != "" {
		fmt.Printf("  Recurs: %s\n", task.RecurFrequency)
	}
	if task.RecurEndDate != nil {
		fmt.Printf("  Until: %s\n", task.RecurEndDate.Format("Mon Jan 2, 2006"))
	}
	if task.Project != "" {
		fmt.Printf("  Project: %s\n", task.Project)
	}
	if task.Tag != "" {
		fmt.Printf("  Tag: %s\n", task.Tag)
	}
	if task.Priority != "" {
		fmt.Printf("  Priority: %s\n", task.Priority)
	}
	if task.Note != "" {
		fmt.Printf("  Note: %s\n", task.Note)
	}
}
