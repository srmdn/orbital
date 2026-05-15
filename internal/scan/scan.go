package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	TierSafe   = 1
	TierReinst = 2
	TierApp    = 3
	TierManual = 4
	TierNever  = 5
)

type Entry struct {
	Path        string
	Label       string
	Description string
	SizeMB      int64
	Tier        int
	Cleanable   bool
	CleanHint   string
	StackTag    string
}

// Collect scans the home directory and returns all reclaimable entries
// sorted by tier, stack relevance, and size.
func Collect(home string) (entries []Entry, t1, t2, t3, t4 int64) {
	targets := GetKnownTargets()

	known := map[string]bool{}
	for _, t := range targets {
		known[t.Path] = true
	}

	for _, t := range targets {
		fullPath := filepath.Join(home, t.Path)
		mb := dirSizeMB(fullPath)
		if mb == 0 {
			continue
		}
		entries = append(entries, Entry{
			Path:        fullPath,
			Label:       t.Label,
			Description: t.Description,
			SizeMB:      mb,
			Tier:        t.Tier,
			Cleanable:   t.Cleanable,
			CleanHint:   t.CleanHint,
			StackTag:    t.StackTag,
		})
	}

	discovered := DiscoverUnknown(home, known)
	entries = append(entries, discovered...)

	dockerEntries := DiscoverDocker(home)
	entries = append(entries, dockerEntries...)

	stacks := DetectStacks(home)
	stackSet := map[string]bool{}
	for _, s := range stacks {
		stackSet[s] = true
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Tier != entries[j].Tier {
			return entries[i].Tier < entries[j].Tier
		}
		iStack := entries[i].StackTag != "" && stackSet[entries[i].StackTag]
		jStack := entries[j].StackTag != "" && stackSet[entries[j].StackTag]
		if iStack != jStack {
			return iStack
		}
		return entries[i].SizeMB > entries[j].SizeMB
	})

	for _, e := range entries {
		switch e.Tier {
		case TierSafe:
			t1 += e.SizeMB
		case TierReinst:
			t2 += e.SizeMB
		case TierApp:
			t3 += e.SizeMB
		case TierManual:
			t4 += e.SizeMB
		}
	}

	return
}

// HasGitTrap checks for an accidental .git repo in the home directory.
func HasGitTrap(home string) (bool, int64) {
	gitPath := filepath.Join(home, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || !info.IsDir() {
		return false, 0
	}
	return true, dirSizeMB(gitPath)
}

// TierLabel returns a human-readable label for a tier number.
func TierLabel(tier int) string {
	switch tier {
	case TierSafe:
		return "Safe caches (auto-regenerate)"
	case TierReinst:
		return "Reinstallable toolchains"
	case TierApp:
		return "App-level cleanup required"
	case TierManual:
		return "Manual review required"
	case TierNever:
		return "Never touch"
	default:
		return fmt.Sprintf("Tier %d", tier)
	}
}

func Run() {
	fmt.Println()
	fmt.Println("  scanning your mac...")
	fmt.Println()
	home := os.Getenv("HOME")

	entries, t1, t2, t3, t4 := Collect(home)

	var t1Entries, t2Entries, t3Entries, t4Entries []Entry
	for _, e := range entries {
		switch e.Tier {
		case TierSafe:
			t1Entries = append(t1Entries, e)
		case TierReinst:
			t2Entries = append(t2Entries, e)
		case TierApp:
			t3Entries = append(t3Entries, e)
		case TierManual:
			t4Entries = append(t4Entries, e)
		}
	}

	printTier(TierSafe, t1Entries)
	fmt.Println()
	printTier(TierReinst, t2Entries)
	fmt.Println()
	printTier(TierApp, t3Entries)
	fmt.Println()
	printTier(TierManual, t4Entries)
	fmt.Println()

	if count, totalMB := ScanStaleDMGs(home); count > 0 {
		fmt.Println("  ── Stale disk images ──")
		fmt.Printf("  %d DMGs in ~/Downloads — %s total\n", count, FormatSize(totalMB))
		fmt.Println("  Run rm ~/Downloads/*.dmg to remove (review first)")
		fmt.Println()
	}

	if files := ScanLargeOldFiles(home); len(files) > 0 {
		fmt.Println("  ── Large old files (>100 MB, 90+ days) ──")
		for _, f := range files {
			relPath := strings.TrimPrefix(f.Path, home)
			relPath = strings.TrimPrefix(relPath, "/")
			fmt.Printf("    %-6s  ~/%s  (modified %s)\n",
				FormatSize(f.SizeMB), relPath, f.ModTime.Format("2006-01-02"))
		}
		fmt.Println()
	}

	checkGitTrapInternal(home)

	total := t1 + t2 + t3 + t4
	if total > 0 {
		fmt.Println("  ────────────────────────────────────────────────")
		fmt.Printf("  Total reclaimable: %s\n", FormatSize(total))
		fmt.Println()
		fmt.Println("  Run 'orbital clean' for interactive cleanup (Tiers 1-2 only)")
		fmt.Println("  Tiers 3-4 require app-level or manual action — see docs")
		fmt.Println("  Reference: docs/cleanup-guide.md")
	}
}

func printTier(tier int, entries []Entry) {
	fmt.Printf("  ── TIER %d: %s ──\n", tier, TierLabel(tier))
	if len(entries) == 0 {
		fmt.Println("    (nothing found)")
		return
	}
	for _, e := range entries {
		stackTag := ""
		if e.StackTag != "" {
			stackTag = " [" + e.StackTag + "]"
		}
		fmt.Printf("    %-6s  %-25s  %s%s\n", FormatSize(e.SizeMB), e.Label, e.Description, stackTag)
	}
}

func checkGitTrapInternal(home string) {
	gitPath := filepath.Join(home, ".git")
	if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
		mb := dirSizeMB(gitPath)
		fmt.Println("  ⚠️  .GIT TRAP DETECTED")
		fmt.Printf("    %s — accidental git repo in home directory\n", FormatSize(mb))
		fmt.Println()
	}
}

