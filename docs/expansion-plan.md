# Plong v0.2 Expansion Plan

## Our Stack

```
Go 1.26+ · single binary · zero external deps · macOS only
Distribution: Homebrew tap (srmdn/homebrew-tap)
CLI commands: scan, clean, serve, size, hogs, git-trap
```

## Why Plong Wins

The macOS cleanup space is crowded. Here's how plong differentiates:

| Plong | Everyone else |
|---|---|
| **Safety tiers** — T1 (auto-safe) through T5 (never touch). You literally cannot delete T3/4 items through `plong clean`. Other tools: generic "safe/review" flags or nothing. | CleanMyMac hides what it deletes. mac-disk-cleaner-cli shows safe/review but no hard enforcement. |
| **Every entry explains itself** — "3.2 GB — npm cache — safe to delete, recovered by next `npm install`". Other tools: just show size. | Nobody tells you *why* it's safe or *how* to clean it if it's locked. |
| **Stack-aware** — detects your dev stack and surfaces relevant waste first. Strange dev gets discovered items with raw paths + generic hints. | Every tool scans the same static list regardless of who's running it. |
| **Open source, auditable** — every `rm -rf` is in `clean.go`. Zero telemetry. Single binary. | CleanMyMac is closed source with telemetry. CLI alternatives are open source but don't explain what they delete. |
| **`plong hogs` is the gateway drug** — fastest command, goes into dotfiles aliases and blog posts. | ncdu/dua are visualizers only, no cleanup. |
| **Non-cleanable items get cleanup instructions** — "Chrome profile → clear from chrome://settings/clearBrowserData". Not just 🔒. | Other tools: either delete it (dangerous) or show a lock icon with no next step. |

**One-liner:** plong tells you what's wasting your space, whether it's safe to delete, and exactly how to clean it — even when it can't delete it for you.

## Architecture Change: Discovery + Registry

**Current:** 28 hardcoded paths in `var tierNTargets` slices. Adding a target = editing Go code.

**After:** Discovery engine + label registry.

```
┌─────────────────────────────────────────────────────────┐
│  Collect(home)                                          │
│  ├── scanKnown()    — label registry lookups            │
│  │   Known path → {label, description, tier, cleanable, │
│  │                  cleanHint, stackTag}                  │
│  │                                                       │
│  ├── scanDiscovery() — scan dirs not in registry        │
│  │   ~/  → top-level dirs > 100MB                       │
│  │   ~/Library/Caches/* → auto-T1                       │
│  │   ~/Library/Application Support/* > 500MB → auto-T3  │
│  │   ~/Library/Developer/* → auto-T2                    │
│  │   ~/Library/Logs/* → auto-T1                         │
│  │   ~/Library/Containers/* → auto-T3 if > 200MB        │
│  │                                                       │
│  ├── classifyUnknown() — heuristic tier assignment      │
│  │   Location rule + size heuristic                     │
│  │                                                       │
│  └── filterSafety() — exclude never-touch paths        │
└─────────────────────────────────────────────────────────┘
```

**Key difference:** Unknown dirs don't get ignored. They get surfaced with a generic label, heuristic tier, and "manual review" hint. Known dirs get the full treatment — friendly name, exact cleanup command, stack tag.

## New Entry Structure

```go
type Entry struct {
    Path        string
    Label       string   // "npm cache" or "~/Library/Application Support/com.unknown.app"
    Description string   // "Node package manager cache — safe to delete"
    SizeMB      int64
    Tier        int      // 1-5
    Cleanable   bool     // true = plong clean can delete it
    CleanHint   string   // "npm cache purge" or "Clear from Chrome settings"
    StackTag    string   // "node", "go", "python", "apple", "system", "" for unknown
}
```

## Safety Exclusion List (never scanned, never shown)

```
~/.ssh/
~/.gitconfig
~/Library/Keychains/
~/Library/Mobile Documents/   (iCloud)
~/Developer/
~/Documents/
~/Desktop/                    (shown in hogs, never in scan)
~/Pictures/
~/Music/
~/Movies/
```

