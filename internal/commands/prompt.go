package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
