package clean

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/srmdn/orbital/internal/scan"
)

var stdin = bufio.NewScanner(os.Stdin)

func Run() {
	home := os.Getenv("HOME")

	fmt.Print("\033[2J\033[H")
	fmt.Println("  scanning...")
	fmt.Println()

	entries, t1, t2, t3, t4 := scan.Collect(home)

	if len(entries) == 0 {
		fmt.Println("  ✅ Nothing reclaimable found. Your Mac is clean!")
		return
	}

	selected := make(map[int]bool)

	for {
		renderMenu(entries, selected, home, t1, t2, t3, t4)
		cmd, ok := readInput()
		if !ok {
			fmt.Print("\033[2J\033[H")
			fmt.Println("  cancelled.")
			return
		}

		switch cmd {
		case "q", "quit":
			fmt.Print("\033[2J\033[H")
			fmt.Println("  cancelled.")
			return
		case "a", "all":
			for i, e := range entries {
				if e.Cleanable {
					selected[i] = true
				}
			}
		case "n", "none":
			selected = make(map[int]bool)
		case "d", "done":
			if len(selected) == 0 {
				continue
			}
			confirmed := confirmDelete(entries, selected)
			if confirmed {
				executeDelete(entries, selected)
				return
			}
		default:
			toggleSelection(entries, selected, cmd)
		}
	}
}

func renderMenu(entries []scan.Entry, selected map[int]bool, home string, t1, t2, t3, t4 int64) {
	fmt.Print("\033[2J\033[H")
	fmt.Println("  orbital clean — reclaim your space")
	fmt.Println()

	currentTier := 0
	for i, e := range entries {
		if e.Tier != currentTier {
			currentTier = e.Tier
			if currentTier > 1 {
				fmt.Println()
			}
			fmt.Printf("  ── TIER %d: %s ──\n", currentTier, scan.TierLabel(currentTier))
		}

		numPrefix := fmt.Sprintf("[%2d]", i+1)
		mark := "[ ]"
		if !e.Cleanable {
			numPrefix = " ── "
			mark = "[🔒]"
		} else if selected[i] {
			mark = "[✓]"
		}
		relPath := strings.TrimPrefix(e.Path, home+"/")

		label := e.Label
		if !e.Cleanable {
			switch e.Tier {
			case scan.TierApp:
				label = fmt.Sprintf("%s  ← %s", e.Label, e.Description)
			case scan.TierManual:
				label = fmt.Sprintf("%s  ← %s", e.Label, e.Description)
			}
		}

		fmt.Printf("    %s %s  %-6s  ~/%s\n", numPrefix, mark, scan.FormatSize(e.SizeMB), relPath)
		if !e.Cleanable {
			fmt.Printf("         %s\n", label)
		}
	}

	fmt.Println()

	if found, sizeMB := scan.HasGitTrap(home); found {
		fmt.Printf("  ⚠️  .git trap in home directory (%s) — run 'orbital git-trap'\n", scan.FormatSize(sizeMB))
		fmt.Println()
	}

	selCount := len(selected)
	var selSize int64
	for i := range selected {
		selSize += entries[i].SizeMB
	}

	totalAll := t1 + t2 + t3 + t4
	totalCleanable := t1 + t2

	fmt.Println("  ──────────────────────────────────────────────")
	fmt.Println()
	fmt.Printf("  ▸  %d selected (%s)  /  %s reclaimable (%s cleanable)\n",
		selCount, scan.FormatSize(selSize), scan.FormatSize(totalAll), scan.FormatSize(totalCleanable))
	fmt.Println()
	fmt.Println("  ── controls ──")
	fmt.Println()
	fmt.Println("    Type a number, then Enter — selects or unselects a cleanable item.")
	fmt.Println("    🔒 items require app-level or manual action — see scan for details.")
	fmt.Println("    a  =  select all cleanable  d  =  review & delete selected")
	fmt.Println("    n  =  unselect all           q  =  cancel & quit")
	fmt.Println()
	fmt.Print("  > ")
}

func toggleSelection(entries []scan.Entry, selected map[int]bool, input string) {
	n, err := strconv.Atoi(input)
	if err != nil || n < 1 || n > len(entries) {
		return
	}
	idx := n - 1
	if !entries[idx].Cleanable {
		return // non-cleanable items cannot be toggled
	}
	if selected[idx] {
		delete(selected, idx)
	} else {
		selected[idx] = true
	}
}

func confirmDelete(entries []scan.Entry, selected map[int]bool) bool {
	fmt.Print("\033[2J\033[H")
	fmt.Println("  confirm deletion")
	fmt.Println()
	fmt.Println("  These items will be permanently deleted:")
	fmt.Println()

	var total int64
	for i := range selected {
		e := entries[i]
		fmt.Printf("    %-6s  %s\n", scan.FormatSize(e.SizeMB), e.Label)
		total += e.SizeMB
	}

	fmt.Println("  ────────────────────────")
	fmt.Printf("    %s total\n", scan.FormatSize(total))
	fmt.Println()
	fmt.Println("  ⚠️  This action cannot be undone.")
	fmt.Println()

	fmt.Print("  Type 'yes, delete' to confirm: ")

	input, _ := readInput()
	return input == "yes, delete"
}

func executeDelete(entries []scan.Entry, selected map[int]bool) {
	fmt.Print("\033[2J\033[H")
	fmt.Println("  deleting...")
	fmt.Println()

	home := os.Getenv("HOME")
	var freed int64
	var failed int

	for i := range selected {
		e := entries[i]

		// Safety: only delete paths under home directory
		rel, err := filepath.Rel(home, e.Path)
		if err != nil || strings.HasPrefix(rel, "..") {
			fmt.Printf("    ⚠️  skipping %s — outside home directory\n", e.Label)
			failed++
			continue
		}

		fmt.Printf("    deleting %s... ", e.Label)
		if err := os.RemoveAll(e.Path); err != nil {
			fmt.Printf("failed (%v)\n", err)
			failed++
		} else {
			fmt.Println("done")
			freed += e.SizeMB
		}
	}

	fmt.Println()
	fmt.Printf("  ✅ Freed %s\n", scan.FormatSize(freed))
	if failed > 0 {
		fmt.Printf("  (%d items could not be deleted)\n", failed)
	}
	fmt.Println()
	fmt.Println("  Run 'orbital scan' to verify.")
}

func readInput() (string, bool) {
	if stdin.Scan() {
		return strings.TrimSpace(strings.ToLower(stdin.Text())), true
	}
	return "", false
}