## Tier Definitions (refined)

| Tier | Name | plong clean | CleanHint |
|---|---|---|---|
| T1 | Safe (auto-regenerates) | ✅ deletes | "recovered by next install/use" |
| T2 | Reinstallable | ✅ deletes | "reinstall with [command]" |
| T3 | App-level cleanup required | ❌ locked | exact app action or CLI command |
| T4 | Manual review | ❌ locked | "review before removing" |
| T5 | Never touch | — (not shown) | safety exclusion list |

## Phase 1 — Architecture Refactor + Quick Wins

### 1a. File restructure

```
internal/scan/
  registry.go      ← NEW: 60+ known targets, each with label/tier/cleanHint/stackTag
  discover.go      ← NEW: auto-discovery for unknown dirs, heuristic classification
  safety.go        ← NEW: never-touch exclusion list + stack detection
  scan.go          ← REWRITE: Collect() merges known + discovered, stack sorting
  format.go        ← NEW: FormatSize moved here + FormatCleanHint helper
internal/clean/
  clean.go         ← UPDATE: show CleanHint for locked items, stack-aware display
internal/serve/
  serve.go         ← UPDATE: include cleanHint/stackTag in JSON API
  templates/index.html ← UPDATE: show CleanHint, stack badges
docs/
  cleanup-guide.md ← UPDATE: add all new targets, cleanHint commands
CLAUDE.md           ← UPDATE: command status table, new architecture notes
```

### 1b. Registry — all known targets

**T1 — Safe caches (auto-regenerate)**

| Path | Label | StackTag | CleanHint |
|---|---|---|---|
| `~/.npm/` | npm cache | node | `npm cache purge` |
| `~/.bun/` | Bun cache | node | reinstalls on next `bun install` |
| `~/.cache/` | System caches | system | various caches auto-regenerate |
| `~/Library/Caches/go-build/` | Go build cache | go | `go clean -cache` |
| `~/Library/Caches/goimports/` | Go imports cache | go | auto-regenerates |
| `~/Library/Caches/node-gyp/` | node-gyp cache | node | rebuilds on next `npm install` |
| `~/Library/Caches/Homebrew/` | Homebrew cache | system | `brew cleanup` |
| `~/Library/Caches/Homebrew/Cask/` | Homebrew cask cache | system | `brew cleanup` |
| `~/Library/Caches/SiriTTS/` | Siri TTS cache | system | re-downloads as needed |
| `~/Library/Caches/com.apple.geod/` | Maps cache | system | re-downloads as needed |
| `~/Library/pnpm/store/` | pnpm store | node | `pnpm store prune` |
| `~/go/pkg/mod/` | Go module cache | go | `go clean -modcache` |
| `~/Library/Caches/Google/Chrome/` | Chrome cache | browser | re-downloads on next visit |
| `~/Library/Caches/com.apple.Safari/` | Safari cache | browser | re-downloads on next visit |
| `~/Library/Caches/org.mozilla.firefox/` | Firefox cache | browser | re-downloads on next visit |
| `~/Library/Caches/com.microsoft.VSCode/` | VS Code cache | editor | regenerates on next launch |
| `~/Library/Caches/com.apple.dt.Xcode/` | Xcode caches | apple | regenerates on next build |
| `~/Library/Caches/com.apple.helpd/` | Help Viewer cache | system | regenerates on demand |
| `~/Library/Caches/com.apple.wallpaper/` | Apple TV aerials | system | re-downloads if wallpaper needs it |
| `~/Library/Logs/` | System & user logs | system | logs regenerate as apps run |
| `~/Library/Developer/Xcode/DerivedData/` | Xcode DerivedData | apple | regenerates on next build |
| `~/.Trash/` | Trash | system | `plong clean` empties trash |
| `~/.cargo/registry/` | Cargo registry cache | rust | `cargo cache --autoclean` |
| `~/.gradle/caches/` | Gradle cache | jvm | `gradle cleanBuildCache` |
| `~/.m2/repository/` | Maven cache | jvm | re-downloads on next build |
| `~/.conda/pkgs/` | Conda packages | python | `conda clean --all` |
| `~/.cache/pip/` | pip cache | python | `pip cache purge` |
| `~/.cache/pypoetry/` | Poetry cache | python | `pip cache purge` |
| `~/.codex/` | Codex CLI cache | ai | regenerates on next run |
| `~/.dartServer/` | Dart analysis server cache | dart | regenerates on next analysis |
| `~/.claude/` | Claude CLI cache | ai | regenerates on next run |

