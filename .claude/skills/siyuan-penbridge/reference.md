# PenBridge reference

Complete reference for the `penbridge` CLI and the SiYuan kernel API it wraps.

## Contents

- [Global flags](#global-flags)
- [Configuration & auth](#configuration--auth)
- [Output formats](#output-formats)
- [Generic API access](#generic-api-access)
- [notebook](#notebook)
- [doc](#doc)
- [block](#block)
- [attr](#attr)
- [sql / flush](#sql--flush)
- [search](#search)
- [file](#file)
- [export](#export)
- [template](#template)
- [notify](#notify)
- [proxy](#proxy)
- [tx (transactions)](#tx-transactions)
- [system / config](#system--config)
- [SQL database schema](#sql-database-schema)
- [Block types & subtypes](#block-types--subtypes)
- [Key endpoint parameters](#key-endpoint-parameters)

## Global flags

Available on every command:

| Flag | Meaning |
|------|---------|
| `--base-url <url>` | Kernel API base URL (default `http://127.0.0.1:6806`) |
| `--token <tok>` | API token; sent as `Authorization: Token <tok>` |
| `--user` / `--password` | HTTP Basic auth (publish service / reverse proxy) |
| `--config <path>` | Config file path |
| `--timeout <sec>` | Request timeout (default 120) |
| `--insecure` | Skip TLS verification |
| `-o, --output <fmt>` | `pretty` (default) / `json` / `raw` / `envelope` |
| `--dry-run` | Print the request as JSON; do not send |
| `-v, --verbose` | Log request/response to stderr (token redacted) |

## Configuration & auth

Resolution order (highest wins): **flags > env vars > config file > defaults**.

Env vars: `PENBRIDGE_BASE_URL` (or `SIYUAN_API_URL`), `PENBRIDGE_TOKEN` (or
`SIYUAN_TOKEN`), `PENBRIDGE_USER`, `PENBRIDGE_PASSWORD`, `PENBRIDGE_TIMEOUT`,
`PENBRIDGE_INSECURE`.

Config file (`penbridge config path` prints the location, default
`<os-config-dir>/penbridge/config.json`, perms 0600):
```bash
penbridge config set --base-url http://127.0.0.1:6806 --token <TOKEN>
penbridge config show        # secrets masked
penbridge config test        # calls system/version
```

Auth header formats accepted by the kernel: `Authorization: Token <tok>`,
`Bearer <tok>`, or `?token=<tok>`. On localhost with an empty access-auth-code,
requests are allowed without a token. The publish service is read-only + Basic
auth; writes will fail there.

## Output formats

- `pretty` (default): indented JSON of the `data` field; a bare string is printed
  unquoted; `null`/empty prints nothing (success).
- `json`: compact JSON of `data` (best for parsing with `ConvertFrom-Json`/`jq`).
- `raw`: a bare string value is printed unquoted; otherwise compact JSON.
- `envelope`: the full `{code, msg, data}` response.

## Generic API access

```bash
penbridge api <path> [flags]
penbridge api list [flags]
```

`<path>` accepts `block/insertBlock`, `api/block/insertBlock`, or
`/api/block/insertBlock`. Body sources merge low->high: `--data-file`, `--data`,
`--param`.

| Flag | Meaning |
|------|---------|
| `-d, --data <json>` | JSON object body string |
| `--data-file <path>` | Read JSON body from a file (`-` = stdin) |
| `-p, --param key=value` | Add one field (repeatable); value parsed as JSON if valid, else string |
| `--method <m>` | HTTP method (default POST) |
| `--raw-body` | Send `--data`/`--data-file` verbatim, no parsing/merging |

`api list` flags: `-s/--search <term>`, `-c/--category <cat>`, `--writes`,
`--json`. Flags after each path: `W`=write/mutating, `A`=admin role,
`D`=deprecated.

## notebook

| Command | Endpoint | Notes |
|---------|----------|-------|
| `notebook ls [--flashcard]` | lsNotebooks | returns `{notebooks:[...]}` |
| `notebook open <id>` | openNotebook | mount |
| `notebook close <id>` | closeNotebook | unmount |
| `notebook create <name>` | createNotebook | returns `{notebook:{id,name,...}}` |
| `notebook remove <id>` | removeNotebook | destructive |
| `notebook rename <id> --name <n>` | renameNotebook | |
| `notebook info <id>` | getNotebookInfo | |
| `notebook conf-get <id>` | getNotebookConf | |
| `notebook conf-set <id> -d '{...}'` | setNotebookConf | conf object |

## doc

| Command | Endpoint | Notes |
|---------|----------|-------|
| `doc create --notebook <id> --path <hPath> (--markdown\|--content-file)` | createDocWithMd | returns the new doc ID (string); `--id --parent-id --tags --with-math` optional |
| `doc create-daily --notebook <id>` | createDailyNote | returns `{id}` |
| `doc rename (--id\|--notebook+--path) --title <t>` | renameDocByID / renameDoc | |
| `doc remove (--id\|--notebook+--path)` | removeDocByID / removeDoc | destructive |
| `doc move (--from-ids+--to-id)\|(--from-paths+--to-notebook+--to-path)` | moveDocsByID / moveDocs | |
| `doc get <id> [--mode --size]` | getDoc | rendered block DOM; mode 0=root,1=up,2=down,3=both,4=tail |
| `doc list --notebook <id> [--path /]` | listDocsByPath | `path` is a `.sy` path (`/` for root) |
| `doc tree --notebook <id> [--path /]` | listDocTree | nested `{id, children}` |
| `doc duplicate <id>` | duplicateDoc | |
| `doc hpath <id>` | getHPathByID | human path (string) |
| `doc path <id>` | getPathByID | `{notebook, path}` |
| `doc ids --notebook <id> <hPath>` | getIDsByHPath | `[ids]` |

Note: `doc create --path` is an **hPath** (title path). `doc rename/remove
--path` and `doc list/tree --path` are **`.sy` storage paths**.

## block

Content is supplied via exactly one of: `--markdown <text>`, `--dom <html>`,
`--data <text> --data-type <markdown|dom>`, or `--content-file <path>` (`-` =
stdin; recommended for long/multi-line content). Default dataType is `markdown`.

| Command | Endpoint | Notes |
|---------|----------|-------|
| `block insert (--previous-id\|--next-id\|--parent-id) <content>` | insertBlock | one position flag |
| `block append <parent-id> <content>` | appendBlock | last child |
| `block prepend <parent-id> <content>` | prependBlock | first child |
| `block update <id> <content>` | updateBlock | replaces the whole block |
| `block delete <id>` | deleteBlock | destructive |
| `block move <id> [--parent-id] [--previous-id]` | moveBlock | |
| `block fold <id>` / `block unfold <id>` | foldBlock / unfoldBlock | headings/lists |
| `block get <id> [--mode md\|textmark]` | getBlockKramdown | `{id, kramdown}` - read before editing |
| `block info <id>` | getBlockInfo | `{box,path,rootID,rootTitle,...}` |
| `block dom <id>` | getBlockDOM | `{id, dom}` |
| `block children <id>` | getChildBlocks | direct children |
| `block exist <id>` | checkBlockExist | bool |
| `block breadcrumb <id>` | getBlockBreadcrumb | |
| `block sibling <id>` | getBlockSiblingID | `{parent,next,previous}` |
| `block refs <id>` | getRefIDs | backlink def IDs |

## attr

| Command | Endpoint | Notes |
|---------|----------|-------|
| `attr get <id>` | getBlockAttrs | `map[name]value` |
| `attr set <id> (--attr k=v ... \| --remove k ... \| -d '{...}')` | setBlockAttrs | custom attrs must be prefixed `custom-`; null value deletes |
| `attr batch-get --ids a,b` | batchGetBlockAttrs | |

## sql / flush

```bash
penbridge sql "SELECT id,content FROM blocks WHERE type='p' LIMIT 5"
penbridge sql --file query.sql        # or stdin: --file -
penbridge flush                       # persist queued writes before querying
```

`sql` calls `/api/query/sql` (read-only; results limited by the server's search
limit). Always `flush` after writes so the indexed DB is current.

## search

| Command | Endpoint | Notes |
|---------|----------|-------|
| `search fulltext <query> [--page --page-size --method --order-by --group-by --paths]` | fullTextSearchBlock | method 0=keyword,1=query,2=SQL,3=regex |
| `search asset <keyword> [--exts .pdf,.png]` | searchAsset | |
| `search replace --find <t> --replace <t> --ids a,b` | findReplace | destructive |

## file

Workspace-relative paths (e.g. `/data/...`, `/assets/...`, `/conf/...`).

| Command | Endpoint | Notes |
|---------|----------|-------|
| `file get <path> [--out f]` | getFile | raw bytes to stdout or `--out` |
| `file put <path> (--file <local>\|--content-file -)` | putFile | multipart upload |
| `file put-dir <path>` | putFile | create directory |
| `file remove <path>` | removeFile | destructive |
| `file rename <path> <newPath>` | renameFile | |
| `file readdir <path>` | readDir | `[{name,isDir,isSymlink,updated}]` |
| `file copy <src> <dest>` | copyFile | src=asset path, dest=absolute path |

## export

| Command | Endpoint | Notes |
|---------|----------|-------|
| `export md <id> [--out f]` | exportMdContent | `{hPath, content}`; `--out` writes the Markdown |
| `export md-zip --ids a,b` | exportMds | `{name, zip}` |
| `export resources --paths a,b [--name n]` | exportResources | `{path}` |

## template

| Command | Endpoint | Notes |
|---------|----------|-------|
| `template render --path <ws path> --id <block id> [--preview]` | render | `{path, content}` |
| `template sprig <template-string>` | renderSprig | rendered string |

## notify

| Command | Endpoint |
|---------|----------|
| `notify msg <text> [--timeout ms]` | pushMsg |
| `notify err <text> [--timeout ms]` | pushErrMsg |

## proxy

`penbridge proxy <url>` -> `/api/network/forwardProxy` (server-side HTTP fetch;
private/loopback addresses are blocked by SiYuan for SSRF safety).

Flags: `-X/--method`, `-H/--header 'K: V'` (repeatable), `--payload`,
`--payload-file`, `--content-type`, `--payload-encoding`
(`json|text|base64|base64-url|base32|hex`), `--response-encoding`,
`--request-timeout` (ms). With `--payload-encoding json` (default) the payload is
parsed as JSON.

## tx (transactions)

`penbridge tx` -> `/api/transactions`, the low-level batch editing primitive.
Input via `-d/--data` or `--data-file` (`-` = stdin), as either:

- a transactions array (auto-wrapped with a generated `reqId`), or
- a full object `{"transactions":[...], "reqId":<ms>}`.

A transaction is `{"doOperations":[op,...], "undoOperations":[...]}`. An operation
is `{"action": "...", "id": "...", "data": "...", "parentID": "...",
"previousID": "...", "nextID": "..."}`. `data` is usually block **DOM**.

Common actions: `insert`, `update`, `delete`, `move`, `appendInsert`,
`prependInsert`, `foldHeading`, `unfoldHeading`, `setAttrs`. Prefer the typed
`block` commands for everyday edits; use `tx` for atomic multi-step edits or
attribute-view (database) operations. Run `penbridge api list -c av` for the
database actions.

## system / config

`system version|time|boot|conf`; `config show|set|path|test`.

## SQL database schema

Read-only tables exposed via `penbridge sql`:

- **blocks**: `id, parent_id, root_id, hash, box, path, hpath, name, alias, memo,
  tag, content, fcontent, markdown, length, type, subtype, ial, sort, created,
  updated`
- **attributes**: `id, name, value, type, block_id, root_id, box, path`
- **spans**: `id, block_id, root_id, box, path, content, markdown, type, ial`
- **refs**: `id, def_block_id, def_block_parent_id, def_block_root_id,
  def_block_path, block_id, root_id, box, path, content, markdown, type`
- **assets**: `id, block_id, root_id, box, docpath, path, name, title, hash`
- **file_annotation_refs**: `id, file_path, annotation_id, block_id, root_id,
  box, path, content, type`

Notes: `box` = notebook ID; `root_id` = the document's root block ID; `hpath` =
human path; `content` = plain text; `markdown` = Kramdown source; `ial` = the
inline attribute list (`{: id="..." ...}`). Times are `yyyymmddhhmmss` strings.

Useful queries:
```sql
-- all blocks of a document, in order
SELECT id, type, subtype, content FROM blocks WHERE root_id='<DOC>' ORDER BY sort;
-- find a document by title
SELECT id, hpath FROM blocks WHERE type='d' AND content LIKE '%Title%';
-- paragraphs containing text
SELECT id, content FROM blocks WHERE type='p' AND content LIKE '%term%';
-- a block's custom attributes
SELECT name, value FROM attributes WHERE block_id='<ID>';
```

## Block types & subtypes

`blocks.type` codes: `d` document, `h` heading, `p` paragraph, `l` list, `i` list
item, `c` code block, `b` blockquote, `t` table, `s` super block, `m` math block,
`tb` thematic break, `html` HTML block, `query_embed` embed block, `av`
attribute view (database), `audio`, `video`, `iframe`, `widget`.

`blocks.subtype`: headings `h1`..`h6`; lists/items `u` (unordered), `o`
(ordered), `t` (task).

## Key endpoint parameters

For endpoints without a typed command, call them with `penbridge api <path> -d
'{...}'`. Highlights (see `penbridge api list` for the full set of ~460):

- **createDocWithMd** `{notebook, path(hPath), markdown, id?, parentID?, tags?,
  withMath?}` -> new doc ID.
- **insertBlock** `{data, dataType, parentID?, previousID?, nextID?}`;
  **appendBlock/prependBlock** `{data, dataType, parentID}`; **updateBlock**
  `{data, dataType, id}`; **deleteBlock** `{id}`; **moveBlock** `{id, parentID?,
  previousID?}`.
- **getDoc** `{id, mode?, size?, startID?, endID?, ...}` -> `{content(DOM),
  blockCount, eof, ...}`.
- **setBlockAttrs** `{id, attrs:{k:v}}` (null deletes); **getBlockAttrs** `{id}`.
- **fullTextSearchBlock** `{query, page?, pageSize?, types?(map), paths?, method?,
  orderBy?, groupBy?}`.
- **exportMdContent** `{id, refMode?, embedMode?, yfm?, addTitle?}` -> `{hPath,
  content}`.
- **forwardProxy** `{url, method?, timeout?, headers?[], contentType?, payload?,
  payloadEncoding?, responseEncoding?}`.

Every endpoint uses POST, expects a JSON object body, and returns
`{code, msg, data}` where `code==0` means success.
