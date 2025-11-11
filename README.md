# Recur

A CLI Todo app with support for recurring tasks, written in Go.

## Installation

```bash
go build -o recur cmd/recur/main.go
```

## Usage

```bash
recur add "Task name @(date) #tag !priority *note"
recur ls
recur done <id>
recur cp <id>
recur mv <id>
recur rm <id>
```

See `recur help` for more information.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
