# Recur

A CLI todo app with support for recurring tasks, written in Go.

## Installation

```bash
go build -o recur cmd/recur/main.go
```

## Usage

```bash
recur add "Task name @(date) #tag !project !priority *note"
recur ls
recur done <id>
```

See `recur help` for more information.
