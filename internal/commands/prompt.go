package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ishrq/recur/internal/models"
)

var ConfirmPrompt = defaultConfirmPrompt
var ConfirmSpecific = defaultConfirmSpecific

func defaultConfirmPrompt(msg string) (bool, error) {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func defaultConfirmSpecific(msg, expected string) (bool, error) {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}
	response = strings.TrimSpace(response)
	return response == expected, nil
}

func confirmTasks(tasks []models.Task, title, prompt string) (bool, error) {
	fmt.Printf("\n%s\n", title)
	for _, t := range tasks {
		fmt.Printf("#%-4d %s\n", t.ID, t.Name)
	}
	fmt.Println()

	ok, err := ConfirmPrompt(prompt)
	if err != nil {
		return false, err
	}
	if !ok {
		fmt.Println("Operation cancelled.")
		return false, nil
	}
	return true, nil
}
