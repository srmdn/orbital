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
	entries = append(entries, discoverCaches(home, known)...)
	entries = append(entries, discoverAppSupport(home, known)...)
	entries = append(entries, discoverDeveloper(home, known)...)
	entries = append(entries, discoverLogs(home, known)...)
	entries = append(entries, discoverContainers(home, known)...)

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
			Tier:        TierSafe,
			Cleanable:   false,
			CleanHint:   "manual review — not in orbital's registry",
			StackTag:    "",
		})
	}
	return entries
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
		name := d.Name() + "/"
		relPath := "Library/Containers/" + name
		if known[relPath] {
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
			CleanHint:   "manual review — not in orbital's registry",
			StackTag:    "",
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
		if known[relPath] {
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
			CleanHint:   "manual review — not in orbital's registry",
			StackTag:    "",
		})
	}
	return entries
}
