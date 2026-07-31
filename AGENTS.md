# AGENTS.md

This file provides guidance for AI coding agents working in this repository.

## Project Overview

A todo management application written in Go, providing a single binary that acts
as both a gRPC server (`todo serve`) and a CLI client. The server uses GORM with
SQLite for persistence and is instrumented with OpenTelemetry. Records (todos,
lists, labels) are scoped per user; authentication is pluggable via a gRPC
interceptor (dummy by default, Tailscale when `--hostname` is set).

### Directory Structure

- `cmd/` -- Cobra CLI commands (both server and client commands)
- `internal/` -- gRPC server implementations, interceptors, helpers
- `database/` -- GORM models and database operations
- `proto/` -- Protobuf definitions and generated Go code

### Current State

`TodoService` is implemented in `internal/todo_server.go` (todos, moves,
completion) and `internal/label_server.go` (labels, tagging) and registered in
`getGrpcServer` in `cmd/serve.go`. `todo serve` starts a working gRPC server
with OpenTelemetry instrumentation, the authentication interceptor, the
error-logging interceptor, graceful shutdown, and server reflection enabled.

`getGrpcServer(dbConn, unaryAuth, streamAuth)` chains the auth interceptor
**before** `ErrorLoggingInterceptor` so unauthenticated calls are rejected
before any handler runs. Every handler calls `userIDFromContext(ctx)` first
(returns `codes.Unauthenticated` if missing) and threads the returned `userID`
into every `database.*` call.

The noun-group commands (`list`, `get`, `create`, `update`, `delete`) are
parent commands only; leaf commands attach to them via `<noun>Cmd.AddCommand`.

## Build, Test, and Lint Commands

Task runner: [Task](https://taskfile.dev/) (`Taskfile.yml`). All commands below
can also be run directly without `task`.

```bash
task build                    # or: go build -o /dev/null ./...
go test ./...                 # run all tests
task coverage                 # or: go test --cover ./...
go test -run TestRequireSecureConnection ./cmd/            # single test
go test -run TestRequireSecureConnection/localhost ./cmd/  # single subtest
go test -v -run TestRequireSecureConnection ./cmd/         # verbose
task lint                     # golangci-lint run
task sec                      # gosec ./...
task bench                    # go test -bench=. -benchmem ./...
task install                  # go install
task proto                    # generate protobuf Go code
```

## Configuration

Configuration is resolved by Viper in this precedence order: command-line flag,
then environment variable (`TODO_` prefix), then config file
(`$HOME/.todo.yaml`).

Client flags are persistent on the root command (`--service`, `--insecure`);
server flags belong to `serve` (`--port`, `--database`, `--hostname`,
`--ts-auth-key`, `--ts-state-dir`). A command is only required to supply
`--service` when it carries the annotation `requiresService: "true"` -- client
commands must set this, server commands must not.

OpenTelemetry is configured entirely through the standard `OTEL_*` environment
variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, and so on) and is
only initialised by `serve`.

## Authentication & User Scoping

By default the server runs with the dummy interceptor, which injects a fixed
`userID = 1` so the server is usable without Tailscale (e.g. local development).
When `--hostname` is set, the server starts an in-process Tailscale node (via
`github.com/alexhokl/privateserver` / `tsnet`) that terminates the Tailscale
connection inside the process. The node's `GetCallerIdentityFromRemoteIPAddress`
API resolves the caller's peer IP to a Tailscale identity, creates a `User` on
first sight, caches the peer-IP-to-user mapping in `TailscaleAddress`, and
injects the resolved `userID` into the context under `contextKeyUser{}`.
`--ts-auth-key` is required when `--hostname` is set; `--ts-state-dir` persists
the `tsnet` state across restarts.

Every `database.*` function takes a `userID uint` parameter and scopes queries
with `Where("user_id = ?", userID)`. `findTodo` and `findLabel` are scoped the
same way, so cross-user access is reported as `ErrNotFound` rather than leaking
existence. `List`, `Label`, and `Todo` each carry a `UserID` foreign key; `List`
and `Label` use per-user composite unique indexes (`idx_list_user`,
`idx_label_user`), so two users can each own a label named "work".

