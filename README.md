# PenBridge CLI

A scriptable command-line bridge to the [SiYuan](https://github.com/siyuan-note/siyuan) (思源笔记) kernel API. Drive notebooks, documents, blocks, search, and exports from the terminal or scripts — no desktop UI required.

## Install

```bash
go install github.com/zerx-lab/penbridge-cli/cmd/penbridge@latest
```

Or build from source:

```bash
go build ./cmd/penbridge
```

## Configuration

PenBridge talks to a running SiYuan kernel (default `http://127.0.0.1:6806`). Get your API token from SiYuan → Settings → About → API token.

Settings resolve in order: **flags > environment variables > config file**.

```bash
penbridge config set --base-url http://127.0.0.1:6806 --token <your-token>
penbridge config test   # verify connectivity
```

| Env var | Description |
|---------|-------------|
| `PENBRIDGE_BASE_URL` | Kernel API URL |
| `PENBRIDGE_TOKEN` | API token |
| `PENBRIDGE_TIMEOUT` | Request timeout (seconds) |

Config file location: `%APPDATA%\penbridge\config.json` (Windows), `~/.config/penbridge/config.json` (Linux), `~/Library/Application Support/penbridge/config.json` (macOS).

## Usage

```bash
penbridge notebook ls
penbridge doc create --notebook <id> --path /notes --markdown "# Hello"
penbridge block get <id>
penbridge block update <id> --markdown "new content"
penbridge sql "SELECT * FROM blocks LIMIT 5"
penbridge search fulltext "keyword"
penbridge export md <id> --out note.md
```

### Commands

| Command | Description |
|---------|-------------|
| `notebook` | List, create, open, close, rename, remove notebooks |
| `doc` | Create, rename, move, get, list, tree documents |
| `block` | Read and edit blocks (insert, update, delete, move, fold) |
| `attr` | Get and set block attributes |
| `sql` | Run read-only SQL queries |
| `file` | Read/write workspace files and assets |
| `search` | Full-text search, asset search, find & replace |
| `export` | Export Markdown and resources |
| `template` | Render templates |
| `notify` | Push messages to the SiYuan UI |
| `system` | Kernel version, time, config |
| `proxy` | Forward HTTP requests through the kernel |
| `tx` | Low-level block transactions |
| `api` | Call any kernel endpoint directly (`api list` to browse) |
| `config` | Manage connection settings |

Run `penbridge <command> --help` for details.

### Global flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `pretty`, `json`, `raw`, `envelope` |
| `--dry-run` | Print the request without sending it |
| `-v, --verbose` | Log requests/responses to stderr |
