package scan

import (
	"os"
	"path/filepath"
)

var safetyExclusions = []string{
	".ssh/",
	".gitconfig",
	"Library/Keychains/",
	"Library/Mobile Documents/",
	"Developer/",
	"Documents/",
	"Desktop/",
	"Pictures/",
	"Music/",
	"Movies/",
}

// IsSafetyExcluded reports whether a path (relative to $HOME) is on the
// never-touch exclusion list.
func IsSafetyExcluded(path string) bool {
	for _, excl := range safetyExclusions {
		if path == excl || path == excl[:len(excl)-1] {
			return true
		}
		if len(path) > len(excl) && path[:len(excl)] == excl {
			return true
		}
	}
	return false
}

// GetSafetyExclusions returns the list of relative paths plong never scans.
func GetSafetyExclusions() []string {
	return safetyExclusions
}

// DetectStacks scans the home directory for dev stack markers and returns
// a deduplicated list of stack tags.
func DetectStacks(home string) []string {
	markers := []struct {
		path string
		tag  string
	}{
		{".nvm", "node"},
		{"package.json", "node"},
		{".npm", "node"},
		{"go/pkg", "go"},
		{".go-version", "go"},
		{".rustup", "rust"},
		{"Cargo.toml", "rust"},
		{".android", "mobile"},
		{"Library/Developer/Xcode", "apple"},
		{".cache/pip", "python"},
		{".conda", "python"},
		{"requirements.txt", "python"},
		{".gradle", "jvm"},
		{".m2", "jvm"},
		{".cursor", "editor"},
		{".windsurf", "editor"},
		{".vscode", "editor"},
		{"Library/Containers/com.docker.docker", "docker"},
	}

	seen := map[string]bool{}
	var stacks []string
	for _, m := range markers {
		if seen[m.tag] {
			continue
		}
		full := filepath.Join(home, m.path)
		if _, err := os.Stat(full); err == nil {
			seen[m.tag] = true
			stacks = append(stacks, m.tag)
		}
	}
	return stacks
}
