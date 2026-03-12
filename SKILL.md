# xsh Skill

Twitter/X CLI tool for terminal usage and AI agents.

## Quick Reference

```bash
# Auth
xsh auth login              # Extract from browser
xsh auth import file.json   # Import cookies
xsh auth set                # Manual entry

# Read
xsh feed                    # Timeline
xsh search "query"          # Search
xsh tweet ID                # View tweet
xsh user handle             # User profile

# Write  
xsh post "text"             # Post tweet
xsh post "text" --reply-to ID
xsh like ID
xsh retweet ID
xsh bookmark ID

# JSON mode for agents
xsh feed --json
xsh search "query" --json | jq '.[].text'

# MCP Server
xsh mcp                     # Start stdio server
```

## JSON Schema (Tweet)

```json
{
  "id": "string",
  "text": "string",
  "author_handle": "string",
  "author_name": "string",
  "author_verified": true,
  "created_at": "ISO8601",
  "engagement": {
    "likes": 0, "retweets": 0, "replies": 0,
    "views": 0, "bookmarks": 0, "quotes": 0
  },
  "media": [{"type": "photo|video", "url": "string"}],
  "tweet_url": "string"
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error |
| 3 | Rate limit |

## Config

Location: `~/.config/xsh/config.toml`

```toml
default_count = 20
[display]
show_engagement = true
[request]
delay = 1.5
timeout = 30
```