The Tailscale interceptor depends on `github.com/alexhokl/privateserver`, which
embeds a `tsnet.Server`. `*pserver.Server` implicitly satisfies the unexported
`callerIdentityLookup` interface; tests substitute a `fakeIdentityLookup` stub
so they do not require a live Tailscale node.

No database migration logic exists -- a fresh database is required. Existing
rows would violate the new `NOT NULL user_id` columns.

## Code Style Guidelines

### Imports

Use `goimports`-style grouping with two groups separated by a blank line:

1. Standard library (`context`, `fmt`, `log/slog`, `os`, etc.)
2. All external packages including project-internal ones, sorted alphabetically

```go
import (
    "context"
    "fmt"
    "log/slog"

    "github.com/alexhokl/todo-cli/database"
    "github.com/alexhokl/todo-cli/internal"
    "github.com/spf13/cobra"
    "google.golang.org/grpc"
)
```

### Naming Conventions

- **Files:** `snake_case.go` (e.g., `error_logging.go`, `list_todos.go`)
- **Test files:** `*_internal_test.go` for white-box tests in the same package
- **Packages:** lowercase single words (`cmd`, `internal`, `database`)
- **Exported types/functions:** `PascalCase` (`TodoServer`, `AutoMigrate`)
- **Unexported types/functions:** `camelCase` (`rootOptions`, `runServe`)
- **Constants (exported):** `PascalCase` (`AppName`, `DefaultPort`)
- **Constants (unexported):** `camelCase` (`maxMessageSize`)

### Cobra CLI Command Pattern

Each command lives in its own file in `cmd/` following this structure:

1. Options struct: `<command>Options` (unexported, e.g., `serveOptions`)
2. Options variable: `<command>Opts` (package-level, e.g., `serveOpts`)
3. Command variable: `<command>Cmd` (e.g., `serveCmd`)
4. `init()` function: registers flags, binds them to Viper, adds the command to
   its parent
5. Run function: `run<Command>` (e.g., `runServe`) with signature
   `func(cmd *cobra.Command, args []string) error`

### Error Handling

**CLI commands** -- wrap with `fmt.Errorf` using `%w` (or `%v` for non-wrapping):

```go
return fmt.Errorf("failed to read file: %w", err)
```

**gRPC server handlers** -- return gRPC status errors with appropriate codes:

```go
return nil, status.Errorf(codes.NotFound, "todo not found: %d", id)
return nil, status.Errorf(codes.Internal, "failed to query: %v", err)
return nil, status.Errorf(codes.InvalidArgument, "title is required")
```

**Deferred resource cleanup** -- suppress close errors with blank identifier:

```go
defer func() { _ = conn.Close() }()
```

**Non-critical calls** -- suppress with blank identifier:

```go
_ = viper.BindPFlag("port", flags.Lookup("port"))
```

### Logging

Use `log/slog` structured logging (not `log` or third-party loggers). Prefer the
`Context` variants inside the server so log records are correlated with the
active trace:

```go
slog.ErrorContext(ctx, "gRPC error", slog.String("method", info.FullMethod), slog.String("error", err.Error()))
slog.Warn("failed to create directory", slog.String("path", dir), slog.String("error", err.Error()))
slog.InfoContext(ctx, "gRPC server is serving", slog.Int("port", port))
```

### Testing

- Use standard `testing` package (no testify or other assertion libraries)
- Table-driven tests with `t.Run()` subtests
- Manual assertions with `t.Errorf()`
- Database tests use an in-memory SQLite database via `setupTestDB(t)`, which
  also seeds a `User{Username: "testuser"}` (id 1) matching the dummy
  interceptor's `uint(1)`. `testUserID = 1` is the convention across the
  `database` and `internal` test packages; per-user function calls take it as
  the `userID` argument

```go
func TestExample(t *testing.T) {
    tests := []struct {
        input    string
        expected bool
    }{
        {"value1", true},
        {"value2", false},
    }
    for _, test := range tests {
        t.Run(test.input, func(t *testing.T) {
            result := someFunc(test.input)
            if result != test.expected {
                t.Errorf("expected %v but got %v", test.expected, result)
            }
        })
    }
}
```

### Types and Structs

- Database models use GORM conventions with struct tags: `gorm:"not null;unique"`
- gRPC servers embed `Unimplemented*Server` for forward compatibility
- Pointer receivers for all methods on server structs
