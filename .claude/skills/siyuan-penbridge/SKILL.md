---
name: siyuan-penbridge
description: Read, create, and edit SiYuan (思源笔记) notebooks, documents, blocks and paragraphs through the penbridge CLI, a wrapper over the SiYuan kernel API. Use whenever the user wants to operate a local SiYuan workspace - create/read/update/move/delete notes and blocks, edit any paragraph, run SQL over their notes, manage notebooks, block attributes, assets, templates, exports, or call any SiYuan kernel API endpoint.
---

# PenBridge: operate SiYuan from the command line

`penbridge` is a CLI bridge to the SiYuan (思源笔记) kernel API. Use it to read and
edit any notebook, document, block, or paragraph in a local SiYuan workspace.

- Full command + flag reference: [reference.md](reference.md)
- Copy-paste editing recipes: [recipes.md](recipes.md)

## Setup (do this first)

1. Install the `penbridge` binary (once). Preferred - install onto PATH directly:
   ```bash
   go install github.com/zerx-lab/penbridge-cli/cmd/penbridge@latest
   ```
   Or build from a checkout of this repo (module root):
   ```bash
   go build -o penbridge ./cmd/penbridge      # add .exe on Windows
   ```
   `go install` puts `penbridge` in `$(go env GOPATH)/bin`; ensure that is on
   PATH. Examples below say `penbridge`.

2. Confirm connectivity:
   ```bash
   penbridge config test
   ```
   If it prints `OK: connected to ... (kernel version X)`, you are ready.

3. Auth: the kernel API defaults to `http://127.0.0.1:6806`. On localhost with no
   access-auth-code set, no token is needed. Otherwise set the API token (SiYuan:
   Settings -> About -> API token):
   ```bash
   penbridge config set --token <TOKEN>          # or env PENBRIDGE_TOKEN
   ```
   The publish service (`:6808`) is READ-ONLY (Basic auth via `--user/--password`);
   write commands fail there. For editing, target the kernel API (`:6806`).

## Core concepts

- **Notebook**: a top-level container; identified by a `notebook` ID like
  `20210808180117-czj9bvb`.
- **Document**: a note; its root is a block (`type='d'`). Identified by a block ID.
- **Block**: every paragraph, heading, list, code block, etc. is a block with a
  unique ID like `20210808180117-6v0mkxr`. Editing "a paragraph" means editing a
  block.
- **Two paths**: `hPath` is the human-readable title path (`/Projects/Notes`);
  the `.sy` path is the storage path (`/20210808.../20210809....sy`). Commands
  say which one they expect.
- **dataType**: block content is either `markdown` (Kramdown) or `dom` (HTML).
  Prefer `markdown`. The CLI defaults to `markdown`.
- **IDs are time-based**: `yyyymmddhhmmss-7rand`. Always discover real IDs (via
  SQL or list/tree commands) before editing; never invent them.

## Quick start

```bash
penbridge notebook ls -o json                 # list notebooks (get IDs)
penbridge doc create --notebook <NB> --path "/Inbox/Today" --markdown "# Today"
penbridge doc tree --notebook <NB> --path /   # see documents + IDs
penbridge block get <BLOCK_ID>                # read a block's Markdown source
```

## The editing workflow (read - modify - write)

To safely edit existing content, always locate the block first, read it, then
write. **Never guess a block ID.**

1. **Find** the target block via SQL (most precise) or `doc get`:
   ```bash
   penbridge flush                             # persist pending writes first
   penbridge sql "SELECT id,type,content FROM blocks WHERE root_id='<DOC_ID>' ORDER BY sort" -o json
   ```
2. **Read** the block's current source:
   ```bash
   penbridge block get <BLOCK_ID> -o json      # {id, kramdown}
   ```
3. **Write** the new content. `block update` replaces the WHOLE block, so include
   all of its content. For multi-line content, pipe via stdin to avoid quoting
   issues:
   ```bash
   penbridge block update <BLOCK_ID> --content-file - --data-type markdown <<'MD'
   ## New heading

   New paragraph text with **bold** and a [[wikilink]].
   MD
   ```
4. **Verify** (optional): `penbridge flush && penbridge block get <BLOCK_ID> -o json`.

Add or remove blocks instead of replacing:
```bash
penbridge block append  <PARENT_ID> --markdown "Appended paragraph."   # last child
penbridge block prepend <PARENT_ID> --markdown "First child."          # first child
penbridge block insert  --previous-id <ID> --markdown "After that block."
penbridge block insert  --next-id <ID>     --markdown "Before that block."
penbridge block move    <ID> --parent-id <P> --previous-id <SIBLING>
penbridge block delete  <ID>
```

See [recipes.md](recipes.md) for more (edit Nth paragraph, replace a heading's
section, bulk edits, daily notes, assets).

## Discovering and calling any endpoint

The typed commands cover common tasks. For anything else, every one of the ~460
kernel endpoints is reachable:

```bash
penbridge api list                      # all endpoints (W=write A=admin D=deprecated)
penbridge api list -s block --json      # filter + machine-readable
penbridge api <path> -d '{"json":...}'  # call any endpoint
penbridge api block/getBlockInfo -p id=<ID>   # build body from key=value params
```

## Common commands

| Task | Command |
|------|---------|
| List notebooks | `penbridge notebook ls -o json` |
| Create document | `penbridge doc create --notebook <NB> --path "/A/B" --markdown "..."` |
| List docs / tree | `penbridge doc list --notebook <NB> --path /` / `doc tree ...` |
| Read block source | `penbridge block get <ID> -o json` |
| Edit a block | `penbridge block update <ID> --content-file -` (stdin) |
| Append paragraph | `penbridge block append <PARENT_ID> --markdown "..."` |
| Delete block | `penbridge block delete <ID>` |
| Find blocks (SQL) | `penbridge sql "SELECT id,content FROM blocks WHERE ..." -o json` |
| Block attributes | `penbridge attr get <ID>` / `attr set <ID> --attr custom-k=v` |
| Export Markdown | `penbridge export md <ID> --out note.md` |
| Any endpoint | `penbridge api <path> -d '{...}'` |

## Safety and boundaries

- **Preview before mutating**: append `--dry-run` to print the exact request
  without sending it. Use `--verbose` to log requests/responses to stderr.
- **Flush before SQL reads** after writes: `penbridge flush` (writes are queued;
  SQL reads the indexed DB).
- **Destructive ops** (`delete`, `remove`, `findReplace`) cannot be undone via the
  API. Confirm the ID first; consider `--dry-run`.
- **Exit codes**: non-zero on any API error; the error message (code + msg) goes
  to stderr. Check it.
- **Output formats** (`-o`): `pretty` (default), `json` (compact, for parsing),
  `raw` (bare string, e.g. a new doc ID), `envelope` (full `{code,msg,data}`).
- **Read-only targets**: against the publish service or in read-only mode, write
  endpoints return an error - that is expected.

For the complete command/flag list, endpoint parameters, SQL schema, and block
types, read [reference.md](reference.md).
