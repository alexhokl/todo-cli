# todo-cli

A todo management application written in Go. A single binary acts as both a gRPC
server and a CLI client. Records are scoped per user; when Tailscale
authentication is enabled, each caller sees only their own todos, lists, and
labels.

## Installation

```sh
go install github.com/alexhokl/todo-cli@latest
```

Or, from a checkout:

```sh
task install
```

## Usage

### Server

```sh
todo serve --port 8080 --database ~/.todo.db
```

The server exposes gRPC server reflection, so it can be inspected with
`grpcurl`:

```sh
grpcurl -plaintext localhost:8080 list
```

### Authentication

By default the server runs without authentication, attributing all data to a
single dummy user. To require Tailscale identity, run the server with
`--hostname` and a Tailscale auth key:

```sh
todo serve --hostname todo --ts-auth-key tskey-auth-... --port 443 --database ~/.todo.db
```

The server runs an in-process Tailscale node (via `tsnet`) that terminates the
Tailscale connection inside the process. It resolves each caller's identity via
`WhoIs`, creates a `User` on first sight, and caches the peer-IP-to-user
mapping. Every todo, list, and label is scoped to the owning user; cross-user
access is reported as not found. The service is accessible at
`https://todo.<tailnet>.ts.net` once the node is approved (see the
[Tailscale Services docs](https://tailscale.com/docs/features/tailscale-services)).

### Client

Client commands are grouped by verb:

```sh
todo list --service todo.example.com:443
todo get --service todo.example.com:443
todo create --service todo.example.com:443
todo update --service todo.example.com:443
todo delete --service todo.example.com:443
```

Use `--insecure` to dial without TLS. Loopback addresses (`localhost`,
`127.0.0.1`, `[::1]`) are dialled insecurely by default.

## Configuration

Values are resolved in this order of precedence:

1. Command-line flag
2. Environment variable, prefixed with `TODO_` (e.g. `TODO_SERVICE`)
3. Config file, `$HOME/.todo.yaml` by default (override with `--config`)

| Setting        | Flag                | Environment variable  | Applies to |
| ----------     | ------------------- | --------------------  | ---------- |
| Service URI    | `--service`, `-s`   | `TODO_SERVICE`        | client     |
| Insecure       | `--insecure`, `-i`  | `TODO_INSECURE`       | client     |
| Verbose        | `--verbose`, `-v`   | `TODO_VERBOSE`        | all        |
| Config file    | `--config`          | `TODO_CONFIG`         | all        |
| Port           | `--port`, `-p`      | `TODO_PORT`           | `serve`    |
| Database       | `--database`, `-d`  | `TODO_DATABASE`       | `serve`    |
| Hostname       | `--hostname`        | `TODO_HOSTNAME`       | `serve`    |
| TS auth key    | `--ts-auth-key`     | `TODO_TS_AUTH_KEY`    | `serve`    |
| TS state dir   | `--ts-state-dir`    | `TODO_TS_STATE_DIR`   | `serve`    |

Example `$HOME/.todo.yaml`:

```yaml
service: todo.example.com:443
port: 443
database: /var/lib/todo/todo.db
hostname: todo
ts_state_dir: /home/appuser/tailscale
```

### Observability

The server is instrumented with OpenTelemetry (traces, metrics, and logs) and is
configured entirely through the standard `OTEL_*` environment variables:

```sh
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_SERVICE_NAME=todo
todo serve
```

## Development

```sh
task build      # build
task test       # run tests
task coverage   # run tests with coverage
task lint       # golangci-lint
task sec        # gosec
task proto      # regenerate protobuf Go code
```

See [AGENTS.md](AGENTS.md) for architecture and code style details.
