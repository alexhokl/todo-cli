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

### Inspecting the API with grpcurl

The server enables gRPC server reflection, so `grpcurl` needs no `-proto` or
`-import-path` flags — services, methods, and message schemas are discovered
from the running server. JSON request bodies use the proto field names in
`snake_case`; enum values are accepted by name (e.g. `"ITEM_VIEW_DONE"`),
and `google.protobuf.Timestamp` fields accept an RFC3339 string.

Against `localhost` the dummy authentication interceptor is active (userID=1),
so no auth metadata is required. Against a Tailscale-exposed server the
caller's identity is resolved automatically from the peer IP — again, no
auth metadata to pass.

**Discovery** (no request body):

```sh
grpcurl -plaintext localhost:8080 list                           # all services
grpcurl -plaintext localhost:8080 list item.ItemService           # RPCs of ItemService
grpcurl -plaintext localhost:8080 describe item.Item              # Item message fields
grpcurl -plaintext localhost:8080 describe item.MoveItemRequest  # see the anchor oneof
```

**Calling RPCs** (curated to exercise every field shape):

*Empty request:*

```sh
grpcurl -plaintext -d '{}' localhost:8080 item.ItemService/ListItems   # default (triaged active)
```

*Enum field:*

```sh
grpcurl -plaintext -d '{"view":"ITEM_VIEW_DONE"}' localhost:8080 item.ItemService/ListItems
grpcurl -plaintext -d '{"view":"ITEM_VIEW_UNTRIAGED"}' localhost:8080 item.ItemService/ListItems
```

*Repeated string (label filter):*

```sh
grpcurl -plaintext -d '{"labels":["urgent","work"]}' localhost:8080 item.ItemService/ListItems
```

*Scalar + repeated + optional string:*

```sh
grpcurl -plaintext -d '{"title":"ship release","description":"draft changelog","labels":["urgent"]}' localhost:8080 item.ItemService/CreateItem
grpcurl -plaintext -d '{"title":"release","effort":"high"}' localhost:8080 item.ItemService/CreateItem
```

*Well-known Timestamp (optional `due_date`):*

```sh
grpcurl -plaintext -d '{"title":"follow up","due_date":"2026-08-15T00:00:00Z"}' localhost:8080 item.ItemService/CreateItem
```

*Scalar id:*

```sh
grpcurl -plaintext -d '{"id":7}' localhost:8080 item.ItemService/GetItem
```

*Oneof anchor (MoveItem):*

```sh
grpcurl -plaintext -d '{"id":7,"top":true}' localhost:8080 item.ItemService/MoveItem        # triage to top
grpcurl -plaintext -d '{"id":7,"after_id":3}' localhost:8080 item.ItemService/MoveItem     # relative (anchor must be triaged)
```

*Bool:*

```sh
grpcurl -plaintext -d '{"id":7,"done":true}' localhost:8080 item.ItemService/SetItemDone
```

*Add/remove repeated:*

```sh
grpcurl -plaintext -d '{"id":7,"add":["urgent"],"remove":["later"]}' localhost:8080 item.ItemService/UpdateItemLabels
```

*Empty-returning RPC (DeleteLabel → `google.protobuf.Empty`):*

```sh
grpcurl -plaintext -d '{"name":"urgent"}' localhost:8080 item.ItemService/CreateLabel
grpcurl -plaintext -d '{"id":3}' localhost:8080 item.ItemService/DeleteLabel                # prints {}
```

*Comments:*

```sh
grpcurl -plaintext -d '{"item_id":7,"body":"drafted a reply"}' localhost:8080 item.ItemService/CreateComment
grpcurl -plaintext -d '{"item_id":7}' localhost:8080 item.ItemService/ListComments
```

**Output formatting:**

```sh
grpcurl -plaintext -d '{}' -format json localhost:8080 item.ItemService/ListItems    # JSON instead of text
grpcurl -plaintext -d '{}' -emit-defaults localhost:8080 item.ItemService/ListItems  # include default-valued fields
```

**TLS to a Tailscale-exposed server:**

Drop `-plaintext` and use the tailnet hostname; the connection is secured by
Tailscale's mTLS and the in-process `tsnet` node resolves the caller's
Tailscale identity from the peer IP.

```sh
grpcurl todo.<tailnet>.ts.net:443 list
grpcurl -d '{"id":7}' todo.<tailnet>.ts.net:443 item.ItemService/GetItem
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
