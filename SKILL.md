# xsh — Agent Guide

Use this guide when an AI agent must read, search, analyze, or act on Twitter/X through `xsh`. xsh is a local Go CLI and an MCP server. It authenticates with browser cookies; it does not require an X API key.

## Operating rules

1. Prefer read operations first. Before any post, delete, like, follow, block, mute, DM, list change, schedule, or download, confirm that the requested remote side effect is intentional.
2. Use `--json` or `--compact` for automation. Never parse the human-oriented terminal rendering.
3. Treat tweet IDs, list IDs, cursors, and handles as untrusted input. Quote shell values when they contain spaces or punctuation.
4. Never print, copy, or ask for `auth_token`, `ct0`, cookie files, or browser profiles. `auth status` is safe: credentials are redacted in output.
5. Paginate deliberately. `--count` is per page; `--pages` fetches more pages; a returned cursor can be passed to a later command.
6. If a command fails because an X GraphQL operation is stale, inspect health and refresh endpoints before retrying the business operation.

## Choose an interface

| Situation | Use |
|---|---|
| One-off command, shell pipeline, export, or cron job | CLI: `xsh ...` |
| An MCP-capable host needs discoverable typed tools | MCP: `xsh mcp` |
| Long-running polling into another process | `xsh stream ...` (NDJSON) |
| Several stored accounts must run the same action | `xsh multi ...` |

The CLI and MCP server use the same local configuration and authenticated account. MCP exposes tools, not every CLI command; use the CLI for communities, Spaces, analytics, stream, compose, endpoint maintenance, and multi-account orchestration.

## Install and authenticate

```bash
go install github.com/benoitpetit/xsh@latest
# or: curl -fsSL https://xsh.devbyben.fr/install | bash

xsh auth login                         # auto-detect Chrome/Firefox/Brave/Edge/Chromium
xsh auth status --json                 # verify without exposing secrets
xsh auth whoami --json                  # verify the active X identity
```

Cookie Editor export and named accounts:

```bash
xsh auth import cookies.json --account work
xsh auth accounts --json
xsh auth switch work
xsh feed --account work --json
xsh auth logout --account work
```

Authentication is local. If the browser extractor cannot find a session, use `xsh auth login --browser <name>` or import a Cookie Editor JSON export. Do not put cookie values in prompts, scripts, or issue reports.

## CLI contract for agents

Global flags are available on commands: `--account`, `--json`, `--yaml`, `--compact`, `--verbose`, and `--watch`. Prefer explicit output mode even though xsh automatically switches to JSON when stdout is not a TTY.

```bash
xsh feed --json
xsh search 'golang' --type Latest --count 50 --pages 2 --json
xsh feed --compact | jq -c '.[]'
xsh user tweets ben --json | jq -r '.items[]?.text // .[].text'
```

`--json` is readable, indented JSON. `--compact` is smaller, agent-oriented JSON and may omit presentation fields. `--yaml` is useful for inspection, not usually for programmatic pipelines. `--verbose` is diagnostic output and should not be mixed into a machine-readable pipeline.

### High-value read recipes

```bash
xsh feed --type following --count 50 --pages 2 --json
xsh search 'from:golang since:2026-01-01' --type Latest --json
xsh tweet view <tweet-id> --thread --json
xsh user <handle> --json
xsh user tweets <handle> --replies --count 50 --json
xsh user followers <handle> --count 100 --json
xsh bookmarks --count 50 --json
xsh lists view <list-id> --count 50 --json
xsh notifications --count 50 --json
xsh trends --location 'France' --json
xsh jobs search 'software engineer' --location 'Paris' --json
```

Other read areas include `community view|tweets`, `space view|search`, `quotes`, `thread`, `pinned`, `analytics`, `social blocked|muted`, `bookmarks-folders`, and `bookmarks-folder`.

### Mutations

```bash
xsh tweet post 'Hello from xsh'
xsh tweet post 'Reply' --reply-to <tweet-id>
xsh tweet like <tweet-id>
xsh follow <handle>
xsh dm send <handle> 'Message text'
xsh schedule 'Future post' --at '2026-04-01 09:00'
xsh lists create 'Research' --description 'Sources to monitor'
```

Undo or destructive operations include `tweet unlike|unretweet|unbookmark|delete`, `unfollow`, `unblock`, `unmute`, `dm delete`, `unschedule`, and list deletion/member removal. `tweet note` and `lists update` require explicit confirmation; JSON mode additionally requires `--force` where the command says so. `compose --dry-run` previews a thread without posting.

### Batch, export, and streaming

