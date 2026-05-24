# plong

macOS disk cleanup tool for developers. Knows what caches are safe, what's not, and what's silently eating your disk.

![](https://img.shields.io/badge/version-0.2.5-blue)
![](https://img.shields.io/badge/go-1.26%2B-00ADD8)
![](https://img.shields.io/badge/platform-macOS-lightgrey)

**Requirements:** macOS 11+ (Apple Silicon) or macOS 10.13+ (Intel). Single binary, no deps.

<pre>
  ██████╗ ██╗      ██████╗ ███╗   ██╗ ██████╗ 
  ██╔══██╗██║     ██╔═══██╗████╗  ██║██╔════╝ 
  ██████╔╝██║     ██║   ██║██╔██╗ ██║██║  ███╗
  ██╔═══╝ ██║     ██║   ██║██║╚██╗██║██║   ██║
  ██║     ███████╗╚██████╔╝██║ ╚████║╚██████╔╝
  ╚═╝     ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ 
</pre>

## Why

General disk cleaners show you a treemap but can't tell the difference between:

- `~/.npm` (4 GB, safe to delete) vs `~/.ssh` (never touch)
- Chrome **cache** vs Chrome **profile data** (bookmarks, passwords)
- A Claude sandbox VM vs your actual project files

Plong knows. Built from a real audit of a developer MacBook — understands every stack.

## Install

```bash
brew install srmdn/tap/plong
```

Formula maintained in [srmdn/homebrew-tap](https://github.com/srmdn/homebrew-tap).

Or build from source (requires Go 1.26+):

```bash
git clone https://github.com/srmdn/plong.git
cd plong
go build -o /usr/local/bin/plong ./cmd/plong
```

## Usage

```bash
plong scan      # Full audit — finds everything reclaimable
plong size      # Quick disk space check
plong hogs      # Top 20 space consumers in ~
plong git-trap  # Check for accidental .git in home
plong history   # View past scan snapshots with deltas
plong clean     # Interactive cleanup TUI
plong clean --dry-run  # Preview without deleting
plong serve     # Web dashboard (opens in browser)
```

### Demo

```
$ plong scan

  scanning your mac...

  ── TIER 1: Safe · 31.6 GB ──
    ✓  7.2 GB  Homebrew cache              Brew download cache
    ✓  4.1 GB  npm cache                    Node package manager cache [node]
    ✓  2.8 GB  System caches                pip, HuggingFace, SDK caches [system]
    ✓  1.9 GB  Go module cache              Downloaded Go modules [go]
    ✓  1.2 GB  Bun cache                    Bun package manager cache [node]
    ... +26 more (14.4 GB)
    Run 'plong hogs' for full breakdown

  ── TIER 2: Reinstallable · 18.3 GB ──
    ✓  5.6 GB  Node.js versions             Node version manager installs [node]
    ✓  4.2 GB  Android SDK                  Android SDK and emulators [mobile]
    ✓  2.1 GB  Cursor editor data           Cursor IDE data [editor]
    🔒  8.2 GB  Docker data                  Docker images, containers, volumes [docker]
    ✓  1.7 GB  Rust toolchain               Rustup toolchain installs [rust]
    ... +9 more (8.2 GB)
    Run 'plong hogs' for full breakdown

  ── TIER 3: App cleanup · 15.8 GB ──
    🔒  4.1 GB  Chrome profile               Browser profile with bookmarks & passwords [browser]
    🔒  3.2 GB  Slack cache                  Slack workspace cache [messaging]
    🔒  2.5 GB  iOS backups                  iPhone/iPad backups [apple]
    ... +3 more (6.0 GB)
    Run 'plong hogs' for full breakdown

  ── TIER 4: Manual review · 22.1 GB ──
    🔒  12.4 GB  Downloads folder             Review DMGs, zips, old files [system]
    🔒  5.3 GB  VS Code workspaces            Old workspaces — review before removing [editor]
    ... +3 more (4.4 GB)
    Run 'plong hogs' for full breakdown

  ── Code editors ──
    VS Code      45 extensions · 12 stale · 1.2 GB
    Cursor       28 extensions · 8 stale · 890 MB
    Windsurf     12 extensions · 3 stale · 340 MB
    23 stale extensions total — 2.4 GB reclaimable (Tier 2)

  ── Stale disk images ──
  5 DMGs in ~/Downloads — 3.2 GB total
  Run rm ~/Downloads/*.dmg to remove (review first)

  ── Large old files (>100 MB, 90+ days) ──
    1.2 GB  ~/Downloads/old-backup.tar.gz  (modified 2025-01-15)
    450 MB  ~/Downloads/Video.mp4          (modified 2025-03-02)

  ────────────────────────────────────────────────
  Total reclaimable: 85.3 GB

  'plong clean' — free Tiers 1-2  ·  'plong hogs' — full list  ·  docs/cleanup-guide.md

  Feedback? github.com/srmdn/plong/issues/new
```

## Cleanup Tiers

| Tier | Description | Examples |
|---|---|---|
| **1 — Safe** | Auto-regenerates | npm, bun, pip, go, Homebrew caches |
| **2 — Reinstall** | Manual recovery | nvm, rustup, Android SDK |
| **3 — App cleanup** | Clear in-app | Chrome, Telegram, Docker |
| **4 — Manual review** | Your files | Downloads, DMGs, old projects |
| **∞ — Never** | Permanent data | .ssh, .gitconfig, Keychains, iCloud |

## Stack

| Layer | Tech |
|---|---|
| Engine | Go stdlib only — single binary, no deps |
| Dashboard | Go html/template (embedded) |
| Dist | Homebrew tap |

## Reference

[Full cleanup guide](docs/cleanup-guide.md) — every cache, what it is, and how to handle it.

## License

MIT
