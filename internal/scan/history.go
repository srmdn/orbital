package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const maxSnapshots = 30

type snapshotEntry struct {
	Path      string `json:"path"`
	Label     string `json:"label"`
	SizeMB    int64  `json:"size_mb"`
	Tier      int    `json:"tier"`
	Cleanable bool   `json:"cleanable"`
	StackTag  string `json:"stack_tag,omitempty"`
}

// Snapshot is a point-in-time record of a scan result, serialized to JSON.
type Snapshot struct {
	Timestamp time.Time       `json:"timestamp"`
	TotalMB   int64           `json:"total_mb"`
	T1MB      int64           `json:"t1_mb"`
	T2MB      int64           `json:"t2_mb"`
	T3MB      int64           `json:"t3_mb"`
	T4MB      int64           `json:"t4_mb"`
	Entries   []snapshotEntry `json:"entries"`
}

type deltaItem struct {
	Label    string
	ChangeMB int64
	IsNew    bool
	IsGone   bool
}

type scanDelta struct {
	Since    time.Time
	ChangeMB int64
	Items    []deltaItem
}

func historyDir(home string) string {
	return filepath.Join(home, ".plong", "history")
}

// SaveSnapshot writes the current scan result to ~/.plong/history/<timestamp>.json
// and prunes old snapshots keeping at most maxSnapshots.
func SaveSnapshot(home string, entries []Entry, t1, t2, t3, t4 int64) {
	dir := historyDir(home)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	snap := Snapshot{
		Timestamp: time.Now(),
		TotalMB:   t1 + t2 + t3 + t4,
		T1MB:      t1,
		T2MB:      t2,
		T3MB:      t3,
		T4MB:      t4,
	}
	for _, e := range entries {
		snap.Entries = append(snap.Entries, snapshotEntry{
			Path:      e.Path,
			Label:     e.Label,
			SizeMB:    e.SizeMB,
			Tier:      e.Tier,
			Cleanable: e.Cleanable,
			StackTag:  e.StackTag,
		})
	}

	filename := time.Now().Format("2006-01-02T15-04-05") + ".json"
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(filepath.Join(dir, filename), data, 0644)

	pruneHistory(home)
}

// LoadLatest returns the most recent scan snapshot, or an error if no history exists.
func LoadLatest(home string) (*Snapshot, error) {
	dir := historyDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no history")
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var snap Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			continue
		}
		return &snap, nil
	}
	return nil, fmt.Errorf("no valid snapshots")
}

// ComputeDelta compares the current scan entries against a previous snapshot
// and returns the net change plus the top items that changed.
func ComputeDelta(prev *Snapshot, entries []Entry) *scanDelta {
	prevMap := map[string]int64{}
	prevLabel := map[string]string{}
	for _, e := range prev.Entries {
		prevMap[e.Path] = e.SizeMB
		prevLabel[e.Path] = e.Label
	}

	currMap := map[string]int64{}
	currLabel := map[string]string{}
	for _, e := range entries {
		currMap[e.Path] = e.SizeMB
		currLabel[e.Path] = e.Label
	}

	var items []deltaItem
	var totalChange int64

	for path, currSize := range currMap {
		prevSize, existed := prevMap[path]
		if !existed {
			items = append(items, deltaItem{
				Label:    currLabel[path],
				ChangeMB: currSize,
				IsNew:    true,
			})
			totalChange += currSize
		} else if currSize != prevSize {
			diff := currSize - prevSize
			items = append(items, deltaItem{
				Label:    currLabel[path],
				ChangeMB: diff,
			})
			totalChange += diff
		}
	}

	for path := range prevMap {
		if _, exists := currMap[path]; !exists {
			items = append(items, deltaItem{
				Label:    prevLabel[path],
				ChangeMB: -prevMap[path],
				IsGone:   true,
			})
			totalChange -= prevMap[path]
		}
	}

	sort.Slice(items, func(i, j int) bool {
		ai := items[i].ChangeMB
		if ai < 0 {
			ai = -ai
		}
		aj := items[j].ChangeMB
		if aj < 0 {
			aj = -aj
		}
		return ai > aj
	})

	return &scanDelta{
		Since:    prev.Timestamp,
		ChangeMB: totalChange,
		Items:    items,
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		w := int(d.Hours() / (24 * 7))
		if w == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", w)
	default:
		return t.Format("Jan 2")
	}
}

func printDelta(d *scanDelta) {
	if len(d.Items) == 0 {
		return
	}

	fmt.Println()
	sign := "+"
	if d.ChangeMB < 0 {
		sign = ""
	}
	fmt.Printf("  📊 Since %s: %s%s\n", timeAgo(d.Since), sign, FormatSize(d.ChangeMB))

	maxItems := 5
	if len(d.Items) < maxItems {
		maxItems = len(d.Items)
	}
	for _, item := range d.Items[:maxItems] {
		var changeStr string
		if item.ChangeMB > 0 {
			changeStr = "+" + FormatSize(item.ChangeMB)
		} else {
			changeStr = FormatSize(item.ChangeMB)
		}

		tag := ""
		if item.IsNew {
			tag = " (new)"
		} else if item.IsGone {
			tag = " (gone)"
		}

		fmt.Printf("    %-8s  %s%s\n", changeStr, item.Label, tag)
	}
	fmt.Println()
}

func pruneHistory(home string) {
	dir := historyDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= maxSnapshots {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries[:len(entries)-maxSnapshots] {
		os.Remove(filepath.Join(dir, e.Name()))
	}
}
