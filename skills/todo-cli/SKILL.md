---
name: todo-cli
description: Use when the user asks to list, create, update, complete, prioritise or delete todo items, labels, efforts, blockers and comments using the `todo` CLI against a running todo-cli server.
---

# todo-cli

`todo-cli` is a Go (Cobra) command-line client for the todo gRPC server. This
skill explains how to drive the client to manage items, labels, efforts,
blockers and comments. It does **not** cover starting the server (`todo-cli serve`
is out of scope).

## Mode safety: mutating vs read-only commands

Commands under `create`, `update` and `delete` **mutate data**.

- Run them only in an action / edit-capable mode.
- Never run them in a read-only planning mode ("Plan mode", "Ask mode", ...).
  There, restrict yourself to `list` and `get`.

## Connecting to the server

Every client command needs the service URI, via flag or config:

| Flag               | Purpose                                              |
| ------------------ | ---------------------------------------------------- |
| `-s, --service`    | `host:port` of the server (required)                 |
| `-i, --insecure`   | Allow an insecure connection                         |
| `--config`         | Config file (default `$HOME/.todo-cli.yaml`)             |
| `-v, --verbose`    | Verbose output                                       |

Resolution order: flag, then environment variable (`TODO_CLI_` prefix, e.g.
`TODO_CLI_SERVICE`, `TODO_CLI_INSECURE`), then `$HOME/.todo-cli.yaml`.

```bash
export TODO_CLI_SERVICE=localhost:8080
todo-cli list items
```

The server is per-user: every record is scoped to the caller. Dates use
`YYYY-MM-DD`. Confirm any command's flags with `todo-cli <command> --help`.

## Concepts

- **Item**: has a title, optional description, optional due date, labels, at
  most one effort, blockers, comments and linked items.
- **Triaged**: an item with a place in the manual ordering. A new item is
  *untriaged* until prioritised with `update priority`.
- **Done**: completed items leave the manual ordering.
- **Label**: named tag with a colour (`#RRGGBB`); names are case-insensitive
  and stored lower-case. **Effort**: named size (e.g. `high`).

## Read-only commands (safe in any mode)

| Command                      | Notes                                                                                   |
| ---------------------------- | --------------------------------------------------------------------------------------- |
| `list items`                 | Active, triaged items in manual order. See filters below.                               |
| `get item [id]`              | Show one item in detail.                                                                |
| `list labels`                | All labels.                                                                             |
| `list efforts`               | All efforts.                                                                            |
| `list blockers [item-id]`    | Blockers on an item.                                                                    |
| `list comments [item-id]`    | Comments on an item.                                                                    |

`list items` filters:

- `--label NAME` (repeatable): item must carry **every** given label.
- `--untriaged`, `--time-sensitive` (has a due date), `--done`: select one
  bucket instead of the default view. **At most one** may be given.
- `--search TEXT`: case-insensitive substring of title or description.

## Mutating commands

### create

| Command                          | Flags                                                                                                       |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `create item [title]`            | `-d/--description`, `--due YYYY-MM-DD`, `--label` (repeatable, auto-created), `--effort NAME` (must exist), `--link ID` (repeatable) |
| `create label [name]`            | `--colour #RRGGBB`                                                                                          |
| `create effort [name]`           |                                                                                                             |
| `create blocker [item-id]`       | `--description`                                                                                             |
| `create comment [item-id]`       | `--body`                                                                                                    |

### update

| Command                        | Flags                                                                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `update item [id]`             | `--title` and `--description` (empty description clears it), `-e/--edit-description` (opens `$EDITOR`; interactive, so agents should use `--description` instead), `--add-label`, `--remove-label`, `--add-link ID`, `--remove-link ID`, `--due-date`, `--clear-due-date` |
| `update done [id]`             | `--undo` reopens (appended to end of the order)                                                                                |
| `update priority [id]`         | One of `--before ID`, `--after ID`, `--top`, `--bottom` (mutually exclusive); optional `--list ID` or `--clear-list`           |
| `update todo [id]`             | `--effort NAME` or `--clear-effort`                                                                                            |
| `update label [id]`            | `--name`, `--colour #RRGGBB`                                                                                                   |
| `update effort [id]`           | `--name`                                                                                                                       |
| `update blocker [blocker-id]`  | `--description`                                                                                                                |
| `update comment [comment-id]`  | `--body`                                                                                                                       |

### delete

`delete item|label|effort|blocker|comment [id]`

## Rules and gotchas

- `create item` appends an **untriaged** item. Triage it with
  `update priority ID --top` or `--bottom`; `--before`/`--after` need a
  triaged anchor item.
- `delete item` only works on untriaged, not-done items with no linked items
  (unlink first with `update item ID --remove-link N`). Its blockers and
  comments are deleted with it. Use `update done` for anything else.
- Efforts must exist before use: `create effort` first.
- `update item --add-label` creates unknown labels; `--remove-label` ignores
  unknown ones.
- Links are symmetric; self-links and unknown ids are rejected.
- IDs are numeric; get them from `list items` / `list labels` / `list efforts`.
- `list`, `get`, `create`, `update` and `delete` alone are parent commands;
  always add the noun.

## Common workflows

```bash
# Capture and triage a task
todo-cli create item "buy milk" --due 2026-11-01 --label errands
todo-cli list items --untriaged
todo-cli update priority 12 --top

# Re-order
todo-cli update priority 12 --after 3

# Track a blockage, then finish
todo-cli create blocker 12 --description "waiting on supplier"
todo-cli update done 12

# Review
todo-cli list items --time-sensitive
todo-cli list items --label errands --search milk
```
