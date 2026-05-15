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
}

type tierConfig struct {
	Path        string
	Label       string
	Description string
}

var tier1Targets = []tierConfig{
	{".npm", "npm cache", "Node package manager cache"},
	{".bun", "bun cache", "Bun package manager cache"},
	{".cache", "System caches", "pip, HuggingFace, SDK caches"},
	{".codex", "Codex CLI cache", "OpenAI Codex CLI cache"},
	{"Library/Caches/go-build", "Go build cache", "Go compiler build artifacts"},
	{"Library/Caches/goimports", "Go imports cache", "Go imports auto-complete cache"},
	{"Library/Caches/node-gyp", "node-gyp cache", "Native module build cache"},
	{"Library/Caches/Homebrew", "Homebrew cache", "Brew download cache"},
	{"Library/Caches/SiriTTS", "Siri TTS cache", "Text-to-speech voice data"},
	{"Library/Caches/com.apple.geod", "Maps cache", "Geolocation/maps cache"},
	{"Library/pnpm/store", "pnpm store", "pnpm package cache"},
	{"go/pkg/mod", "Go module cache", "Downloaded Go modules"},
}

var tier2Targets = []tierConfig{
	{".nvm", "nvm / Node.js", "Node version manager installs"},
	{".rustup", "Rust toolchain", "Rustup toolchain installs"},
	{".android", "Android SDK", "Android SDK and emulators"},
	{".cursor", "Cursor editor", "Cursor IDE data"},
	{".windsurf", "Windsurf editor", "Windsurf IDE data"},
}

var tier3Targets = []tierConfig{
	{"Library/Application Support/Google/Chrome", "Chrome profile", "Clear from Chrome settings — contains bookmarks & passwords"},
	{"Library/Containers/com.docker.docker/Data", "Docker data", "Use 'docker system prune -a' instead"},
}

var tier4Targets = []tierConfig{
	{"Downloads", "Downloads folder", "Review DMGs, zips, old files manually"},
	{"Library/Application Support/Code", "VS Code workspaces", "Old workspaces — review before removing"},
}

// Collect scans the home directory and returns all reclaimable entries.
func Collect(home string) (entries []Entry, t1, t2, t3, t4 int64) {
	t1 = scanTier(home, tier1Targets, "", &entries, TierSafe, true, true)
	t2 = scanTier(home, tier2Targets, "", &entries, TierReinst, true, true)
	t3 = scanTier(home, tier3Targets, "", &entries, TierApp, false, false)
	t4 = scanTier(home, tier4Targets, "", &entries, TierManual, false, false)
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

// FormatSize formats a size in MB to a human-readable string.
func FormatSize(mb int64) string {
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", float64(mb)/1024.0)
	}
	if mb == 0 {
		return "0 MB"
	}
	return fmt.Sprintf("%d MB", mb)
}

func Run() {
	fmt.Println()
	fmt.Println("  scanning your mac...")
	fmt.Println()
	home := os.Getenv("HOME")

	var entries []Entry

	tier1Total := scanTier(home, tier1Targets, "TIER 1: Safe caches (auto-regenerate)", &entries, TierSafe, true, true)
	fmt.Println()
	tier2Total := scanTier(home, tier2Targets, "TIER 2: Reinstallable toolchains", &entries, TierReinst, true, true)
	fmt.Println()
	tier3Total := scanTier(home, tier3Targets, "TIER 3: App-level cleanup required", &entries, TierApp, false, true)
	fmt.Println()
	tier4Total := scanTier(home, tier4Targets, "TIER 4: Manual review required", &entries, TierManual, false, true)
	fmt.Println()

	checkGitTrapInternal(home)

	total := tier1Total + tier2Total + tier3Total + tier4Total
	if total > 0 {
		fmt.Println("  ────────────────────────────────────────────────")
		fmt.Printf("  Total reclaimable: %s\n", FormatSize(total))
		fmt.Println()
		fmt.Println("  Run 'orbital clean' for interactive cleanup (Tiers 1-2 only)")
		fmt.Println("  Tiers 3-4 require app-level or manual action — see docs")
		fmt.Println("  Reference: docs/cleanup-guide.md")
	}
}

func scanTier(home string, targets []tierConfig, header string, entries *[]Entry, tier int, cleanable, verbose bool) int64 {
	if verbose {
		fmt.Printf("  ── %s ──\n", header)
	}
	var total int64
	for _, t := range targets {
		fullPath := filepath.Join(home, t.Path)
		mb := dirSizeMB(fullPath)
		if mb > 0 {
			*entries = append(*entries, Entry{
				Path:        fullPath,
				Label:       t.Label,
				Description: t.Description,
				SizeMB:      mb,
				Tier:        tier,
				Cleanable:   cleanable,
			})
			total += mb
			if verbose {
				fmt.Printf("    %-6s  %-25s  %s\n", FormatSize(mb), t.Label, t.Description)
			}
		}
	}
	if verbose && total == 0 {
		fmt.Println("    (nothing found)")
	}
	return total
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

// TopHogs prints the largest directories in the home folder.
func TopHogs() {
	home := os.Getenv("HOME")
	entries, _ := os.ReadDir(home)

	type hogEntry struct {
		path string
		size int64
	}
	var hogs []hogEntry

	for _, entry := range entries {
		fullPath := filepath.Join(home, entry.Name())
		mb := dirSizeMB(fullPath)
		if mb > 0 {
			hogs = append(hogs, hogEntry{path: entry.Name(), size: mb})
		}
	}

	sort.Slice(hogs, func(i, j int) bool { return hogs[i].size > hogs[j].size })

	limit := 20
	if len(hogs) < limit {
		limit = len(hogs)
	}

	fmt.Println("  Top space consumers in ~")
	fmt.Println()
	for i := 0; i < limit; i++ {
		fmt.Printf("    %-6s  ~/%s\n", FormatSize(hogs[i].size), hogs[i].path)
	}
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
