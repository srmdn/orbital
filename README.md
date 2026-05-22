# plong

macOS disk cleanup tool for developers. Knows what caches are safe, what's not, and what's silently eating your disk.

![](https://img.shields.io/badge/version-0.2.2-blue)
![](https://img.shields.io/badge/go-1.26%2B-00ADD8)
![](https://img.shields.io/badge/platform-macOS-lightgrey)

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

Plong knows. It was built from a real audit of a developer MacBook and understands every stack.

## Install

```bash
brew install srmdn/tap/plong
```

Or build from source:

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
plong clean     # Interactive cleanup TUI
plong serve     # Web dashboard (opens in browser)
```

### Demo

```
$ plong scan

  🧹 scanning your mac...

  ── GIT TRAP ──
  ✓ clean — no .git repos outside projects

  ── TIER 1: Safe caches · 31.6 GB ──
  ✓ ~/Library/Caches/Homebrew            7.2 GB  · rm -rf
  ✓ ~/.npm                               4.1 GB  · npm cache clean --force
  ✓ ~/Library/Caches/pip                 2.8 GB  · rm -rf
  ✓ ~/.bun/install/cache                 1.9 GB  · rm -rf
  ✓ /opt/homebrew/Caches                 1.2 GB  · rm -rf
    ... +31 more (14.4 GB)

  ── TIER 2: Reinstallable · 18.3 GB ──
  ✓ ~/.nvm                              5.6 GB  · nvm cache clear
  ✓ ~/Android                            4.2 GB  · Android Studio → SDK Manager
  ✓ ~/Library/Application Support/Code   2.1 GB  · rm -rf
  🔒 ~/.docker                           8.2 GB  · docker system prune
    ... +12 more (8.2 GB)

  ── TIER 3: App cleanup · 8.7 GB ──
  ✓ ~/Downloads/*.dmg                    1.8 GB  · review then delete
  🔒 ~/Library/Application Support/Claude 3.2 GB  · clear in Claude
    ... +5 more (3.7 GB)

  ── TIER 4: Manual review · 22.1 GB ──
  🔒 ~/Music                             12.4 GB · manual review
  🔒 ~/Movies                             5.3 GB · manual review
    ... +8 more (4.4 GB)

  ──────────────────────────────────────
  Total reclaimable: 80.7 GB
  'plong clean' — free Tiers 1-2 · 'plong hogs' — full list

  Found a missing cache? github.com/srmdn/plong/issues/new
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