// DiskSize prints a quick disk space summary.
func DiskSize() {
	cmd := exec.Command("df", "-h", "/")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Could not read disk info")
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 5 {
			fmt.Printf("  💾 %s available  (%s used)\n", fields[3], fields[4])
		}
	}

	cmd = exec.Command("df", "-h")
	out, _ = cmd.Output()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "/dev/disk3") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				fmt.Printf("  Container: %s available\n", fields[3])
			}
		}
	}
}

// TopHogs prints the largest directories in the home folder grouped by tier.
func TopHogs() {
	home := os.Getenv("HOME")
	targets := GetKnownTargets()

	registryPaths := map[string]bool{}
	tierEntries := make(map[int][]struct {
		path string
		size int64
	})

	for _, t := range targets {
		registryPaths[t.Path] = true
		fullPath := filepath.Join(home, t.Path)
		mb := dirSizeMB(fullPath)
		if mb == 0 {
			continue
		}
		tierEntries[t.Tier] = append(tierEntries[t.Tier], struct {
			path string
			size int64
		}{t.Path, mb})
	}

	for tier := range tierEntries {
		sort.Slice(tierEntries[tier], func(i, j int) bool {
			return tierEntries[tier][i].size > tierEntries[tier][j].size
		})
	}

	var unregistered []struct {
		path string
		size int64
	}
	homeDirs, err := os.ReadDir(home)
	if err == nil {
		for _, d := range homeDirs {
			name := d.Name()
			if registryPaths[name+"/"] {
				continue
			}
			if IsSafetyExcluded(name + "/") {
				continue
			}
			fullPath := filepath.Join(home, name)
			mb := dirSizeMB(fullPath)
			if mb > 0 {
				unregistered = append(unregistered, struct {
					path string
					size int64
				}{name, mb})
			}
		}
	}
	sort.Slice(unregistered, func(i, j int) bool {
		return unregistered[i].size > unregistered[j].size
	})

	fmt.Println("  Top space consumers in ~")
	fmt.Println()

	for tier := TierSafe; tier <= TierManual; tier++ {
		entries := tierEntries[tier]
		if len(entries) == 0 {
			continue
		}
		fmt.Printf("  ── TIER %d: %s ──\n", tier, TierLabel(tier))
		for _, e := range entries {
			fmt.Printf("    %-6s  ~/%s\n", FormatSize(e.size), e.path)
		}
		fmt.Println()
	}

	if len(unregistered) > 0 {
		fmt.Println("  ── Unregistered ──")
		for _, u := range unregistered {
			fmt.Printf("    %-6s  ~/%s/\n", FormatSize(u.size), u.path)
		}
		fmt.Println()
	}
}

// ScanStacks groups collected entries by detected stack tags and prints
// per-stack space totals.
func ScanStacks(home string) {
	detected := DetectStacks(home)
	if len(detected) == 0 {
		fmt.Println("  No dev stacks detected.")
		return
	}

	stackSet := map[string]bool{}
	for _, s := range detected {
		stackSet[s] = true
	}

	entries, _, _, _, _ := Collect(home)

	type stackGroup struct {
		tag    string
		sizeMB int64
		count  int
	}
	groups := map[string]*stackGroup{}

	for _, e := range entries {
		if e.StackTag == "" || !stackSet[e.StackTag] {
			continue
		}
		if _, ok := groups[e.StackTag]; !ok {
			groups[e.StackTag] = &stackGroup{tag: e.StackTag}
		}
		groups[e.StackTag].sizeMB += e.SizeMB
		groups[e.StackTag].count++
	}

	fmt.Println()
	fmt.Println("  Detected stacks:")
	fmt.Println()
	for _, tag := range detected {
		g, ok := groups[tag]
		if !ok {
			fmt.Printf("    %-10s (none found)\n", tag)
			continue
		}
		fmt.Printf("    %-10s %-8s (%d entries)\n", g.tag, FormatSize(g.sizeMB), g.count)
	}
	fmt.Println()
}

// CheckGitTrap checks for accidental .git in home directory.
func CheckGitTrap() {
	home := os.Getenv("HOME")
	found, sizeMB := HasGitTrap(home)

	if !found {
		fmt.Println("  ✅ No .git trap. You're safe.")
		return
	}

	fmt.Printf("  ⚠️  .git found in home directory — %s\n", FormatSize(sizeMB))
	fmt.Println()
	fmt.Println("  Your home directory is accidentally a git repository.")
	fmt.Println("  This silently tracks everything and can grow to 50+ GB.")
	fmt.Println()
	fmt.Println("  To fix: rm -rf ~/.git")
	fmt.Println("  Add to ~/.zshrc to prevent recurrence (see docs).")
}

// dirSizeMB returns the size of a directory in MB.
func dirSizeMB(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size() / (1024 * 1024)
	}

	cmd := exec.Command("du", "-sk", path)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0
	}
	var kb int64
	fmt.Sscanf(fields[0], "%d", &kb)
	return kb / 1024
}
