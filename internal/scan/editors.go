package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// EditorConfig describes a code editor and its extension storage.
type EditorConfig struct {
	Name         string
	ExtDir       string
	ObsoleteDir  string // dir where old versions are parked (e.g. ".obsolete")
}

var knownEditors = []EditorConfig{
	{"VS Code", ".vscode/extensions/", ".obsolete"},
	{"Cursor", ".cursor/extensions/", ".obsolete"},
	{"Windsurf", ".codeium/windsurf/extensions/", ".obsolete"},
	{"Trae", ".trae/extensions/", ".obsolete"},
	{"Zed", "Library/Application Support/Zed/extensions/installed/", ""},
}

// EditorResult holds the scan result for a single editor.
type EditorResult struct {
	Name          string
	TotalCount    int
	StaleCount    int
	TotalSizeMB   int64
	StaleSizeMB   int64
	Installed     bool
}

// staleThreshold is how old an extension must be to count as stale.
const staleThreshold = 90 * 24 * time.Hour

// ScanEditors detects installed editors, scans their extensions for staleness,
// and returns a slice of results sorted by stale size descending.
func ScanEditors(home string) []EditorResult {
	var results []EditorResult

	for _, ed := range knownEditors {
		fullDir := filepath.Join(home, ed.ExtDir)
		info, err := os.Stat(fullDir)
		if err != nil || !info.IsDir() {
			continue
		}

		r := EditorResult{Name: ed.Name, Installed: true}

		entries, err := os.ReadDir(fullDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == ".obsolete" || name == ".DS_Store" {
				continue
			}

			subPath := filepath.Join(fullDir, name)
			size := dirSizeMB(subPath)
			if size == 0 {
				continue
			}

			r.TotalCount++
			r.TotalSizeMB += size

			if isStale(subPath, ed.ObsoleteDir) {
				r.StaleCount++
				r.StaleSizeMB += size
			}
		}

		if r.TotalCount > 0 {
			results = append(results, r)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StaleSizeMB > results[j].StaleSizeMB
	})

	return results
}

func isStale(path, obsoleteDirName string) bool {
	if obsoleteDirName != "" {
		obsoletePath := filepath.Join(filepath.Dir(path), obsoleteDirName)
		if _, err := os.Stat(obsoletePath); err == nil {
			entries, err := os.ReadDir(obsoletePath)
			if err == nil {
				for _, e := range entries {
					if e.Name() == filepath.Base(path) {
						return true
					}
				}
			}
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) > staleThreshold
}

// PrintEditorSection prints the code editor section in scan output.
func PrintEditorSection(results []EditorResult) {
	if len(results) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("  ── Code editors ──")

	var totalStaleCount int
	var totalStaleSize int64
	for _, r := range results {
		fmt.Printf("    %-10s %3d extensions · %2d stale · %s\n",
			r.Name, r.TotalCount, r.StaleCount, FormatSize(r.StaleSizeMB))
		totalStaleCount += r.StaleCount
		totalStaleSize += r.StaleSizeMB
	}

	if totalStaleCount > 0 {
		fmt.Println()
		fmt.Printf("    %d stale extensions total — %s reclaimable (Tier 2)\n",
			totalStaleCount, FormatSize(totalStaleSize))
	}

	if len(results) < 2 {
		fmt.Println()
		fmt.Println("    Missing your editor? github.com/srmdn/plong/issues/new")
	}
}
