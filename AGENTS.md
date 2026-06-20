# Recur — Agent Guide

## Quick start

```bash
go build -tags fts5 -o recur cmd/recur/main.go
cd tests/integration && go test -v -tags fts5 ./...
```

## Critical: `-tags fts5` is required

All `go build` and `go test` invocations **must** include `-tags fts5`. Without it, the SQLite driver (`mattn/go-sqlite3`) will not have FTS5 support and the binary will fail at runtime or during tests.

## Commands

| Command | What it runs |
|---|---|
| `make build` | `go build -tags fts5 -ldflags="-s -w" -o recur cmd/recur/main.go` |
| `make test` | `cd tests/integration && go test -tags fts5 ./...` |
| `make test-verbose` | same + `-v` |
| `make test-coverage` | same + `-coverprofile` |
| `make test-race` | same + `-race` |
| `make fmt` | `go fmt ./...` |
| `make vet` | `go vet ./...` |
| `make lint` | `golangci-lint run ./...` (requires `golangci-lint`) |
| `make check` | `fmt -> vet -> test` |

All tests live under `tests/integration/` (package `integration`). Run a single test:

```bash
cd tests/integration && go test -v -tags fts5 -run TestAddCommand_Simple ./...
```

## Architecture

```
cmd/recur/main.go           entrypoint (inits DB, calls commands.Execute)
internal/commands/          CLI commands: add, ls, done, rm, cp, mv, root
internal/db/db.go           SQLite operations (tasks table)
internal/models/task.go     Task struct (ID, Name, DueDate, Tag, Project, Priority, Note, RecurFrequency...)
internal/parser/            natural-language date parsing + task string extraction
internal/filter/            task filtering (date range, tags, projects, priorities, query)
internal/editor/            $EDITOR integration for bulk add/edit
internal/utils/config.go    DB path: ~/.local/share/recur/recur.db
```

Task syntax parsed left-to-right: `@(...)` → `*'note'` → `#tag` → `+project` → `!priority` → name.

## Notable conventions

- **Soft deletes**: tasks set `deleted=1`, never actually removed (except `--trash`/`--purge`).
- **Recurrence**: on `done`, `handleRecurringTask` creates the next occurrence via `CalculateNextOccurrence`.
- **Default time**: imprecise dates (e.g., "tomorrow") default to 12:00 PM.
- **No `.gitignore`**, no CI workflows, no `golangci.yml` config.
- The `recur` binary in the repo root is a build artifact, not committed.
- Only dependency outside stdlib: `github.com/mattn/go-sqlite3 v1.14.28`.
