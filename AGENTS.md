# AGENTS.md — plong

## Identity

- Repo: `srmdn/plong`
- Description: macOS disk cleanup tool with developer awareness
- Location: `~/Developer/projects/github/plong`

## Stack

- CLI: Go 1.26+ (single binary, no runtime deps)
- Dashboard: Go HTML templates via embed (no JS framework)
- Distribution: Homebrew tap

## Project structure

```text
cmd/plong/        # CLI entry point
internal/
  scan/           # Disk scanning engine
  clean/          # Interactive cleanup TUI
  serve/          # HTTP server + embedded dashboard
    templates/    # Go HTML templates (embedded)
docs/             # Reference documentation
```

## Build

```bash
go build ./...
go vet ./...
go run ./cmd/plong
```

## Commands

- `scan`: detect Tiers 1-4
- `size`: disk summary
- `hogs`: top disk consumers
- `git-trap`: accidental `~/.git` check
- `history`: saved snapshot diffs
- `clean`: interactive cleanup
- `serve`: embedded dashboard

## Conventions

- Go stdlib only
- Exported functions get doc comments
- Use `du -sk` for directory sizing on macOS
- Keep output formatting stable
- Tier 1 must remain safe for programmatic deletion

## Safety rules

- Never delete without explicit user confirmation
- Tier 1 = safe to delete programmatically
- Tiers 2-4 = require user review or app-level cleanup
- Never touch: `.ssh`, `.gitconfig`, Keychains, iCloud, `~/Developer`

## Testing

```bash
go test ./...
go build ./...
go run ./cmd/plong scan
```
