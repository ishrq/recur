package commands

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ishrq/recur/internal/db"
	"github.com/ishrq/recur/internal/filter"
	"github.com/ishrq/recur/internal/models"
)

func Remove(database *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID or filter required")
	}

	if args[0] == "--help" || args[0] == "-h" {
		PrintHelp("rm")
		return nil
	}

	filters, remaining := extractFilterFlags(args)

	var ids []int
	var removeAll bool
	var removeDone bool
	var removeTrash bool
	var purge bool
	var undo bool

	for _, arg := range remaining {
		switch arg {
		case "--all":
			removeAll = true
		case "--done":
			removeDone = true
		case "--trash":
			removeTrash = true
		case "--purge":
			purge = true
		case "--undo":
			undo = true
		default:
			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", arg)
			}
			ids = append(ids, id)
		}
	}

	if undo {
		if removeAll || removeDone || removeTrash || purge {
			return fmt.Errorf("--undo cannot be combined with --all, --done, --trash, or --purge")
		}
	}

	specialFlags := []bool{removeAll, removeDone, removeTrash, purge}
	specialFlagCount := 0
	for _, flag := range specialFlags {
		if flag {
			specialFlagCount++
		}
	}
	if specialFlagCount > 1 {
		return fmt.Errorf("cannot combine --all, --done, --trash, and --purge flags")
	}

	switch {
	case purge:
		return purgeAllTasks(database)
	case undo:
		return restoreDeletedTasks(database, ids, filters)
	default:
		return removeTasks(database, ids, filters, removeAll, removeDone, removeTrash)
	}
}

func purgeAllTasks(database *sql.DB) error {
	count, err := db.GetAllTasksCount(database)
	if err != nil {
		return fmt.Errorf("failed to count tasks: %w", err)
	}

	if count == 0 {
		fmt.Println("No tasks in database.")
		return nil
	}

	fmt.Println("\n⚠️  WARNING: This will PERMANENTLY delete ALL tasks from the database.")
	fmt.Println("⚠️  This action CANNOT be undone!")
	fmt.Printf("\nTotal tasks in database: %d\n\n", count)

	ok, err := confirmSpecific("Type 'DELETE ALL' to confirm: ", "DELETE ALL")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Purge cancelled.")
		return nil
	}

	if err := db.PurgeAllTasks(database); err != nil {
		return fmt.Errorf("failed to purge tasks: %w", err)
	}

	fmt.Printf("\n✓ All tasks permanently deleted\n")
	return nil
}

func restoreDeletedTasks(database *sql.DB, ids []int, filters filter.Filters) error {
	var tasks []models.Task
	var err error

	if len(ids) > 0 {
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil {
				fmt.Printf("Warning: Task #%d not found\n", id)
				continue
			}
			if task.Deleted {
				tasks = append(tasks, *task)
			} else {
				fmt.Printf("Warning: Task #%d is not deleted\n", id)
			}
		}
	} else {
		tasks, err = db.GetTasks(database, true, false)
		if err != nil {
			return fmt.Errorf("failed to get deleted tasks: %w", err)
		}
		tasks, err = filter.ApplyFilters(tasks, filters)
		if err != nil {
			return err
		}
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no deleted tasks found matching criteria")
	}

	fmt.Printf("\nFound %d deleted task(s) to restore:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := confirmPrompt(fmt.Sprintf("Restore these %d task(s)? (y/n): ", len(tasks)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Restore cancelled.")
		return nil
	}

	restored := 0
	for _, t := range tasks {
		if err := db.RestoreTask(database, t.ID); err != nil {
			fmt.Printf("Warning: Failed to restore task #%d: %v\n", t.ID, err)
			continue
		}
		fmt.Printf("↺ Restored #%d: %s\n", t.ID, t.Name)
		restored++
	}

	if restored > 0 {
		fmt.Printf("\n%d task(s) restored\n", restored)
	}

	return nil
}

func removeTasks(database *sql.DB, ids []int, filters filter.Filters, removeAll, removeDone, removeTrash bool) error {
	var tasks []models.Task
	var err error
	permanentDelete := false

	switch {
	case removeTrash:
		permanentDelete = true
		tasks, err = db.GetTasks(database, true, false)
	case removeAll:
		tasks, err = db.GetTasks(database, false, false)
	case removeDone:
		tasks, err = db.GetTasks(database, false, true)
	case len(ids) > 0:
		for _, id := range ids {
			task, err := db.GetTaskByID(database, id)
			if err != nil {
				fmt.Printf("Warning: Task #%d not found\n", id)
				continue
			}
			if !task.Deleted {
				tasks = append(tasks, *task)
			} else {
				fmt.Printf("Warning: Task #%d is already deleted\n", id)
			}
		}
	default:
		tasks, err = db.GetTasks(database, false, false)
	}

	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	// Apply filters to the initial task set
	tasks, err = filter.ApplyFilters(tasks, filters)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found matching criteria")
	}

	fmt.Println()
	if permanentDelete {
		fmt.Println("⚠️  WARNING: This will PERMANENTLY delete tasks from the database.")
		fmt.Println("⚠️  This action CANNOT be undone!")
		fmt.Println()
	}
	fmt.Printf("Found %d task(s) to delete:\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	prompt := fmt.Sprintf("Delete these %d task(s)? (y/n): ", len(tasks))
	if permanentDelete {
		prompt = fmt.Sprintf("Permanently delete these %d task(s)? (y/n): ", len(tasks))
	}

	ok, err := confirmPrompt(prompt)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	deleted := 0
	for _, t := range tasks {
		var err error
		if permanentDelete {
			err = db.PermanentlyDeleteTask(database, t.ID)
		} else {
			err = db.DeleteTask(database, t.ID)
		}
		if err != nil {
			fmt.Printf("Warning: Failed to delete task #%d: %v\n", t.ID, err)
			continue
		}
		fmt.Printf("✗ Deleted #%d: %s\n", t.ID, t.Name)
		deleted++
	}

	if deleted > 0 {
		if permanentDelete {
			fmt.Printf("\n%d task(s) permanently deleted\n", deleted)
		} else {
			fmt.Printf("\n%d task(s) deleted\n", deleted)
		}
	}

	return nil
}
