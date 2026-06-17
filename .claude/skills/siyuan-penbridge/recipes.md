# PenBridge recipes

Concrete, copy-paste workflows for editing a SiYuan workspace with `penbridge`.
All examples assume `penbridge` is on PATH (or use `./penbridge.exe`). Replace
`<...>` placeholders with real IDs you discovered.

**Related**: [SKILL.md](SKILL.md) for the quick start and core loop, [reference.md](reference.md) for the full command and schema reference.

## Contents

- [Golden rule: discover IDs first](#golden-rule-discover-ids-first)
- [Read a whole document](#read-a-whole-document)
- [Edit a specific paragraph (by its text)](#edit-a-specific-paragraph-by-its-text)
- [Edit the Nth block of a document](#edit-the-nth-block-of-a-document)
- [Replace a heading and add a paragraph under it](#replace-a-heading-and-add-a-paragraph-under-it)
- [Add content to a document found by title](#add-content-to-a-document-found-by-title)
- [Append to today's daily note](#append-to-todays-daily-note)
- [Bulk edit multiple blocks](#bulk-edit-multiple-blocks)
- [Move and delete blocks](#move-and-delete-blocks)
- [Block attributes](#block-attributes)
- [Atomic multi-step edit (transactions)](#atomic-multi-step-edit-transactions)
- [Preview without sending](#preview-without-sending)

## Golden rule: discover IDs first

Block/document IDs are real, time-based values. Never invent them. Discover them
with `notebook ls`, `doc tree`, `doc ids`, or `sql`, and `flush` after writes so
SQL sees them.

```bash
penbridge notebook ls -o json
penbridge doc ids --notebook <NB> "/Projects/Spec"   # hPath -> [doc IDs]
penbridge doc tree --notebook <NB> --path /          # tree of {id, children}
```

## Read a whole document

```bash
# As Markdown (best for understanding/editing):
penbridge export md <DOC_ID> -o json          # {hPath, content}
penbridge export md <DOC_ID> --out doc.md     # write Markdown to a file

# As a block list (to get per-block IDs to edit):
penbridge flush
penbridge sql "SELECT id,type,subtype,content FROM blocks WHERE root_id='<DOC_ID>' ORDER BY sort" -o json
```

## Edit a specific paragraph (by its text)

```bash
penbridge flush
ID=$(penbridge sql "SELECT id FROM blocks WHERE root_id='<DOC_ID>' AND content LIKE '%old phrase%' LIMIT 1" -o json | jq -r '.[0].id')
penbridge block get "$ID" -o json             # read current source first
penbridge block update "$ID" --markdown "The corrected paragraph text."
penbridge flush
```

PowerShell variant for extracting the ID:
```powershell
$ID = (penbridge sql "SELECT id FROM blocks WHERE root_id='<DOC_ID>' AND content LIKE '%old phrase%' LIMIT 1" -o json | ConvertFrom-Json)[0].id
penbridge block update $ID --markdown "The corrected paragraph text."
```

## Edit the Nth block of a document

`sort` reflects document order. To edit the 3rd content block:
```bash
penbridge flush
ID=$(penbridge sql "SELECT id FROM blocks WHERE root_id='<DOC_ID>' AND type!='d' ORDER BY sort LIMIT 1 OFFSET 2" -o json | jq -r '.[0].id')
penbridge block update "$ID" --content-file - --data-type markdown <<'MD'
Rewritten third block. Multiple lines are fine here.

- bullet one
- bullet two
MD
```

## Replace a heading and add a paragraph under it

```bash
penbridge flush
H=$(penbridge sql "SELECT id FROM blocks WHERE root_id='<DOC_ID>' AND type='h' AND content='Old Heading' LIMIT 1" -o json | jq -r '.[0].id')
penbridge block update "$H" --markdown "## New Heading"
penbridge block insert --previous-id "$H" --markdown "Intro paragraph beneath the heading."
penbridge flush
```

## Add content to a document found by title

```bash
penbridge flush
DOC=$(penbridge sql "SELECT id FROM blocks WHERE type='d' AND content='Meeting Notes' LIMIT 1" -o json | jq -r '.[0].id')
penbridge block append "$DOC" --content-file - --data-type markdown <<'MD'
## 2026-06-16

- Decision: ship v1
- Action: write docs
MD
```

If the document does not exist, create it:
```bash
NEW=$(penbridge doc create --notebook <NB> --path "/Meetings/2026-06-16" --markdown "# Meeting Notes" -o raw)
```

## Append to today's daily note

```bash
penbridge api block/appendDailyNoteBlock -p notebook=<NB> -p dataType=markdown -p data="- Quick capture"
# or create/open today's daily note document and append to it:
DN=$(penbridge doc create-daily --notebook <NB> -o json | jq -r '.id')
penbridge block append "$DN" --markdown "Captured at $(date)."
```

## Bulk edit multiple blocks

Loop over query results (bash):
```bash
penbridge flush
for id in $(penbridge sql "SELECT id FROM blocks WHERE type='p' AND content LIKE '%TODO%'" -o json | jq -r '.[].id'); do
  cur=$(penbridge block get "$id" -o json | jq -r '.kramdown')
  penbridge block update "$id" --markdown "${cur/TODO/DONE}"
done
penbridge flush
```

For atomic bulk updates, prefer one `tx` call or `api block/batchUpdateBlock`
with `{"blocks":[{"id","data","dataType"}]}`.

## Move and delete blocks

```bash
penbridge block move <ID> --parent-id <NEW_PARENT> --previous-id <SIBLING_AFTER>
penbridge block move <ID> --previous-id <SIBLING>     # reorder within a parent
penbridge block delete <ID>                            # destructive; verify first
```

## Block attributes

```bash
penbridge attr get <ID> -o json
penbridge attr set <ID> --attr custom-status=review --attr custom-owner=alice
penbridge attr set <ID> --remove custom-status        # delete an attribute
penbridge attr set <ID> -d '{"alias":"My Alias","memo":"note"}'
```
Custom attributes must start with `custom-`. Built-in ones include `name`,
`alias`, `memo`, `bookmark`.

## Atomic multi-step edit (transactions)

Use `tx` when several edits must apply together. `data` is block **DOM**.
```bash
penbridge tx --data-file - <<'JSON'
[
  {"doOperations":[
    {"action":"update","id":"<ID1>","data":"<div data-type=\"NodeParagraph\">...</div>"},
    {"action":"delete","id":"<ID2>"}
  ]}
]
JSON
```
Get the current DOM to base edits on with `penbridge block dom <ID>`. For most
edits the typed `block` commands (Markdown) are simpler and preferred.

## Preview without sending

Append `--dry-run` to any command to print the exact request (method, URL,
redacted auth, body) without sending it — ideal before destructive edits:
```bash
penbridge block delete <ID> --dry-run
penbridge block update <ID> --content-file - --data-type markdown --dry-run <<'MD'
new content
MD
```
Use `--verbose` to additionally log the HTTP response to stderr.
