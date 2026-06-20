package commands

var helpAdd = CommandHelp{
	Name:        "add",
	Description: "Add a new task",
	Usage: []string{
		"recur add \"Task name\"",
		"recur add \"Task name @(date time, frequency, end) #tag +project !priority *note\"",
		"recur add --edit                      Opens editor to add task(s)",
		"recur add --edit \"Task name #tag\"    Opens editor with prefilled content",
	},
	Options: []Option{
		{"-e, --edit", "Open $EDITOR to add task(s) (one per line)"},
		{"-h, --help", "Show this help message"},
	},
	Examples: []string{
		"recur add \"Buy groceries\"",
		"recur add \"Team meeting @(2025-11-20 15:00) +Work\"",
		"recur add \"Water plants @(today 9am, 1d) #chores\"",
		"recur add \"Weekly review @(Friday 5pm, 1w, 2025-12-31) +Personal\"",
		"recur add \"Dentist @(2025-11-20 2pm) !urgent *'Bring insurance card'\"",
		"recur add --edit",
		"recur add --edit \"Read book #leisure\"",
	},
}
