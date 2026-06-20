package commands

var helpCp = CommandHelp{
	Name:        "cp",
	Description: "Copy/duplicate tasks",
	Usage: []string{
		"recur cp <id>                      Duplicate a task",
		"recur cp <id1> <id2> ...           Duplicate multiple tasks",
		"recur cp <id> \"New name @(date) #tag +project !priority *note\"",
		"recur cp --edit <id>               Edit task in $EDITOR before copying",
		"recur cp --edit <id1> <id2> ...    Edit multiple tasks in $EDITOR before copying",
		"recur cp --tag <tag>               Copy tasks by tag",
		"recur cp --project <proj>          Copy tasks by project",
	},
	Options: []Option{
		{"--today", "Copy tasks due today"},
		{"--tomorrow", "Copy tasks due tomorrow"},
		{"--overdue", "Copy overdue tasks"},
		{"--upcoming", "Copy upcoming tasks"},
		{"-d, --due <date>", "Copy tasks due on specific date"},
		{"--from <date>", "Copy tasks from date onwards"},
		{"--to <date>", "Copy tasks up to date"},
		{"-t, --tag <tag>", "Copy tasks with specific tag"},
		{"-p, --project <proj>", "Copy tasks in specific project"},
		{"-P, --priority <pri>", "Copy tasks with specific priority"},
		{"-q, --query <keyword>", "Copy tasks matching search"},
		{"-h, --help", "Show this help message"},
	},
	Examples: []string{
		"recur cp 1",
		"recur cp 1 2 3",
		"recur cp 5 \"Modified copy @(tomorrow) !Urgent\"",
		"recur cp --edit 1 2 3                  # Opens editor",
		"recur cp --tag work",
		"recur cp --project Home",
		"recur cp --due today",
		"recur cp --tag work --project Office",
	},
}
