# macOS Disk Cleanup Guide

A reference for what caches exist on a developer Mac, what's safe to delete, and what's not.

> **orbital CLI**: `orbital scan` detects all 4 reclaimable tiers. `orbital clean` handles Tiers 1-2 automatically. Tiers 3-4 require app-level or manual action. Tier 5 is never touched.

---

## Tier 1 — Safe (auto-regenerates)

| Location | What it is | Recovered by |
|---|---|---|
| `~/.npm/` | npm package cache | `npm install` |
| `~/Library/pnpm/store/` | pnpm package cache | `pnpm install` |
| `~/.cache/` | System/SDK caches (pip, HuggingFace) | Next use |
| `~/go/pkg/mod/` | Go module download cache | `go mod download` |
| `~/Library/Caches/go-build/` | Go build artifacts | Next `go build` |
| `~/Library/Caches/Homebrew/` | Brew downloaded packages | Next `brew install` |
| `~/Library/Caches/SiriTTS/` | Text-to-speech voice data | Re-downloads as needed |
| Various AI/ML tool caches | Transient assistant/IDE data | On next run |

**Cleanup:**
```bash
rm -rf ~/.npm ~/.cache ~/Library/pnpm/store
go clean -modcache -cache
brew cleanup
```

---

## Tier 2 — Reinstallable (needs manual recovery)

| Location | Recovery |
|---|---|
| `~/.nvm/` | `nvm install <version>` |
| `~/.rustup/` | `rustup install stable` |
| `~/.android/` | Android Studio reinstall |
| Third-party editor installs | Reinstall editor |

---

## Tier 3 — App-level cleanup

These need to be cleared from within the app, NOT raw filesystem deletion.

### Chrome

Cache: `~/Library/Caches/Google/Chrome/` — safe to delete (clears on next page load)

Profiles: `~/Library/Application Support/Google/Chrome/` — do NOT delete. Contains bookmarks, passwords, extensions. Clear from Chrome settings instead.

### Docker Desktop

Images, containers, volumes live in `~/Library/Containers/com.docker.docker/Data/`.

```bash
docker system prune -a
```

### Telegram

Media cache lives in `~/Library/Group Containers/*.Telegram/`. Clear from Telegram → Settings → Data and Storage.

---

## Tier 4 — Manual review

### Downloads

- `*.dmg` files after installing: **safe to delete**
- `*.zip` files already extracted: **duplicate, safe to delete**
- Old projects, backups: **manual review**

### IDE editor caches

- VS Code workspaces: `~/Library/Application Support/Code/`
- Other editor support dirs: check `~/Library/Application Support/` for stale editors

---

## Tier 5 — Never touch

| Location | Why |
|---|---|
| `~/.ssh/` | SSH keys |
| `~/.gitconfig` | Git config |
| `~/.config/gh/` | GitHub auth |
| `~/.zshrc`, `~/.bashrc`, `~/.profile` | Shell config |
| `~/Library/Keychains/` | Passwords |
| `~/Library/Mobile Documents/` | iCloud Drive |
| `~/Documents/`, `~/Developer/`, `~/projects/` | Your work |
| `/System/`, `/Library/`, `/usr/` | macOS itself |

---

## .git Trap

If `~/.git` exists, someone ran `git init` in their home directory. The `.git/objects` folder silently tracks everything and can grow past 50 GB.

```bash
# Check
git -C ~ rev-parse --git-dir 2>/dev/null && echo "TRAP DETECTED"

# Fix
rm -rf ~/.git
```

Add this to `~/.zshrc` to prevent recurrence:
```bash
[ -d "$HOME/.git" ] && echo "WARNING: ~/.git exists. Run: rm -rf ~/.git"
```

---

## Quick Commands

```bash
# Safe one-liners
go clean -modcache                  # Go modules
brew cleanup                        # Homebrew downloads
rm -rf ~/.npm ~/.cache              # Package caches
rm -rf ~/Library/Caches/Google/Chrome/*/Cache  # Chrome cache

# Disk check
df -h /                             # Overview
du -sh ~/* ~/.* | sort -rh | head -20  # Top consumers
```
