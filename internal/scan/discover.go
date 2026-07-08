package scan

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscoverUnknown scans directories not covered by the registry and returns
// entries with heuristic tier assignments. It skips anything in the known map
// and anything on the safety exclusion list.
func DiscoverUnknown(home string, known map[string]bool) []Entry {
	var entries []Entry

	entries = append(entries, discoverHome(home, known)...)
	entries = append(entries, discoverDotCache(home, known)...)
	entries = append(entries, discoverCaches(home, known)...)
	entries = append(entries, discoverAppSupport(home, known)...)
	entries = append(entries, discoverDeveloper(home, known)...)
	entries = append(entries, discoverLogs(home, known)...)
	entries = append(entries, discoverContainers(home, known)...)
	entries = append(entries, discoverGroupContainers(home)...)

	return entries
}

func discoverHome(home string, known map[string]bool) []Entry {
	var entries []Entry
	dirs, err := os.ReadDir(home)
	if err != nil {
		return entries
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if known[name+"/"] {
			continue
		}
		if IsSafetyExcluded(name + "/") {
			continue
		}
		full := filepath.Join(home, name)
		mb := dirSizeMB(full)
		if mb < 100 {
			continue
		}
		entries = append(entries, Entry{
			Path:        full,
			Label:       "~/" + name,
			Description: "large directory, manual review needed",
			SizeMB:      mb,
			Tier:        TierManual,
			Cleanable:   false,
			CleanHint:   "manual review — not in plong's registry",
			StackTag:    "",
		})
	}
	return entries
}

func discoverDotCache(home string, known map[string]bool) []Entry {
	return discoverSubdir(home, ".cache", known, TierSafe, true, "cache (auto-regenerates)")
}

func discoverCaches(home string, known map[string]bool) []Entry {
	return discoverSubdir(home, "Library/Caches", known, TierSafe, true, "cache (auto-regenerates)")
}

func discoverAppSupport(home string, known map[string]bool) []Entry {
	return discoverSubdir(home, "Library/Application Support", known, TierApp, false, "app data (review before deleting)")
}

func discoverDeveloper(home string, known map[string]bool) []Entry {
	return discoverSubdir(home, "Library/Developer", known, TierReinst, true, "developer tool data")
}

func discoverLogs(home string, known map[string]bool) []Entry {
	return discoverSubdir(home, "Library/Logs", known, TierSafe, true, "log files (auto-regenerate)")
}

func discoverContainers(home string, known map[string]bool) []Entry {
	var entries []Entry
	parent := filepath.Join(home, "Library/Containers")
	dirs, err := os.ReadDir(parent)
	if err != nil {
		return entries
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		if d.Name() == "com.docker.docker" {
			continue
		}
		name := d.Name() + "/"
		relPath := "Library/Containers/" + name
		if known[relPath] || hasKnownDescendant(home, known, relPath) {
			continue
		}
		full := filepath.Join(parent, d.Name())
		mb := dirSizeMB(full)
		if mb < 200 {
			continue
		}
		entries = append(entries, Entry{
			Path:        full,
			Label:       "~/Library/Containers/" + d.Name(),
			Description: "unknown container data",
			SizeMB:      mb,
			Tier:        TierApp,
			Cleanable:   false,
			CleanHint:   "manual review — not in plong's registry",
			StackTag:    "",
		})
	}
	return entries
}

func discoverGroupContainers(home string) []Entry {
	var entries []Entry
	parent := filepath.Join(home, "Library/Group Containers")
	dirs, err := os.ReadDir(parent)
	if err != nil {
		return entries
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		if !strings.Contains(strings.ToLower(d.Name()), "telegram") {
			continue
		}
		full := filepath.Join(parent, d.Name())
		mb := dirSizeMB(full)
		if mb < 200 {
			continue
		}
		entries = append(entries, Entry{
			Path:        full,
			Label:       "Telegram media",
			Description: "Cached media and files",
			SizeMB:      mb,
			Tier:        TierApp,
			Cleanable:   false,
			CleanHint:   "Telegram → Settings → Data and Storage",
			StackTag:    "messaging",
		})
	}
	return entries
}

func discoverSubdir(home, relParent string, known map[string]bool, tier int, cleanable bool, desc string) []Entry {
	var entries []Entry
	parent := filepath.Join(home, relParent)
	dirs, err := os.ReadDir(parent)
	if err != nil {
		return entries
	}
	sizeThreshold := int64(0)
	if tier == TierApp {
		sizeThreshold = 500
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		name := d.Name() + "/"
		relPath := relParent + "/" + name
		if known[relPath] || hasKnownDescendant(home, known, relPath) {
			continue
		}
		full := filepath.Join(parent, d.Name())
		mb := dirSizeMB(full)
		if mb == 0 || mb < sizeThreshold {
			continue
		}
		entries = append(entries, Entry{
			Path:        full,
			Label:       "~/" + relParent + "/" + d.Name(),
			Description: desc,
			SizeMB:      mb,
			Tier:        tier,
			Cleanable:   cleanable,
			CleanHint:   "manual review — not in plong's registry",
			StackTag:    "",
		})
	}
	return entries
}

func hasKnownDescendant(home string, known map[string]bool, relPath string) bool {
	prefix := relPath
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for candidate := range known {
		if strings.HasPrefix(candidate, prefix) {
			if _, err := os.Stat(filepath.Join(home, candidate)); err == nil {
				return true
			}
			if _, err := os.Stat(filepath.Join(home, strings.TrimSuffix(candidate, "/"))); err == nil &&
				strings.HasPrefix(candidate, prefix) {
				return true
			}
		}
	}
	return false
}
