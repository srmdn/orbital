# CLAUDE.md — orbital

## Identity

- Repo: srmdn/orbital
- Description: macOS disk cleanup tool with developer awareness
- Location: ~/Developer/projects/github/orbital

## Stack

- **CLI**: Go 1.26+ (single binary, no runtime deps)
- **Dashboard**: Astro (planned, served by Go binary via embed)
- **Distribution**: Homebrew tap

## Project Structure

```
cmd/orbital/      # CLI entry point
internal/
  scan/           # Disk scanning engine
  clean/          # Interactive cleanup ✅
  report/         # Report generation (planned)
dashboard/        # Astro web dashboard (planned)
docs/             # Reference documentation
```

## Build

```bash
go build ./...        # Check compilation
go vet ./...          # Lint
go run ./cmd/orbital  # Run
```

## Commands

| Command | Status |
|---|---|
| `scan` | ✅ Working |
| `size` | ✅ Working |
| `hogs` | ✅ Working |
| `git-trap` | ✅ Working |
| `clean` | ✅ Working |
| `serve` (dashboard) | 🚧 Planned |

## Conventions

- Go stdlib only (zero external deps for CLI)
- Exported functions get doc comments (Go std)
- Internal helpers are uncommented (self-documenting names)
- Use `du -sk` for dir sizes (macOS compatible)
- No `find`, `grep`, `cat`, `sed` in Go code — use stdlib
- Output format: 2-space indent, emoji in status lines only

## Safety Rules

- Never delete without explicit user confirmation
- Tier 1 = safe to delete programmatically
- Tier 2-4 = require user interaction
- Never touch: .ssh, .gitconfig, Keychains, iCloud, ~/Developer

## Testing

```bash
go test ./...          # Unit tests
go run ./cmd/orbital scan   # Integration smoke test
```
