# orbital

macOS disk cleanup tool for developers. Knows what caches are safe, what's not, and what's silently eating your disk.

![](https://img.shields.io/badge/version-0.2.1-blue)
![](https://img.shields.io/badge/go-1.26%2B-00ADD8)
![](https://img.shields.io/badge/platform-macOS-lightgrey)

## Why

General disk cleaners show you a treemap but can't tell the difference between:

- `~/.npm` (4 GB, safe to delete) vs `~/.ssh` (never touch)
- Chrome **cache** vs Chrome **profile data** (bookmarks, passwords)
- A Claude sandbox VM vs your actual project files

Orbital knows. It was built from a real audit of a developer MacBook and understands every stack.

## Commands

```bash
orbital scan      # Full audit — finds everything reclaimable
orbital size      # Quick disk space check
orbital hogs      # Top 20 space consumers in ~
orbital git-trap  # Check for accidental .git in home
orbital clean     # Interactive cleanup (coming soon)
orbital serve     # Web dashboard (coming soon)
```

## Install

```bash
# Build from source
git clone https://github.com/srmdn/orbital.git
cd orbital
go build -o /usr/local/bin/orbital ./cmd/orbital

# Homebrew (coming soon)
brew install srmdn/tap/orbital
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
| Dashboard | Astro (planned) |
| Dist | Homebrew tap |

## Reference

[Full cleanup guide](docs/cleanup-guide.md) — every cache, what it is, and how to handle it.

## License

MIT
