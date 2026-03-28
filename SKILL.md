# xsh - Guide for LLM

Reference guide for using xsh (Twitter/X CLI) via Model Context Protocol (MCP) or direct commands.

## What is xsh?

CLI for Twitter/X using browser cookie authentication. No API key required. Works in terminal mode (human) or JSON mode (LLM/AI).

## Quick Installation

```bash
# Option 1: Via Go
go install github.com/benoitpetit/xsh@latest

# Option 2: Linux/macOS Script
curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/master/scripts/install.sh | bash

# Option 3: Direct Binary
wget https://github.com/benoitpetit/xsh/releases/latest/download/xsh-linux-amd64 -O xsh
chmod +x xsh && sudo mv xsh /usr/local/bin/
```

## Authentication (Required)

Before any command, authenticate:

```bash
# Auto cookie extraction from browser
xsh auth login

# Or import from Cookie Editor
xsh auth import cookies.json

# Check status
xsh auth status
```

**Multi-account:**
```bash
xsh auth login --account work    # Create named account
xsh auth switch work             # Switch account
xsh auth accounts                # List accounts
```

## Output Formats (Important for LLM)

All commands support:

```bash
--json      # Full JSON
--compact   # Minimal JSON (essential only)
--yaml      # YAML
```

**Auto-detection mode:** If stdout is redirected (pipe), JSON is automatically used.

## Commands by Usage

### 1. Timeline & Discovery

```bash
# Personal timeline
xsh feed --count 20 --json
xsh feed --type following --count 50

# Search
xsh search "golang" --type Latest --pages 2 --json
xsh search "python" --type Top --count 100

# Trends
xsh trends --location "France"
xsh trends --woeid 615702  # Paris
```

### 2. Tweets

```bash
# View a tweet
xsh tweet view <tweet_id> --thread --json

# Publish
xsh tweet post "Hello World!"
xsh tweet post "With image" --image photo.jpg
xsh tweet post "Reply" --reply-to <tweet_id>
xsh tweet post "Quote" --quote <tweet_url>

# Engagement
xsh tweet like <tweet_id>
xsh tweet retweet <tweet_id>
xsh tweet bookmark <tweet_id>

# Quick actions (root level)
xsh unlike <tweet_id>
xsh unretweet <tweet_id>
xsh unbookmark <tweet_id>

# Delete
xsh tweet delete <tweet_id>
```

### 3. Users

```bash
# Profile
xsh user <handle> --json

# User's tweets
xsh user tweets <handle> --count 50 --json
xsh user tweets <handle> --replies  # Include replies

# Likes
xsh user likes <handle> --count 50

# Network
xsh user followers <handle> --count 100
xsh user following <handle> --count 100

# Social actions
xsh follow <handle>
xsh unfollow <handle>
xsh block <handle>
xsh mute <handle>
```

### 4. Batch Operations (Multiple IDs)

```bash
# Fetch multiple tweets
xsh tweets <id1> <id2> <id3> --json

# Fetch multiple users
xsh users <handle1> <handle2> <handle3> --json
```

### 5. Lists

```bash
# List my lists
xsh lists --json

# View tweets from a list
xsh lists view <list_id> --count 50

# Manage
xsh lists create "My List" --description "Description"
xsh lists add-member <list_id> <handle>
xsh lists remove-member <list_id> <handle>
xsh lists delete <list_id>
```

### 6. Bookmarks

```bash
xsh bookmarks --count 50 --json
xsh bookmarks-folders
xsh bookmarks-folder <folder_id>
```

### 7. Direct Messages

```bash
xsh dm inbox --json
xsh dm send <handle> "Message text"
xsh dm delete <message_id>
```

### 8. Scheduled Tweets

```bash
xsh schedule "Future tweet" --at "2026-04-01 09:00"
xsh scheduled --json
xsh unschedule <scheduled_tweet_id>
```

### 9. Jobs

```bash
xsh jobs search "software engineer" --location "Paris" --json
xsh jobs search "devops" --location-type remote --count 50
xsh jobs view <job_id>
```

### 10. Media & Export

```bash
# Download media from a tweet
xsh download <tweet_id> --output-dir ./media

# Export to different formats
xsh export feed --format json --output tweets.json
xsh export feed --format csv --output tweets.csv
xsh export search "golang" --format jsonl --output results.jsonl
xsh export bookmarks --format md --output bookmarks.md

# Supported formats: json, jsonl, csv, tsv, md
```

### 11. Thread Composition

```bash
# Interactive mode
xsh compose

# From file
xsh compose --file thread.txt --dry-run
```