**T2 — Reinstallable toolchains**

| Path | Label | StackTag | CleanHint |
|---|---|---|---|
| `~/.nvm/` | Node.js versions | node | `nvm install <version>` to reinstall |
| `~/.rustup/` | Rust toolchain | rust | `rustup install stable` to reinstall |
| `~/.android/` | Android SDK | mobile | reinstall via Android Studio |
| `~/.cursor/` | Cursor editor data | editor | reinstall Cursor |
| `~/.windsurf/` | Windsurf editor data | editor | reinstall Windsurf |
| `~/Library/Developer/Xcode/Archives/` | Xcode archives | apple | old iOS builds, manual review |
| `~/Library/Developer/Xcode/iOS DeviceSupport/` | iOS device support | apple | regenerates on next device connect |

**T3 — App-level cleanup (detect, don't delete)**

| Path | Label | StackTag | CleanHint |
|---|---|---|---|
| `~/Library/Application Support/Google/Chrome/` | Chrome profile | browser | `chrome://settings/clearBrowserData` |
| `~/Library/Containers/com.docker.docker/Data/` | Docker data | docker | `docker system prune -a` |
| `~/Library/Group Containers/*.Telegram/` | Telegram media | messaging | Telegram → Settings → Data and Storage |
| `~/Library/Application Support/MobileSync/Backup/` | iOS backups | apple | delete in Finder → iPhone → Backups |
| `~/Library/Caches/com.docker.docker/` | Docker build cache | docker | `docker builder prune` |
| `~/Library/Application Support/Slack/` | Slack cache | messaging | Slack → Settings → Clear Cache |
| `~/Library/Application Support/discord/` | Discord cache | messaging | Discord → Settings → Clear Cache |

**T4 — Manual review**

| Path | Label | StackTag | CleanHint |
|---|---|---|---|
| `~/Downloads/` | Downloads folder | system | review DMGs, zips, old files |
| `~/Library/Application Support/Code/` | VS Code workspaces | editor | review before removing |
| `~/.vscode/extensions/` | VS Code extensions | editor | review stale extensions |
| `~/Library/Application Support/com.apple.wallpaper/` | Wallpaper assets | system | review before removing |

### 1c. Discovery engine — unknown dirs

Scans locations the registry doesn't cover:

```
~/*                              → dirs > 100MB, not in safety list
~/Library/Caches/*              → auto-T1 if not in registry
~/Library/Application Support/* → auto-T3 if > 500MB, not in registry
~/Library/Developer/*           → auto-T2 if not in registry
~/Library/Logs/*                → auto-T1 if not in registry
~/Library/Containers/*          → auto-T3 if > 200MB, not in registry
```

Unknown entries get:
- `Label`: relative path (e.g., `~/Library/Caches/com.unknown.app/`)
- `Description`: "unknown cache dir" / "unknown app data"
- `Tier`: heuristic (location-based)
- `Cleanable`: false (always — we don't auto-delete unknowns)
- `CleanHint`: "manual review — not in plong's registry"
- `StackTag`: "" (empty)

### 1d. Stack detection logic

Scan home dir for stack markers, then sort relevant entries first in output:

```go
func detectStacks(home string) []string {
    var stacks []string
    markers := map[string]string{
        ".nvm": "node", "package.json": "node", ".npm": "node",
        "go/pkg": "go", ".go-version": "go",
        ".rustup": "rust", "Cargo.toml": "rust",
        ".android": "mobile",
        "Library/Developer/Xcode": "apple",
        ".cache/pip": "python", ".conda": "python", "requirements.txt": "python",
        ".gradle": "jvm", ".m2": "jvm",
        ".cursor": "editor", ".windsurf": "editor", ".vscode": "editor",
        "Library/Containers/com.docker.docker": "docker",
    }
    // check if path exists → add stack tag
    return stacks
}
```

Entries matching detected stacks get sorted to the top within each tier. Unknown stacks still appear — just at the bottom.

### 1e. Clean TUI update

Current locked items show `🔒`. After this phase, they show the CleanHint:

```
  🔒  Chrome profile — 8.1 GB
       chrome://settings/clearBrowserData
  🔒  iOS backups — 2.4 GB
       delete in Finder → iPhone → Backups
  🔒  Docker data — 12.3 GB
       docker system prune -a
```

## Phase 2 — Medium Wins

| Feature | Details |
|---|---|
| `plong scan --stacks` | Shows detected stacks and their total waste |
| Docker integration | `docker system df` when daemon running, graceful skip when off |
| Stale .dmg detection | `~/Downloads/*.dmg` — show count + total size |
| Large/old files | Find files >100 MB older than 90 days across home |
| `plong hogs` improvement | Per-tier breakdown, not just flat top-20 |
| `plong clean` — empty trash | T1 item: empty `~/.Trash/` with confirmation |

## Phase 3 — Distribution

| Feature | Details |
|---|---|
| Homebrew tap | Create `srmdn/homebrew-tap` repo, write formula, `brew install plong` |
| `plong serve` performance | Start HTTP first, lazy-scan in background, stream results |
| Dashboard upgrade | Show CleanHint for locked items, stack tags, tier breakdown chart |

## Phase 4 — Later

| Feature | Reason to defer |
|---|---|
| Duplicate file finder | Expensive (hash-based), needs careful UX to avoid deleting wrong files |
| App uninstaller | Complex (track `.app` + all Library dirs), legal risk if we delete wrong data |
| Smart scan | Depends on phases 1-3 being solid |
| Telegram media scanner | Need specific macOS path research |
| VS Code extension staleness | Need usage tracking that doesn't exist |

## File Changes Summary

```
internal/scan/
  registry.go      ← NEW: 60+ known targets, each with label/tier/cleanHint/stackTag
  discover.go      ← NEW: auto-discovery for unknown dirs, heuristic classification
  safety.go        ← NEW: never-touch exclusion list + stack detection
  scan.go          ← REWRITE: Collect() merges known + discovered, stack sorting
  format.go        ← NEW: FormatSize moved here + FormatCleanHint helper
internal/clean/
  clean.go         ← UPDATE: show CleanHint for locked items, stack-aware display
internal/serve/
  serve.go         ← UPDATE: include cleanHint/stackTag in JSON API
  templates/index.html ← UPDATE: show CleanHint, stack badges
docs/
  cleanup-guide.md ← UPDATE: add all new targets, cleanHint commands
CLAUDE.md           ← UPDATE: command status table, new architecture notes
```

## Build & Test Protocol

After every code change:
1. `go vet ./...` — zero errors
2. `go build ./...` — compiles clean
3. `go run ./cmd/plong scan` — integration smoke test
4. Zero external deps — Go stdlib only

## Safety Principles (unchanged, reinforced)

- T1/T2 only deletable through `plong clean`
- T3/T4 always locked with explicit CleanHint
- T5 never shown in scan output
- Discovery scanner respects safety exclusion list
- Unknown dirs always `Cleanable: false`
- Home directory boundary check on every deletion
- `"yes, delete"` confirmation gate stays
