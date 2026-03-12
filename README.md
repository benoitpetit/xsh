# xsh

```
 ▀▄▀ ▄▀▀ █▄█
 █ █ ▄██ █ █
```

**Twitter/X CLI without API keys.**

xsh is a command-line interface for Twitter/X that works with cookie-based authentication. No developer account, no API keys, no rate limit headaches. Just log in with your browser and go.

## Features

- 🔑 **Cookie-based auth** - Use your existing browser session
- 📊 **Rich terminal output** - Beautiful formatted tweets and threads
- 🤖 **JSON mode** - Perfect for scripting and AI agents (`--json`)
- 🔍 **Search, timeline, user profiles** - Full read access
- 💬 **Post, like, retweet, bookmark** - Write operations supported
- 🏷️ **Bookmarks management** - View and manage your bookmarks

## 📦 Installation

> [📥 **Download Latest Release**](https://github.com/benoitpetit/xsh/releases/latest)

### 🚀 Quick Install (One-Liner)

<details>
<summary><strong>🪟 Windows PowerShell (AMD64)</strong></summary>

```powershell
$exe="xsh-windows-amd64.exe"; Invoke-WebRequest -Uri ((Invoke-RestMethod "https://api.github.com/repos/benoitpetit/xsh/releases/latest").assets | Where-Object name -like "*windows_amd64.exe").browser_download_url -OutFile $exe; Write-Host "Downloaded: .\$exe"
```

</details>

<details>
<summary><strong>🪟 Windows PowerShell (ARM64)</strong></summary>

```powershell
$exe="xsh-windows-arm64.exe"; Invoke-WebRequest -Uri ((Invoke-RestMethod "https://api.github.com/repos/benoitpetit/xsh/releases/latest").assets | Where-Object name -like "*windows_arm64.exe").browser_download_url -OutFile $exe; Write-Host "Downloaded: .\$exe"
```

</details>

<details>
<summary><strong>🐧 Linux AMD64</strong></summary>

```bash
curl -LO $(curl -s https://api.github.com/repos/benoitpetit/xsh/releases/latest | grep -oP 'https.*linux-amd64.tar.gz' | head -1) && tar -xzf xsh-linux-amd64.tar.gz && chmod +x xsh-linux-amd64 && sudo mv xsh-linux-amd64 /usr/local/bin/xsh && rm xsh-linux-amd64.tar.gz
```

</details>

<details>
<summary><strong>🐧 Linux ARM64</strong></summary>

```bash
curl -LO $(curl -s https://api.github.com/repos/benoitpetit/xsh/releases/latest | grep -oP 'https.*linux-arm64.tar.gz' | head -1) && tar -xzf xsh-linux-arm64.tar.gz && chmod +x xsh-linux-arm64 && sudo mv xsh-linux-arm64 /usr/local/bin/xsh && rm xsh-linux-arm64.tar.gz
```

</details>

<details>
<summary><strong>🍎 macOS Intel (AMD64)</strong></summary>

```bash
curl -LO $(curl -s https://api.github.com/repos/benoitpetit/xsh/releases/latest | grep -oP 'https.*darwin-amd64.tar.gz' | head -1) && tar -xzf xsh-darwin-amd64.tar.gz && chmod +x xsh-darwin-amd64 && sudo mv xsh-darwin-amd64 /usr/local/bin/xsh && rm xsh-darwin-amd64.tar.gz
```

</details>

<details>
<summary><strong>🍎 macOS Apple Silicon (ARM64)</strong></summary>

```bash
curl -LO $(curl -s https://api.github.com/repos/benoitpetit/xsh/releases/latest | grep -oP 'https.*darwin-arm64.tar.gz' | head -1) && tar -xzf xsh-darwin-arm64.tar.gz && chmod +x xsh-darwin-arm64 && sudo mv xsh-darwin-arm64 /usr/local/bin/xsh && rm xsh-darwin-arm64.tar.gz
```

</details>

### 📋 Via Go

```bash
go install github.com/benoitpetit/xsh@latest
```

### 🔨 Build from source

```bash
git clone https://github.com/benoitpetit/xsh
cd xsh
go build -o xsh
./xsh --help
```

## Quick Start

### 1. Login

Extract cookies from your browser (Chrome, Firefox, Brave, Edge supported):

```bash
xsh auth login --browser chrome
```

Or manually set credentials:

```bash
xsh auth set --auth-token "your_auth_token" --ct0 "your_ct0"
```

### 2. Use it

```bash
# View your timeline
xsh feed

# Search tweets
xsh search "golang"

# View a tweet thread
xsh tweet view 1234567890 --thread

# Post a tweet
xsh tweet post "Hello from xsh!"

# Like/Retweet
xsh like 1234567890
xsh retweet 1234567890

# User profile
xsh user elonmusk

# Your bookmarks
xsh bookmarks
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with browser cookies |
| `feed` | View home timeline |
| `search <query>` | Search tweets |
| `tweet view <id>` | View a tweet |
| `tweet post <text>` | Post a tweet |
| `tweet delete <id>` | Delete your tweet |
| `like <id>` | Like a tweet |
| `unlike <id>` | Unlike a tweet |
| `retweet <id>` | Retweet |
| `unretweet <id>` | Undo retweet |
| `bookmark <id>` | Bookmark a tweet |
| `unbookmark <id>` | Remove bookmark |
| `bookmarks` | View your bookmarks |
| `user <handle>` | View user profile |
| `user tweets <handle>` | View user's tweets |
| `user following <handle>` | View following list |
| `user followers <handle>` | View followers list |
| `trends` | View trending topics |
| `count <text>` | Count characters in a tweet |

## Global Flags

- `--json` - Output as JSON for scripting
- `-v, --verbose` - Show HTTP requests
- `--account <name>` - Use specific account (for multi-account)

## Authentication

xsh uses your browser's Twitter/X cookies to authenticate. The login command extracts:

- `auth_token` - Session token
- `ct0` - CSRF token

Cookies are stored securely in `~/.config/xsh/auth.json`.

### Manual Auth

If browser extraction fails, you can manually set credentials:

1. Open your browser's DevTools (F12)
2. Go to Application/Storage → Cookies → x.com
3. Copy `auth_token` and `ct0` values
4. Set credentials with:

```bash
xsh auth set --auth-token "..." --ct0 "..."
```

## JSON Mode

All commands support `--json` for machine-readable output:

```bash
xsh search "python" --json | jq '.[].text'
xsh user elonmusk --json | jq '.followers_count'
xsh feed --json | jq '.[].id'
```

Auto-detected when piping output.

## Configuration

Config file: `~/.config/xsh/config.toml`

```toml
[general]
default_count = 20
show_engagement = true
theme = "dark"

[request]
delay = 1.5
timeout = 30
max_retries = 3
```

## Development

```bash
# Run tests
go test ./...

# Build
go build -o xsh

# Run with verbose output
./xsh -v feed
```

## 404 Errors Endpoints
Occasional 404 responses can happen. x.com regularly changes the format and structure of its API to discourage automation, so endpoints may change or disappear unexpectedly. xsh provides an internal command, `xsh endpoints`, which tries to detect the current endpoints on x.com, but 404s can still occur.

> If you run into a 404, please open an issue — I will investigate and update the retry logic and the dynamic endpoint discovery/recovery as needed.

## Disclaimer

This tool is for educational purposes. Use at your own risk. Respect Twitter/X's Terms of Service.

## License

MIT
