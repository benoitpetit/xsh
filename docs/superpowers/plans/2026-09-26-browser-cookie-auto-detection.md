# Browser Cookie Auto-Detection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `xsh auth login` discover the active browser profiles from the host OS and extract valid `auth_token` and `ct0` cookies without requiring `--browser`.

**Architecture:** Keep browser-specific extraction in the existing `browser` package, but centralize platform-aware profile discovery into deterministic browser candidates. Each candidate owns only its own cookie databases; extraction tries existing profiles in order and returns the first valid credential pair. Cookie selection prefers the exact X/Twitter host so duplicate cookie names cannot overwrite essential credentials unpredictably.

**Tech Stack:** Go 1.24+, `runtime`, `os.UserConfigDir`, `database/sql`, SQLite, existing CGO/browser decryptors, Go standard tests.

**Spec:** In-chat design approved by the user on 2026-09-26; root cause is the Linux `google-chrome-stable` executable being ignored by auto-detection while `--browser chrome` searches Chrome cookie files directly.

## Global Constraints

- Preserve the existing `ExtractFromBrowser`, `ExtractFromAllBrowsers`, and CLI flag interfaces.
- Detect cookie databases by browser profile paths; executable presence is only diagnostic and must not be required for extraction.
- Support Chrome, Brave, Edge, Chromium, and Firefox on Linux, macOS, and Windows.
- Require both `auth_token` and `ct0` before accepting credentials.
- Do not print cookie values or add new dependencies.

## Review Focus

- Linux installations using `google-chrome-stable` rather than `google-chrome` — auto-detection must include Chrome profiles.
- Chromium-family profiles stored under `Network/Cookies` and older `Cookies` layouts — both must be searched.
- Multiple profiles and duplicate cookie names across `x.com`/subdomains — exact hosts must win deterministically.
- Firefox profiles using `.default-esr` or another suffix — any profile containing `cookies.sqlite` must be eligible.
- A browser binary missing from `PATH` while its profile database exists — extraction must still be attempted.

### Task 1: Centralize platform-aware browser profile discovery

**Files:**
- Create: `browser/discovery.go`
- Modify: `browser/chromium_paths.go`
- Modify: `browser/extractor.go`
- Modify: `browser/firefox.go`
- Test: `browser/discovery_test.go`
- Test: `browser/paths_test.go`

**Interfaces:**
- Produce `type BrowserCandidate struct { Name string; CookiePaths []string }`.
- Produce `DiscoverBrowserCandidates() []BrowserCandidate` and make `ListAvailableBrowsers()` derive names from it.
- Produce deterministic helpers for Chromium user-data roots and Firefox profile roots that accept an explicit OS/home/config input so tests do not depend on the host OS.

- [ ] Write failing tests for `google-chrome-stable`-style Chrome profile discovery, per-browser path isolation, all Chromium profile layouts, and Firefox profiles with arbitrary suffixes.
- [ ] Run `go test ./browser -run 'Test(Discover|Chromium|Firefox)'` and confirm the new expectations fail for the current implementation.
- [ ] Implement deterministic platform path discovery and candidate filtering based on existing cookie databases, without checking only executable names.
- [ ] Run the focused browser tests and confirm they pass.
- [ ] Run `gofmt` on changed Go files and inspect the diff.

### Task 2: Make extraction deterministic and profile-complete

**Files:**
- Modify: `browser/extractor.go`
- Modify: `browser/chrome.go`
- Modify: `browser/firefox.go`
- Test: `browser/extractor_test.go`

**Interfaces:**
- Preserve public extraction functions.
- Add an internal path-loop helper that accepts a path and extraction callback, allowing tests to prove fallback behavior without opening real cookie databases.
- Explicit browser extraction must use only paths belonging to that browser; automatic extraction must process candidates in stable priority order.

- [ ] Write failing tests proving a failed first profile falls through to a later profile and automatic extraction returns candidates in stable order.
- [ ] Run the focused extraction tests and confirm they fail before the implementation.
- [ ] Implement per-browser path loops and replace concurrent first-result selection with deterministic candidate/profile traversal.
- [ ] Run the focused extraction tests and confirm they pass.

### Task 3: Select essential cookies consistently

**Files:**
- Create: `browser/cookie_selection.go`
- Modify: `browser/chrome.go`
- Modify: `browser/firefox.go`
- Test: `browser/cookie_selection_test.go`

**Interfaces:**
- Add an internal accumulator that stores cookies by name and prefers exact `x.com`/`twitter.com` hosts over subdomains while retaining sanitized values.
- Restrict SQL host predicates to `x.com`, `twitter.com`, and their subdomains.

- [ ] Write failing tests for exact-host precedence, required `auth_token`/`ct0`, and sanitization.
- [ ] Run the focused cookie-selection tests and confirm they fail before the implementation.
- [ ] Integrate the accumulator into both SQLite extractors and keep decryption errors scoped to the selected essential cookie.
- [ ] Run the focused cookie-selection tests and confirm they pass.

### Task 4: Full verification and integration review

**Files:**
- Modify: `cmd/auth.go` only if diagnostics need to reflect the new candidate model.
- Test: existing repository test suites.

- [ ] Run `go test ./...` with the repository's required CGO/SQLite setup.
- [ ] Run `go vet ./...` if dependencies are available.
- [ ] Run a non-secret smoke check of `xsh auth login --help` and the automatic discovery path; do not print cookie values.
- [ ] Review the final diff for unintended changes and confirm `xsh auth login --browser chrome` remains compatible.