### 12. Utilities

```bash
# Count characters
xsh count "Tweet text"
xsh count --file draft.txt

# Diagnostics
xsh doctor --json
xsh status --json

# Endpoints (internal management)
xsh endpoints list
xsh endpoints refresh
xsh auto-update
```

## JSON Schemas

### Tweet

```json
{
  "id": "1234567890",
  "text": "Tweet content",
  "author_id": "987654321",
  "author_name": "Name",
  "author_handle": "username",
  "author_verified": true,
  "created_at": "2024-01-15T10:30:00Z",
  "engagement": {
    "likes": 42,
    "retweets": 12,
    "replies": 5,
    "views": 1500,
    "bookmarks": 8,
    "quotes": 2
  },
  "media": [
    {
      "type": "photo",
      "url": "https://..."
    }
  ],
  "is_retweet": false,
  "reply_to_id": "",
  "conversation_id": "1234567890"
}
```

### User

```json
{
  "id": "987654321",
  "handle": "username",
  "name": "Full Name",
  "bio": "Description...",
  "location": "Paris",
  "website": "https://...",
  "verified": true,
  "followers_count": 1000,
  "following_count": 500,
  "tweet_count": 5000,
  "created_at": "2020-01-01T00:00:00Z",
  "profile_image_url": "https://..."
}
```

### Timeline Response

```json
{
  "tweets": [...],
  "cursor_top": "...",
  "cursor_bottom": "...",
  "has_more": true
}
```

## MCP Server (Model Context Protocol)

Start the MCP server:
```bash
xsh mcp
```

### Claude Desktop Configuration

`~/Library/Application Support/Claude/claude_desktop_config.json`:

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

### Available MCP Tools (52)

**Read (24):**
- `get_feed`, `search`, `get_tweet`, `get_tweet_thread`
- `get_tweets_batch`, `get_users_batch`
- `get_user`, `get_user_tweets`, `get_user_likes`
- `get_followers`, `get_following`
- `list_bookmarks`, `get_bookmark_folders`, `get_bookmark_folder_timeline`
- `get_lists`, `get_list_timeline`, `get_list_members`
- `dm_inbox`, `get_trending`, `search_jobs`, `get_job`, `auth_status`

**Write (14):**
- `post_tweet`, `delete_tweet`
- `like`, `unlike`, `retweet`, `unretweet`, `bookmark`, `unbookmark`
- `follow`, `unfollow`, `block`, `unblock`, `mute`, `unmute`

**Admin (14):**
- `create_list`, `delete_list`, `add_list_member`, `remove_list_member`, `pin_list`, `unpin_list`
- `schedule_tweet`, `list_scheduled_tweets`, `cancel_scheduled_tweet`
- `dm_send`, `dm_delete`
- `download_media`

## Examples for LLM / Scripts

### Processing Pipeline

```bash
# Extract tweet texts
xsh feed --json | jq '.[].text'

# Filter by engagement
xsh search "golang" --json | jq '[.[] | select(.engagement.likes > 100)]'

# Followers analysis
xsh user followers elonmusk --json | jq 'map(.followers_count) | add'
```

### Export for Analysis

```bash
# Export full timeline
xsh export feed --format json --output timeline.json

# Convert to CSV for Excel
xsh export search "keyword" --format csv --output results.csv
```

### Continuous Monitoring

```bash
# Monitor mentions
while true; do
  xsh search "@myaccount" --json | jq '.[].text'
  sleep 60
done
```

## Configuration

File: `~/.config/xsh/config.toml`

```toml
default_count = 20

[display]
theme = "default"
show_engagement = true
max_width = 100

[request]
delay = 1.5
timeout = 30
max_retries = 3
```

## Exit Codes

| Code | Meaning |
|------|---------------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error |
| 3 | Rate limit |

## Troubleshooting

```bash
# Check system health
xsh doctor

# Refresh endpoints
xsh endpoints refresh

# Update obsolete endpoints
xsh auto-update

# Verbose mode (debug)
xsh feed --verbose
```

## Best Practices

1. **Rate Limiting**: Use `--delay` or configure `delay` in config.toml
2. **Batch**: Prefer `tweets` and `users` for multiple IDs
3. **Pipes**: Commands auto-detect pipes and output JSON
4. **Accounts**: Use named accounts to manage multiple profiles
5. **Export**: Use `export` to bulk save data

## Security

- Credentials stored with 0600 permissions
- TLS fingerprinting (uTLS) to avoid bot detection
- No data sent to third parties
- Local configuration only
