package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type fileEntry struct {
	Path    string
	SizeMB  int64
	ModTime time.Time
}

// ScanStaleDMGs finds .dmg files in ~/Downloads and returns the count
// and total size in MB.
func ScanStaleDMGs(home string) (count int, totalMB int64) {
	pattern := filepath.Join(home, "Downloads", "*.dmg")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return 0, 0
	}
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || info.IsDir() {
			continue
		}
		count++
		totalMB += info.Size() / (1024 * 1024)
	}
	return
}

var allowedHidden = map[string]bool{
	".npm":   true,
	".cache": true,
	".cargo": true,
	".conda": true,
	".gradle": true,
	".m2":    true,
}

// ScanLargeOldFiles walks the home directory (skipping safety-excluded and
// most hidden dirs) to find files >100 MB that haven't been modified in
// 90+ days. Returns up to 10 results sorted by size descending.
func ScanLargeOldFiles(home string) []fileEntry {
	cutoff := time.Now().AddDate(0, 0, -90)
	var results []fileEntry
	walkLargeOld(home, home, "", cutoff, &results)
	sort.Slice(results, func(i, j int) bool { return results[i].SizeMB > results[j].SizeMB })
	if len(results) > 10 {
		results = results[:10]
	}
	return results
}

func walkLargeOld(home, dir, rel string, cutoff time.Time, results *[]fileEntry) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		full := filepath.Join(dir, name)
		childRel := rel + name
		if e.IsDir() {
			if e.Type()&os.ModeSymlink != 0 {
				continue
			}
			checkRel := childRel + "/"
			if IsSafetyExcluded(checkRel) {
				continue
			}
			if strings.HasPrefix(name, ".") && !allowedHidden[name] {
				continue
			}
			walkLargeOld(home, full, childRel+"/", cutoff, results)
		} else {
			if e.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(full)
				if err != nil || info.IsDir() {
					continue
				}
				mb := info.Size() / (1024 * 1024)
				if mb > 100 && info.ModTime().Before(cutoff) {
					*results = append(*results, fileEntry{Path: full, SizeMB: mb, ModTime: info.ModTime()})
				}
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			mb := info.Size() / (1024 * 1024)
			if mb > 100 && info.ModTime().Before(cutoff) {
				*results = append(*results, fileEntry{Path: full, SizeMB: mb, ModTime: info.ModTime()})
			}
		}
	}
}
