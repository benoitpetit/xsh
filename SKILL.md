# xsh Skill

Twitter/X CLI tool for terminal usage and AI agents.

## Overview

xsh is a command-line interface for Twitter/X that uses cookie-based authentication. No API keys required. It provides both human-readable output and structured JSON for AI agents.

## Installation

```bash
# Via Go
go install github.com/benoitpetit/xsh@latest

# Via install script (Linux/macOS)
curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/main/core/scripts/install.sh | bash

# Download binary directly
wget https://github.com/benoitpetit/xsh/releases/latest/download/xsh-linux-amd64 -O xsh
chmod +x xsh
```

## Authentication

```bash
# Extract cookies from browser (Chrome, Firefox, Brave, Edge)
xsh auth login

# Import from Cookie Editor JSON export
xsh auth import cookies.json

# Check authentication status
xsh auth status

# List stored accounts
xsh auth accounts

# Switch default account
xsh auth switch <account_name>
```

## Core Commands

### Timeline & Search

```bash
# View timeline (For You or Following)
xsh feed
xsh feed --type following
xsh feed --count 50

# Search tweets
xsh search "golang"
xsh search "python" --type Latest  # Top, Latest, Photos, Videos
xsh search "dev" --count 50 --pages 2
```

### Tweet Operations

```bash
# View tweet with optional thread
xsh tweet view <tweet_id>
xsh tweet view <tweet_id> --thread

# Post tweet
xsh tweet post "Hello World!"
xsh tweet post "With image" --image photo.jpg
xsh tweet post "Reply" --reply-to <tweet_id>
xsh tweet post "Quote" --quote <tweet_url>

# Engagement
xsh tweet like <tweet_id>
xsh tweet unlike <tweet_id>
xsh tweet retweet <tweet_id>
xsh tweet unretweet <tweet_id>
xsh tweet bookmark <tweet_id>
xsh tweet unbookmark <tweet_id>

# Root shortcuts for undo actions
xsh unlike <tweet_id>
xsh unretweet <tweet_id>
xsh unbookmark <tweet_id>

# Delete tweet
xsh tweet delete <tweet_id>
```

### User Operations

```bash
# View user profile
xsh user <handle>
xsh user golang --json

# User tweets
xsh user tweets <handle>
xsh user tweets <handle> --replies

# User likes
xsh user likes <handle>

# Followers/Following
xsh user followers <handle>
xsh user following <handle>

# Social actions
xsh follow <handle>
xsh unfollow <handle>
xsh block <handle>
xsh unblock <handle>
xsh mute <handle>
xsh unmute <handle>
```

### Bookmarks

```bash
xsh bookmarks
xsh bookmarks --count 50
xsh bookmarks-folders
xsh bookmarks-folder <folder_id>
```

### Lists

```bash
xsh lists
xsh lists view <list_id>
xsh lists create "My List" --description "Description"
xsh lists delete <list_id>
xsh lists members <list_id>
xsh lists add-member <list_id> <handle>
xsh lists remove-member <list_id> <handle>
xsh lists pin <list_id>
xsh lists unpin <list_id>
```

### Direct Messages

```bash
xsh dm inbox
xsh dm send <handle> "Message text"
xsh dm delete <message_id>
```

### Scheduled Tweets

```bash
xsh schedule "Future tweet" --at "2026-04-01 09:00"
xsh scheduled
xsh unschedule <scheduled_tweet_id>
```

### Jobs

```bash
xsh jobs search "software engineer"
xsh jobs search "data engineer" --location "Paris"
xsh jobs search "devops" --location-type remote
xsh jobs view <job_id>
```

### Media & Export

```bash
# Download media from tweet
xsh download <tweet_id>
xsh download <tweet_id> --output-dir ./media

# Export tweets
xsh export feed --format json --output tweets.json
xsh export search "golang" --format csv --output results.csv
```

## Output Formats

All commands support structured output:

```bash
--json     # JSON output
--yaml     # YAML output
--compact  # Compact JSON (essential fields only)
--verbose  # Show HTTP requests
```

## JSON Schemas

### Tweet

```json
{
  "id": "string",
  "text": "string",
  "author_handle": "string",
  "author_name": "string",
  "author_id": "string",
  "author_verified": true,
  "created_at": "ISO8601_string",
  "engagement": {
    "likes": 0,
    "retweets": 0,
    "replies": 0,
    "views": 0,
    "bookmarks": 0,
    "quotes": 0
  },
  "media": [
    {
      "type": "photo|video",
      "url": "string"
    }
  ],
  "tweet_url": "string",
  "is_reply": false,
  "is_quote": false,
  "is_retweet": false
}
```

### User

```json
{
  "id": "string",
  "handle": "string",
  "name": "string",
  "bio": "string",
  "followers_count": 0,
  "following_count": 0,
  "tweets_count": 0,
  "verified": true,
  "protected": false,
  "location": "string",
  "website": "string",
  "created_at": "ISO8601_string"
}
```

## MCP Server

Start the Model Context Protocol server:

```bash
xsh mcp
```

Configuration for Claude Desktop (`claude_desktop_config.json`):

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

### Available MCP Tools (52 total)

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

## Configuration

Location: `~/.config/xsh/config.toml`

```toml
default_count = 20

[display]
theme = "default"
show_engagement = true
show_timestamps = true
max_width = 100

[request]
delay = 1.5
timeout = 30
max_retries = 3

[filter]
likes_weight = 1.0
retweets_weight = 1.5
replies_weight = 0.5
bookmarks_weight = 2.0
views_log_weight = 0.3
min_score = 0
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error |
| 3 | Rate limit |

## Best Practices

1. **Rate Limiting**: Add delays between operations with `--delay` or config
2. **Batch Operations**: Use `tweets` and `users` for multiple items
3. **JSON Piping**: Chain with jq for data processing
4. **Account Management**: Use named accounts for multiple profiles
5. **Endpoint Health**: Run `xsh doctor` if issues occur

## Examples for AI Agents

```bash
# Get latest tweets and process with jq
xsh feed --json | jq '.[].text'

# Search and filter by engagement
xsh search "golang" --json | jq '[.[] | select(.engagement.likes > 100)]'

# Export user tweets for analysis
xsh user tweets golang --json > golang_tweets.json

# Batch get user profiles
xsh users user1 user2 user3 --json | jq '.[].followers_count'

# Monitor timeline continuously
while true; do xsh feed --json | jq '.[0].text'; sleep 60; done
```

## Troubleshooting

```bash
# Check system health
xsh doctor

# Check endpoint status
xsh endpoints status

# Refresh endpoints from X.com
xsh endpoints refresh

# Auto-update obsolete endpoints
xsh auto-update
```

## Security

- Credentials stored with 0600 permissions
- TLS fingerprinting (uTLS) for bot detection avoidance
- No data sent to third parties
- Local configuration only
