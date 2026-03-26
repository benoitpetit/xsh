# xsh — Twitter/X CLI

[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-Compatible-green?style=flat-square)](https://modelcontextprotocol.io/)

A fast, feature-complete command-line interface for Twitter/X, written in Go.  
No API keys required — authenticates directly via browser cookies.

![xshwall](xshwall.png)

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Commands](#commands)
  - [Tweet Operations](#tweet-operations)
  - [Timeline & Search](#timeline--search)
  - [Users](#users)
  - [Social Actions](#social-actions)
  - [Bookmarks](#bookmarks)
  - [Direct Messages](#direct-messages)
  - [Lists](#lists)
  - [Scheduled Tweets](#scheduled-tweets)
  - [Trends](#trends)
  - [Jobs](#jobs)
  - [Media Download](#media-download)
  - [Compose Thread](#compose-thread)
  - [Export](#export)
  - [Batch Operations](#batch-operations)
  - [Character Count](#character-count)
- [Output Formats](#output-formats)
- [MCP Server](#mcp-server)
- [Configuration](#configuration)
- [Endpoint Management](#endpoint-management)
- [Diagnostics & Status](#diagnostics--status)
- [Development](#development)

---

## Features

- **Zero API keys** — cookie-based authentication extracted from your browser
- **AI-ready** — built-in MCP server compatible with Claude, Cursor, and any MCP client
- **Multi-account** — manage and switch between multiple Twitter/X accounts
- **Structured output** — JSON, YAML, and compact modes for scripting and agents
- **Auto-updating endpoints** — dynamically discovers GraphQL operation IDs from X.com JS bundles
- **Media support** — post up to 4 images per tweet, download media from any tweet
- **Full coverage** — tweets, DMs, lists, bookmarks, jobs, trends, schedule, and more

---

## Installation

### Using `go install` (Recommended)

Requires [Go 1.25+](https://go.dev/doc/install):

```bash
go install github.com/benoitpetit/xsh@latest
```

### Using Install Script

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/main/core/scripts/install.sh | bash
```

**Windows (PowerShell as Administrator):**
```powershell
iwr -useb https://raw.githubusercontent.com/benoitpetit/xsh/main/core/scripts/install.ps1 | iex
```

### Pre-built Binaries

Download from [Releases](https://github.com/benoitpetit/xsh/releases):

```bash
# Linux (amd64)
curl -L https://github.com/benoitpetit/xsh/releases/latest/download/xsh-linux-amd64 -o xsh && \
chmod +x xsh && \
sudo mv xsh /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/benoitpetit/xsh
cd xsh/core
go build -o xsh .
sudo mv xsh /usr/local/bin/
```

---

## Quick Start

```bash
# 1. Authenticate (extracts cookies from your browser)
xsh auth login

# 2. View your timeline
xsh feed

# 3. Post a tweet
xsh tweet post "Hello from xsh!"

# 4. Search tweets
xsh search "golang" --pages 2

# 5. View a user profile
xsh user elonmusk
```

---

## Authentication

### Automatic Browser Extraction

```bash
# Auto-detect browser (Chrome, Firefox, Brave, Edge, Chromium)
xsh auth login

# Target a specific browser
xsh auth login --browser firefox
xsh auth login --browser brave

# Save under a named account
xsh auth login --account work
```

### Manual / Import

```bash
# Import from Cookie Editor JSON export
xsh auth import cookies.json
xsh auth import cookies.json --account work

# Set credentials manually
xsh auth set

# Check authentication status
xsh auth status

# Show current user info
xsh auth whoami

# List all stored accounts
xsh auth accounts

# Switch default account
xsh auth switch work

# Remove stored credentials
xsh auth logout
```

> Shortcuts `xsh accounts`, `xsh switch <name>`, and `xsh import <file>` are also available at root level.

---

## Commands

### Global Flags

These flags apply to every command:

| Flag               | Short | Description                   |
| ------------------ | ----- | ----------------------------- |
| `--account <name>` |       | Use a specific stored account |
| `--json`           |       | Output as JSON                |
| `--yaml`           |       | Output as YAML                |
| `--compact`        | `-c`  | Compact output for AI agents  |
| `--verbose`        | `-v`  | Show HTTP requests (debug)    |

---

### Tweet Operations

```bash
# View a tweet (with optional thread tree)
xsh tweet view <id>
xsh tweet view <id> --thread
xsh tweet view <id> --count 50

# Post a tweet
xsh tweet post "Hello world!"
xsh tweet post "With image" --image photo.jpg
xsh tweet post "Multiple images" -i img1.jpg -i img2.jpg -i img3.jpg
xsh tweet post "Reply" --reply-to <tweet-id>
xsh tweet post "Quote" --quote https://x.com/user/status/<id>

# Delete your tweet
xsh tweet delete <id>

# Like / unlike
xsh tweet like <id>
xsh tweet unlike <id>

# Retweet / undo
xsh tweet retweet <id>
xsh tweet unretweet <id>

# Bookmark / remove
xsh tweet bookmark <id>
xsh tweet unbookmark <id>
```

**Shortcuts** available at root level:

```bash
xsh unlike <id>
xsh unretweet <id>
xsh unbookmark <id>
```

---

### Timeline & Search

```bash
# Home timeline
xsh feed
xsh feed --type following              # Following timeline
xsh feed --count 50 --pages 3         # 150 tweets total
xsh feed --filter top --top 20        # Top 20 by engagement
xsh feed --filter score --threshold 100
xsh feed --cursor <cursor>            # Paginate

# Search
xsh search "golang"
xsh search "golang" --type Latest     # Types: Top, Latest, Photos, Videos
xsh search "golang" --count 50 --pages 2
xsh search "golang" --cursor <cursor>
```

---

### Users

```bash
# View profile
xsh user <handle>

# User's tweets
xsh user tweets <handle>
xsh user tweets <handle> --replies   # Include replies
xsh user tweets <handle> --count 50

# Liked tweets
xsh user likes <handle> --count 30

# Followers / following
xsh user followers <handle> --count 50
xsh user following <handle> --count 50
```

---

### Social Actions

```bash
# Follow / unfollow
xsh follow <handle>
xsh unfollow <handle>

# Block / unblock
xsh block <handle>
xsh unblock <handle>

# Mute / unmute
xsh mute <handle>
xsh unmute <handle>
```

---

### Bookmarks

```bash
# View all bookmarks
xsh bookmarks
xsh bookmarks --count 50

# List bookmark folders
xsh bookmarks-folders

# View tweets in a specific folder
xsh bookmarks-folder <folder-id>
```

---

### Direct Messages

```bash
# View inbox
xsh dm inbox

# Send a DM
xsh dm send <handle> "Your message"

# Delete a DM message
xsh dm delete <message-id>
```

---

### Lists

```bash
# View all your lists
xsh lists

# View tweets from a list
xsh lists view <list-id>
xsh lists view <list-id> --count 50

# Create / delete a list
xsh lists create "List name"
xsh lists delete <list-id>

# Manage members
xsh lists members <list-id>
xsh lists add-member <list-id> <handle>
xsh lists remove-member <list-id> <handle>

# Pin / unpin
xsh lists pin <list-id>
xsh lists unpin <list-id>
```

---

### Scheduled Tweets

```bash
# Schedule a tweet
xsh schedule "My future tweet" --at "2026-04-01 09:00"

# List scheduled tweets
xsh scheduled

# Cancel a scheduled tweet
xsh unschedule <scheduled-tweet-id>
```

---

### Trends

```bash
# Worldwide trends
xsh trends

# By location name
xsh trends --location "Paris"
xsh trends --location "France"

# By WOEID
xsh trends --woeid 1        # Worldwide
xsh trends --woeid 615702   # Paris
```

---

### Jobs

```bash
# Search job listings
xsh jobs search "software engineer"
xsh jobs search "data engineer" --location "Paris"
xsh jobs search "devops" --location-type remote
xsh jobs search "backend" --employment-type full_time
xsh jobs search "intern" --seniority entry_level
xsh jobs search "manager" --company Google --pages 2

# Available filters:
#   --location <city/country>
#   --location-type  remote | onsite | hybrid
#   --employment-type  full_time | part_time | contract | internship
#   --seniority  entry_level | mid_level | senior
#   --company <name>
#   --industry <sector>
#   --count <n>  (default 25)
#   --pages <n>  (default 1)

# View job details
xsh jobs view <job-id>
```

---

### Media Download

```bash
# Download all media from a tweet
xsh download <tweet-id>
xsh download <tweet-id> --output-dir ./media
```

---

### Compose Thread

```bash
# Interactive mode (prompts for each tweet)
xsh compose

# From a text file (auto-splits into thread)
xsh compose --file thread.txt

# From stdin
cat thread.txt | xsh compose

# Preview without posting
xsh compose --file thread.txt --dry-run
```

---

### Export

Export tweets to file in multiple formats: `json`, `jsonl`, `csv`, `tsv`, `md`.

```bash
# Export timeline
xsh export feed --format csv --output timeline.csv
xsh export feed --format jsonl --output tweets.jsonl --count 200
xsh export feed --type following --filter top

# Export search results
xsh export search "golang" --format md --output results.md

# Export bookmarks
xsh export bookmarks --format json --output bookmarks.json

# Write to stdout
xsh export feed --format jsonl --output -
```

---

### Batch Operations

```bash
# Fetch multiple tweets by ID
xsh tweets <id1> <id2> <id3>

# Fetch multiple user profiles
xsh users <handle1> <handle2> <handle3>
```

---

### Character Count

```bash
# Count characters in text
xsh count "My tweet draft"

# From a file
xsh count --file draft.txt

# From stdin
echo "My tweet" | xsh count

# Show formatted preview
xsh count "My tweet" --preview
xsh count --file draft.txt --preview --width 80
```

---

## Output Formats

All commands support structured output for scripting and AI integration:

```bash
# JSON
xsh feed --json
xsh user elonmusk --json

# YAML
xsh search "golang" --yaml

# Compact (essential fields only, ideal for AI agents)
xsh feed --compact

# Pipe auto-detection: JSON is used automatically when stdout is not a TTY
xsh feed | jq '.[] | .text'

# Use a specific account
xsh feed --account work
```

---

## MCP Server

xsh includes a full [Model Context Protocol](https://modelcontextprotocol.io/) server for use with Claude, Cursor, and other MCP-compatible AI clients.

```bash
xsh mcp
```

### Claude Desktop Configuration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "xsh": {
      "command": "xsh",
      "args": ["mcp"]
    }
  }
}
```

### Available MCP Tools

| Category   | Tools                                                                                                                                                      |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Timeline   | `get_feed`, `get_home_latest_timeline`                                                                                                                     |
| Search     | `search`, `search_bookmarks`                                                                                                                               |
| Tweets     | `get_tweet`, `get_tweet_thread`, `post_tweet`, `delete_tweet`, `get_tweets_batch`                                                                          |
| Engagement | `like`, `unlike`, `retweet`, `unretweet`, `bookmark`, `unbookmark`                                                                                         |
| Users      | `get_user`, `get_users_batch`, `get_user_tweets`, `get_user_likes`, `get_followers`, `get_following`                                                       |
| Social     | `follow`, `unfollow`, `block`, `unblock`, `mute`, `unmute`                                                                                                 |
| Bookmarks  | `list_bookmarks`, `get_bookmark_folders`, `get_bookmark_folder`                                                                                            |
| Lists      | `get_user_lists`, `get_list_timeline`, `get_list_members`, `create_list`, `delete_list`, `add_list_member`, `remove_list_member`, `pin_list`, `unpin_list` |
| DMs        | `dm_inbox`, `send_dm`, `delete_dm`                                                                                                                         |
| Scheduled  | `schedule_tweet`, `get_scheduled_tweets`, `delete_scheduled_tweet`                                                                                         |
| Trends     | `get_trending`                                                                                                                                             |
| Jobs       | `search_jobs`, `get_job`                                                                                                                                   |
| Media      | `download_media`, `upload_media`                                                                                                                           |
| Compose    | `compose_thread`                                                                                                                                           |
| Utils      | `count_characters`, `export_tweets`                                                                                                                        |
| Endpoints  | `get_endpoints`, `refresh_endpoints`                                                                                                                       |

---

## Configuration

Configuration is stored at `~/.config/xsh/config.toml`:

```bash
# Show current configuration
xsh config
xsh config show

# Get a specific value
xsh config get filter.likes_weight

# Set a value
xsh config set display.theme dark
xsh config set request.timeout 60

# Open in editor
xsh config edit

# Show config file path
xsh config path

# Reset to defaults
xsh config reset
```

### Example `config.toml`

```toml
default_count = 20

[display]
theme           = "default"
show_engagement = true
show_timestamps = true
max_width       = 100

[request]
delay       = 1.5
timeout     = 30
max_retries = 3

[filter]
likes_weight     = 1.0
retweets_weight  = 1.5
replies_weight   = 0.5
bookmarks_weight = 2.0
views_log_weight = 0.3
min_score        = 0
```

---

## Endpoint Management

xsh dynamically discovers GraphQL operation IDs from X.com's JavaScript bundles and caches them locally.

```bash
# List all cached endpoints
xsh endpoints list

# Check status of a specific endpoint
xsh endpoints check HomeTimeline

# Refresh endpoints from X.com (no auth required)
xsh endpoints refresh

# Show endpoint system status
xsh endpoints status

# Manually update a single endpoint
xsh endpoints update <operation> <endpoint-id>

# Reset all endpoints to bundled defaults
xsh endpoints reset

# Auto-update obsolete endpoints (safe, no auth needed)
xsh auto-update
xsh auto-update --dry-run   # Check only, don't update
xsh auto-update --force     # Force refresh (ignore cache)
```

---

## Diagnostics & Status

```bash
# System status: auth, endpoints, cache health
xsh status
xsh status --check   # Run fresh endpoint health check
xsh status --json

# Full diagnostic report
xsh doctor

# Print version
xsh version
```

---

## Development

```bash
# Clone
git clone https://github.com/benoitpetit/xsh
cd xsh/core

# Build
go build -o xsh .

# Run tests
go test ./tests/...

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o xsh .
```

### Project Structure

```
core/
├── main.go           # Entry point
├── cmd/              # Cobra CLI commands
├── core/             # API client, auth, config, endpoints
├── models/           # Tweet, User, DM, etc.
├── display/          # Terminal formatters
├── utils/            # Helpers (filter, validation, article, delay, hash)
├── browser/          # Browser cookie extraction
└── tests/            # Integration tests
```

### Requirements

- Go 1.25+
- A Twitter/X account
- A supported browser (for cookie extraction)

---

## Security

- Credentials stored locally with `0600` permissions
- No data sent to third parties — direct API calls to X/Twitter only
- TLS fingerprinting (uTLS) to avoid bot detection
- Cookie values sanitized per RFC 6265 before use

---

## Documentation

- [WHITEPAPER.md](WHITEPAPER.md) — Architecture and technical design
- [CHANGELOG.md](CHANGELOG.md) — Version history
- [CONTRIBUTING.md](CONTRIBUTING.md) — Contribution guidelines

---

## License

[MIT](LICENSE)

---

**Made with ❤️ and Go**
