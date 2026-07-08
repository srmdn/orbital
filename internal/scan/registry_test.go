package scan

import "testing"

func TestFastTargetsExcludeNicheAppState(t *testing.T) {
	fast := targetPathSet(GetFastTargets())

	for _, path := range []string{
		".codex/",
		".claude/",
		".trae/extensions/",
		"Library/Application Support/com.apple.wallpaper/",
		".npm/",
		".bun/",
		"go/pkg/mod/",
		"Library/Caches/Google/Chrome/",
	} {
		if fast[path] {
			t.Fatalf("fast scan should not include %q", path)
		}
	}
}

func TestFastTargetsKeepUniversalHighSignalPaths(t *testing.T) {
	fast := targetPathSet(GetFastTargets())

	for _, path := range []string{
		"Library/Caches/go-build/",
		"Library/pnpm/store/",
		"Library/Caches/Homebrew/",
		".Trash/",
		"Downloads/",
	} {
		if !fast[path] {
			t.Fatalf("fast scan should include %q", path)
		}
	}
}

func targetPathSet(targets []knownTarget) map[string]bool {
	out := make(map[string]bool, len(targets))
	for _, target := range targets {
		out[target.Path] = true
	}
	return out
}
