package commands

var helpMv = CommandHelp{
	Name:        "mv",
	Description: "Edit/modify tasks",
	Usage: []string{
		"recur mv <id> \"New name @(date) #tag +project !priority *note\"",
		"recur mv <id>               Edit task in $EDITOR",
		"recur mv <id1> <id2> ...    Edit multiple tasks in $EDITOR",
		"recur mv --tag <tag> \"...\" Edit tasks by tag",
		"recur mv --project <proj> \"...\" Edit tasks by project",
	},
	Options: []Option{
		{"--today", "Edit tasks due today"},
		{"--tomorrow", "Edit tasks due tomorrow"},
		{"--overdue", "Edit overdue tasks"},
		{"--upcoming", "Edit upcoming tasks"},
		{"-d, --due <date>", "Edit tasks due on specific date"},
		{"--from <date>", "Edit tasks from date onwards"},
		{"--to <date>", "Edit tasks up to date"},
		{"-t, --tag <tag>", "Edit tasks with specific tag"},
		{"-p, --project <proj>", "Edit tasks in specific project"},
		{"-P, --priority <pri>", "Edit tasks with specific priority"},
		{"-q, --query <keyword>", "Edit tasks matching search"},
		{"-h, --help", "Show this help message"},
	},
	Examples: []string{
		"recur mv 1 \"Updated task name\"",
		"recur mv 3 \"@(tomorrow 3pm)\"",
		"recur mv 5 \"New name +NewProject\"",
		"recur mv 1",
		"recur mv 1 2 3",
		"recur mv --tag work",
		"recur mv --tag work \"@(tomorrow)\"",
		"recur mv --due today \"!urgent\"",
		"recur mv --project Home \"+Personal\"",
	},
}