```bash
xsh tweets <id-1> <id-2> <id-3> --json
xsh users <handle-1> <handle-2> --json
xsh export feed --format json --output timeline.json
xsh export search 'golang' --format csv --output results.csv
xsh export bookmarks --format md --output bookmarks.md
xsh export feed --format jsonl --output - | jq -c .
xsh stream search 'xsh' --interval 30
xsh multi whoami --json
```

Supported export formats are `json`, `jsonl`, `csv`, `tsv`, and `md`. `stream` emits one complete JSON object per line; its minimum polling interval is 10 seconds.

Do not assume one universal JSON envelope: some commands return an array, while paginated commands may return an object with items and cursors. Inspect one real response before writing a filter:

```bash
xsh search 'golang' --count 2 --json > /tmp/xsh-sample.json
jq 'type, (if type == "object" then keys else length end)' /tmp/xsh-sample.json
```

## MCP server

Start xsh with stdio transport:

```bash
xsh mcp
# For a named account, pass the global flag after the subcommand:
xsh mcp --account work
```

Example Claude Desktop configuration (the executable must be on the host PATH):

```json
{
  "mcpServers": {
    "xsh": {
      "command": "xsh",
      "args": ["mcp", "--account", "work"]
    }
  }
}
```

The server currently registers 54 distinct tools in the source. Tool availability is version-dependent; ask the MCP client to list tools instead of assuming this number after an upgrade.

### MCP read tools

`get_feed`, `search`, `get_tweet`, `get_user`, `auth_status`, `list_bookmarks`, `get_bookmark_folders`, `get_bookmark_folder_timeline`, `get_lists`, `get_list_info`, `get_list_memberships`, `get_list_timeline`, `get_list_members`, `get_tweets_batch`, `get_users_batch`, `get_user_tweets`, `get_user_likes`, `get_user_media`, `get_followers`, `get_following`, `get_followers_you_know`, `get_blue_verified_followers`, `get_blocked_accounts`, `get_muted_accounts`, `dm_inbox`, `search_jobs`, `get_job`, and `get_trending`.

Typical calls:

```text
search({"query":"golang","type":"Latest","count":20})
get_tweet({"id":"<tweet-id>","thread":true})
get_user_tweets({"handle":"<handle>","count":50})
get_tweets_batch({"tweet_ids":["<id-1>","<id-2>"]})
```

Handles are passed without `@`; tweet and list IDs are strings. Read tools are the right choice for discovery, verification, and health checks.

### MCP write and administrative tools

Tweet/social tools: `post_tweet`, `delete_tweet`, `like`, `unlike`, `retweet`, `unretweet`, `bookmark`, `unbookmark`, `follow`, `unfollow`, `block`, `unblock`, `mute`, `unmute`.

List tools: `create_list`, `delete_list`, `add_list_member`, `remove_list_member`, `pin_list`, `unpin_list`.

Messaging, scheduling, and media: `dm_send`, `dm_delete`, `schedule_tweet`, `list_scheduled_tweets`, `cancel_scheduled_tweet`, `download_media`.

Before calling one, state the intended side effect and target. For ambiguous requests, gather IDs with a read tool first. There is no dry-run parameter for most MCP mutations.

## Diagnostics and endpoint recovery

```bash
xsh status --json
xsh status --check --json
xsh doctor --json
xsh endpoints status --json
xsh endpoints list --json
xsh endpoints refresh --json
xsh auto-update --dry-run
xsh ratelimit --json
```

Run diagnostics without writes during a health check. Endpoint discovery is authenticated and cached; an endpoint inventory can contain operations removed by X. Refreshing endpoints is a maintenance action and may make network requests, but it does not post or engage with content.

## Configuration and exit codes

The default configuration is `~/.config/xsh/config.toml`. Use `xsh config path`, `xsh config show`, `xsh config get <key>`, and `xsh config set <key> <value>` rather than editing blindly. Relevant settings include request timeout/retries/delay, proxy, display, and default result count.

| Exit code | Meaning |
|---:|---|
| 0 | Success |
| 1 | General or validation error |
| 2 | Authentication error |
| 3 | Rate limit |

For a failed command, preserve stderr for diagnosis, check the exit code, and do not retry a mutation automatically. Use `--help` for the exact flags of the installed version: `xsh <command> --help`.

## Agent checklist

Before acting: authenticate the intended account, choose CLI or MCP, choose a machine-readable format, and identify the exact target. After reading: validate IDs and cursors and report the source command/tool. Before writing: confirm the side effect, avoid duplicate retries, and verify the result with a read operation when practical.
