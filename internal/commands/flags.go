package commands

import (
	"strings"

	"github.com/ishrq/recur/internal/filter"
)

func extractFilterFlags(args []string) (filter.Filters, []string) {
	var filters filter.Filters
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--today":
			filters.Today = true
		case "--tomorrow":
			filters.Tomorrow = true
		case "--overdue":
			filters.Overdue = true
		case "--upcoming":
			filters.Upcoming = true
		case "--due", "-d":
			if i+1 < len(args) {
				filters.DueDate = args[i+1]
				i++
			}
		case "--from":
			if i+1 < len(args) {
				filters.FromDate = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				filters.ToDate = args[i+1]
				i++
			}
		case "--query", "-q":
			if i+1 < len(args) {
				filters.Query = args[i+1]
				i++
			}
		case "--tag", "-t":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Tags = append(filters.Tags, args[i+1])
				i++
			}
		case "--project", "-p":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Projects = append(filters.Projects, args[i+1])
				i++
			}
		case "--priority", "-P":
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filters.Priorities = append(filters.Priorities, args[i+1])
				i++
			}
		default:
			remaining = append(remaining, arg)
		}
	}

	return filters, remaining
}
